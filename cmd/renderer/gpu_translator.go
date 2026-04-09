package renderer

import (
	"fmt"
	"sync"

	"github.com/LamkasDev/sharkie/cmd/structs/gcn"
	. "github.com/LamkasDev/sharkie/cmd/structs/gpu"
	. "github.com/LamkasDev/sharkie/cmd/structs/spirv"
	"github.com/elokore/cimgui-go-vulkan/backend"
	glfwvulkanbackend "github.com/elokore/cimgui-go-vulkan/backend/glfwvulkan-backend"
	"github.com/elokore/cimgui-go-vulkan/imgui"
	vk "github.com/vulkan-go/vulkan"
)

// GpuTranslator converts decoded DrawCalls into Vulkan commands.
type GpuTranslator struct {
	handles VulkanHandles
	backend backend.Backend[glfwvulkanbackend.GLFWWindowFlags]

	// Vulkan surfaces mirroring guest framebuffers.
	surfacesMutex sync.Mutex
	surfaces      map[uintptr]*GpuSurface

	// Stub pipeline shared across all draws until real shaders are available.
	stubPipelineLayout vk.PipelineLayout
	stubPipeline       vk.Pipeline
	stubVertShader     vk.ShaderModule
	stubFragShader     vk.ShaderModule

	// Recompiled SPIR-V shaders mirroring Liverpool.LoadedShaders.
	shadersMutex sync.Mutex
	shaders      map[uintptr]*SpirvShader

	// Command pool/buffer for this frame's GPU work.
	pool vk.CommandPool
}

// NewGpuTranslator creates a GpuTranslator, loads stub shaders and builds the stub pipeline layout.
func NewGpuTranslator(handles VulkanHandles, bknd backend.Backend[glfwvulkanbackend.GLFWWindowFlags]) (*GpuTranslator, error) {
	t := &GpuTranslator{
		handles:       handles,
		backend:       bknd,
		surfacesMutex: sync.Mutex{},
		surfaces:      map[uintptr]*GpuSurface{},
		shadersMutex:  sync.Mutex{},
		shaders:       map[uintptr]*SpirvShader{},
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
	if t.stubPipeline != vk.NullPipeline {
		vk.DestroyPipeline(t.handles.Device, t.stubPipeline, nil)
	}
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

// RegisterSurface registers a GPU address as a Vulkan render target.
func (t *GpuTranslator) RegisterSurface(address uintptr, width, height uint32) (imgui.TextureRef, error) {
	// Check if it already exists.
	t.surfacesMutex.Lock()
	surface, exists := t.surfaces[address]
	t.surfacesMutex.Unlock()
	if exists {
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

// Translate translates a slice of DrawCalls into Vulkan commands and returns the command buffer.
func (t *GpuTranslator) Translate(draws []LiverpoolDrawCall) *vk.CommandBuffer {
	if len(draws) == 0 {
		return nil
	}

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

func (t *GpuTranslator) FreeBuffer(commandBuffer vk.CommandBuffer) {
	vk.FreeCommandBuffers(t.handles.Device, t.pool, 1, []vk.CommandBuffer{commandBuffer})
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

// SurfaceImageView returns the VkImageView for a registered surface so the renderer can display it as a texture.
// Returns vk.NullImageView if unknown.
func (t *GpuTranslator) SurfaceImageView(gpuAddress uintptr) vk.ImageView {
	t.surfacesMutex.Lock()
	defer t.surfacesMutex.Unlock()
	if s, ok := t.surfaces[gpuAddress]; ok {
		return s.imageView
	}
	return vk.NullImageView
}
