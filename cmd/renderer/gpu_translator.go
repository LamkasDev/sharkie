package renderer

import "C"
import (
	"fmt"
	"sync"

	as "github.com/LamkasDev/asche"
	"github.com/LamkasDev/cimgui-go-vulkan/backend"
	glfwvulkanbackend "github.com/LamkasDev/cimgui-go-vulkan/backend/glfwvulkan-backend"
	"github.com/LamkasDev/cimgui-go-vulkan/imgui"
	"github.com/LamkasDev/sharkie/cmd/structs/gcn"
	. "github.com/LamkasDev/sharkie/cmd/structs/gpu"
	. "github.com/LamkasDev/sharkie/cmd/structs/spirv"
	vk "github.com/goki/vulkan"
)

type GpuTranslatorPipelineKey struct {
	PixelShaderAddress uintptr
	SurfaceAddress     uintptr
}

// GpuTranslator converts decoded DrawCalls into Vulkan commands.
type GpuTranslator struct {
	handles VulkanHandles
	backend backend.Backend[glfwvulkanbackend.GLFWWindowFlags]

	// Vulkan surfaces mirroring guest framebuffers.
	surfacesMutex sync.Mutex
	surfaces      map[uintptr]*GpuSurface

	// Stub pipeline shared across all draws until real shaders are available.
	stubPipelineLayout vk.PipelineLayout
	stubVertShader     vk.ShaderModule
	stubFragShader     vk.ShaderModule

	// Recompiled SPIR-V shaders mirroring Liverpool.LoadedShaders.
	shadersMutex sync.Mutex
	shaders      map[uintptr]*SpirvShader

	// VkShaderModules created from SPIR-V shaders.
	shaderModulesMutex sync.Mutex
	shaderModules      map[uintptr]vk.ShaderModule

	// Per-draw compiled pipelines.
	pipelinesMutex sync.Mutex
	pipelines      map[GpuTranslatorPipelineKey]vk.Pipeline

	// Physical buffers for Constant RAM snapshots.
	constRamBuffersMutex sync.Mutex
	constRamBuffers      map[uint32]vk.Buffer
	constRamBufferMems   map[uint32]vk.DeviceMemory

	// Physical buffers for User Data snapshots.
	userDataBuffersMutex sync.Mutex
	userDataBuffers      map[uint32]vk.Buffer
	userDataBufferMems   map[uint32]vk.DeviceMemory

	// Command pool/buffer for this frame's GPU work.
	pool vk.CommandPool
}

// NewGpuTranslator creates a GpuTranslator, loads stub shaders and builds the stub pipeline layout.
func NewGpuTranslator(handles VulkanHandles, bknd backend.Backend[glfwvulkanbackend.GLFWWindowFlags]) (*GpuTranslator, error) {
	t := &GpuTranslator{
		handles:              handles,
		backend:              bknd,
		surfacesMutex:        sync.Mutex{},
		surfaces:             map[uintptr]*GpuSurface{},
		shadersMutex:         sync.Mutex{},
		shaders:              map[uintptr]*SpirvShader{},
		shaderModulesMutex:   sync.Mutex{},
		shaderModules:        map[uintptr]vk.ShaderModule{},
		pipelinesMutex:       sync.Mutex{},
		pipelines:            map[GpuTranslatorPipelineKey]vk.Pipeline{},
		constRamBuffersMutex: sync.Mutex{},
		constRamBuffers:      map[uint32]vk.Buffer{},
		constRamBufferMems:   map[uint32]vk.DeviceMemory{},
		userDataBuffersMutex: sync.Mutex{},
		userDataBuffers:      map[uint32]vk.Buffer{},
		userDataBufferMems:   map[uint32]vk.DeviceMemory{},
	}
	if err := t.createCommandPool(); err != nil {
		return nil, fmt.Errorf("GpuTranslator: command pool: %w", err)
	}
	if err := t.loadStubShaders(); err != nil {
		return nil, fmt.Errorf("GpuTranslator: stub shaders: %w", err)
	}
	if err := t.createStubPipelineLayout(); err != nil {
		return nil, fmt.Errorf("GpuTranslator: pipeline layout: %w", err)
	}

	return t, nil
}

// Destroy frees all Vulkan resources.
func (t *GpuTranslator) Destroy() {
	vk.DeviceWaitIdle(t.handles.Device)
	t.surfacesMutex.Lock()
	for _, s := range t.surfaces {
		s.Destroy(t.handles.Device)
	}
	t.surfacesMutex.Unlock()
	t.pipelinesMutex.Lock()
	for _, p := range t.pipelines {
		vk.DestroyPipeline(t.handles.Device, p, nil)
	}
	t.pipelinesMutex.Unlock()
	t.constRamBuffersMutex.Lock()
	for h, b := range t.constRamBuffers {
		vk.DestroyBuffer(t.handles.Device, b, nil)
		vk.FreeMemory(t.handles.Device, t.constRamBufferMems[h], nil)
	}
	t.constRamBuffersMutex.Unlock()
	t.userDataBuffersMutex.Lock()
	for h, b := range t.userDataBuffers {
		vk.DestroyBuffer(t.handles.Device, b, nil)
		vk.FreeMemory(t.handles.Device, t.userDataBufferMems[h], nil)
	}
	t.userDataBuffersMutex.Unlock()
	t.shaderModulesMutex.Lock()
	for _, m := range t.shaderModules {
		vk.DestroyShaderModule(t.handles.Device, m, nil)
	}
	t.shaderModulesMutex.Unlock()
	if t.stubPipelineLayout != vk.NullPipelineLayout {
		vk.DestroyPipelineLayout(t.handles.Device, t.stubPipelineLayout, nil)
	}
	if t.stubVertShader != vk.NullShaderModule {
		vk.DestroyShaderModule(t.handles.Device, t.stubVertShader, nil)
	}
	if t.stubFragShader != vk.NullShaderModule {
		vk.DestroyShaderModule(t.handles.Device, t.stubFragShader, nil)
	}
	if t.pool != vk.NullCommandPool {
		vk.DestroyCommandPool(t.handles.Device, t.pool, nil)
	}
}

// Translate translates a slice of DrawCalls into Vulkan commands and returns the command buffer.
func (t *GpuTranslator) Translate(draws []LiverpoolDrawCall) *vk.CommandBuffer {
	if len(draws) == 0 {
		return nil
	}

	// Update buffers holding const ram.
	t.UpdateConstRamBuffers(draws)
	t.UpdateUserDataBuffers(draws)

	// Begin recording.
	commandBuffer := t.handles.AllocateCommandBuffer(t.pool)
	vk.BeginCommandBuffer(commandBuffer, &vk.CommandBufferBeginInfo{
		SType: vk.StructureTypeCommandBufferBeginInfo,
		Flags: vk.CommandBufferUsageFlags(vk.CommandBufferUsageOneTimeSubmitBit),
	})
	for i := range draws {
		t.recordDraw(commandBuffer, &draws[i])
	}
	vk.EndCommandBuffer(commandBuffer)

	return &commandBuffer
}

func (t *GpuTranslator) GetShader(drawShader *gcn.GcnShader) *SpirvShader {
	// Get already loaded shader.
	t.shadersMutex.Lock()
	shader, ok := t.shaders[drawShader.Address]
	t.shadersMutex.Unlock()
	if ok {
		return shader
	}

	// Load the shader.
	t.shadersMutex.Lock()
	shader, err := NewSpirvShader(drawShader, SpirvShaderContext{})
	if err != nil {
		panic(err)
	}
	if err = t.DumpShaderOnce(shader); err != nil {
		panic(err)
	}
	t.shaders[drawShader.Address] = shader
	t.shadersMutex.Unlock()

	return shader
}

func (t *GpuTranslator) GetShaderModule(shader *SpirvShader) (vk.ShaderModule, error) {
	// Get already created shader module.
	t.shaderModulesMutex.Lock()
	mod, ok := t.shaderModules[shader.Address]
	t.shaderModulesMutex.Unlock()
	if ok {
		return mod, nil
	}

	// Create the shader module.
	var module vk.ShaderModule
	result := vk.CreateShaderModule(t.handles.Device, &vk.ShaderModuleCreateInfo{
		SType:    vk.StructureTypeShaderModuleCreateInfo,
		CodeSize: uint64(len(shader.Code) * 4),
		PCode:    shader.Code,
	}, nil, &module)
	if err := as.NewError(result); err != nil {
		return vk.NullShaderModule, fmt.Errorf("vkCreateShaderModule 0x%X: %w", shader.Address, err)
	}
	t.shaderModulesMutex.Lock()
	t.shaderModules[shader.Address] = module
	t.shaderModulesMutex.Unlock()

	return module, nil
}

func (t *GpuTranslator) GetPipeline(key GpuTranslatorPipelineKey, psModule vk.ShaderModule, renderPass vk.RenderPass, width, height uint32) (vk.Pipeline, error) {
	// Get already created pipeline.
	t.pipelinesMutex.Lock()
	pipeline, ok := t.pipelines[key]
	t.pipelinesMutex.Unlock()
	if ok {
		return pipeline, nil
	}

	// Create the pipeline.
	pipeline, err := t.createPipelineFromModules(t.stubVertShader, psModule, renderPass, width, height)
	if err != nil {
		return vk.NullPipeline, fmt.Errorf("createCompiledPipeline 0x%X: %w", key.PixelShaderAddress, err)
	}
	t.pipelinesMutex.Lock()
	t.pipelines[key] = pipeline
	t.pipelinesMutex.Unlock()

	return pipeline, nil
}

func (t *GpuTranslator) GetSurface(address uintptr, width, height uint32) (imgui.TextureRef, error) {
	// Check if it already exists.
	t.surfacesMutex.Lock()
	surface, ok := t.surfaces[address]
	t.surfacesMutex.Unlock()
	if ok {
		return surface.TextureId, nil
	}

	// Create a new one.
	surface = &GpuSurface{
		GPUAddress: address,
		Width:      width,
		Height:     height,
		Format:     vk.FormatR8g8b8a8Unorm,
		firstUse:   true,
	}
	if err := t.allocSurface(surface); err != nil {
		return imgui.TextureRef{}, fmt.Errorf("RegisterSurface 0x%X: %w", address, err)
	}
	surface.TextureId = t.backend.CreateVulkanTexture(surface.sampler, surface.imageView, vk.ImageLayoutShaderReadOnlyOptimal)
	t.surfacesMutex.Lock()
	t.surfaces[address] = surface
	t.surfacesMutex.Unlock()

	return surface.TextureId, nil
}

// GetSurfaceImageView returns the VkImageView for a registered surface so the renderer can display it as a texture.
// Returns vk.NullImageView if unknown.
func (t *GpuTranslator) GetSurfaceImageView(gpuAddress uintptr) vk.ImageView {
	t.surfacesMutex.Lock()
	defer t.surfacesMutex.Unlock()
	if s, ok := t.surfaces[gpuAddress]; ok {
		return s.imageView
	}
	return vk.NullImageView
}

func (t *GpuTranslator) GetBufferAddress(buffer vk.Buffer) uint64 {
	return uint64(GetBufferDeviceAddress(t.handles.Instance, t.handles.Device, buffer))
}

func (t *GpuTranslator) FreeBuffer(commandBuffer vk.CommandBuffer) {
	vk.FreeCommandBuffers(t.handles.Device, t.pool, 1, []vk.CommandBuffer{commandBuffer})
}
