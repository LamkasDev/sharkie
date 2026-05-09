package vulkan

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/LamkasDev/cimgui-go-vulkan/backend"
	glfwvulkanbackend "github.com/LamkasDev/cimgui-go-vulkan/backend/glfwvulkan-backend"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/spirv"
	"github.com/LamkasDev/sharkie/cmd/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs/gpu"
	vk "github.com/goki/vulkan"
	"github.com/gookit/color"
)

// GpuTranslator converts decoded DrawCalls into Vulkan commands.
type GpuTranslator struct {
	handles VulkanHandles
	backend backend.Backend[glfwvulkanbackend.GLFWWindowFlags]

	// Vulkan surfaces mirroring guest framebuffers.
	surfacesMutex sync.Mutex
	surfaces      map[uintptr]*GpuSurface

	// Stub pipeline shared across all draws.
	stubDescriptorSetLayout      vk.DescriptorSetLayout
	texelDescriptorSetLayout     vk.DescriptorSetLayout
	discoveryDescriptorSetLayout vk.DescriptorSetLayout
	stubPipelineLayout           vk.PipelineLayout

	// Descriptor sets.
	descriptorPool          vk.DescriptorPool
	texelDescriptorSets     []vk.DescriptorSet
	texelDescriptorSetIndex uint32
	discoveryDescriptorSet  vk.DescriptorSet
	bindlessDescriptorSet   vk.DescriptorSet

	// Discovery buffers.
	discoveryMapBuffer       vk.Buffer
	discoveryMapMem          vk.DeviceMemory
	discoveryMap             map[[12]uint32]uint32
	discoveryNextVulkanIndex uint32
	missingResourceBuffer    vk.Buffer
	missingResourceMem       vk.DeviceMemory

	// Recompiled SPIR-V shaders mirroring Liverpool.LoadedShaders.
	shadersMutex sync.Mutex
	shaders      map[uintptr]*spirv.SpirvShader

	// VkShaderModules created from SPIR-V shaders.
	shaderModulesMutex sync.Mutex
	shaderModules      map[uintptr]vk.ShaderModule

	// Per-draw compiled pipelines.
	pipelinesMutex sync.Mutex
	pipelines      map[GpuTranslatorPipelineKey]vk.Pipeline

	// Caches for images, image views and samplers.
	imagesMutex   sync.Mutex
	images        map[uintptr]vk.Image
	imageViews    map[uintptr]vk.ImageView
	imageMems     map[uintptr]vk.DeviceMemory
	samplersMutex sync.Mutex
	samplers      map[uint64]vk.Sampler

	// Physical buffers for User Data snapshots.
	userDataBuffersMutex sync.Mutex
	userDataBuffers      map[uint32]vk.Buffer
	userDataBuffersDebug map[uint32][]uint32
	userDataBufferMems   map[uint32]vk.DeviceMemory

	// Command pool/buffer for this frame's GPU work.
	pool vk.CommandPool
}

// NewGpuTranslator creates a GpuTranslator, loads stub shaders and builds the stub pipeline layout.
func NewGpuTranslator(handles VulkanHandles, bknd backend.Backend[glfwvulkanbackend.GLFWWindowFlags]) (*GpuTranslator, error) {
	t := &GpuTranslator{
		handles: handles,
		backend: bknd,

		surfacesMutex:      sync.Mutex{},
		surfaces:           map[uintptr]*GpuSurface{},
		shadersMutex:       sync.Mutex{},
		shaders:            map[uintptr]*spirv.SpirvShader{},
		shaderModulesMutex: sync.Mutex{},
		shaderModules:      map[uintptr]vk.ShaderModule{},
		pipelinesMutex:     sync.Mutex{},
		pipelines:          map[GpuTranslatorPipelineKey]vk.Pipeline{},

		discoveryMap:             map[[12]uint32]uint32{},
		discoveryNextVulkanIndex: 1, // 0 is reserved for missing.

		imagesMutex: sync.Mutex{},
		images:      map[uintptr]vk.Image{},
		imageViews:  map[uintptr]vk.ImageView{},
		imageMems:   map[uintptr]vk.DeviceMemory{},

		samplersMutex: sync.Mutex{},
		samplers:      map[uint64]vk.Sampler{},

		userDataBuffersMutex: sync.Mutex{},
		userDataBuffers:      map[uint32]vk.Buffer{},
		userDataBuffersDebug: map[uint32][]uint32{},
		userDataBufferMems:   map[uint32]vk.DeviceMemory{},
	}

	// Allocate memory buffers.
	onionBuffer, onionMemory, err := t.AllocExternalBuffer(vk.DeviceSize(structs.GlobalAllocator.Size),
		vk.BufferUsageFlags(vk.BufferUsageShaderDeviceAddressBit|vk.BufferUsageStorageBufferBit|vk.BufferUsageVertexBufferBit|vk.BufferUsageIndexBufferBit|vk.BufferUsageUniformTexelBufferBit|vk.BufferUsageTransferSrcBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit|vk.MemoryPropertyHostCachedBit))
	if err != nil {
		return nil, fmt.Errorf("GpuTranslator: onion buffer: %w", err)
	}
	structs.GlobalAllocator.Buffer = onionBuffer
	structs.GlobalAllocator.DeviceAddress = t.GetBufferAddress(onionBuffer)

	garlicBuffer, garlicMemory, err := t.AllocExternalBuffer(vk.DeviceSize(structs.GlobalGpuAllocator.Size),
		vk.BufferUsageFlags(vk.BufferUsageShaderDeviceAddressBit|vk.BufferUsageStorageBufferBit|vk.BufferUsageVertexBufferBit|vk.BufferUsageIndexBufferBit|vk.BufferUsageUniformTexelBufferBit|vk.BufferUsageTransferSrcBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
	if err != nil {
		return nil, fmt.Errorf("GpuTranslator: garlic buffer: %w", err)
	}
	structs.GlobalGpuAllocator.Buffer = garlicBuffer
	structs.GlobalGpuAllocator.DeviceAddress = t.GetBufferAddress(garlicBuffer)

	// Map memory buffers.
	if runtime.GOOS == "windows" {
		err = structs.MapVulkanMemory(structs.GlobalAllocator.Base, structs.GlobalAllocator.Size, GetMemoryWin32Handle(t.handles.Instance, t.handles.Device, onionMemory))
		if err != nil {
			return nil, fmt.Errorf("GpuTranslator: map onion buffer: %w", err)
		}
		err = structs.MapVulkanMemory(structs.GlobalGpuAllocator.Base, structs.GlobalGpuAllocator.Size, GetMemoryWin32Handle(t.handles.Instance, t.handles.Device, garlicMemory))
		if err != nil {
			return nil, fmt.Errorf("GpuTranslator: map garlic buffer: %w", err)
		}
	} else {
		err = structs.MapVulkanMemory(structs.GlobalAllocator.Base, structs.GlobalAllocator.Size, uintptr(GetMemoryFd(t.handles.Instance, t.handles.Device, onionMemory)))
		if err != nil {
			return nil, fmt.Errorf("GpuTranslator: map onion buffer: %w", err)
		}
		err = structs.MapVulkanMemory(structs.GlobalGpuAllocator.Base, structs.GlobalGpuAllocator.Size, uintptr(GetMemoryFd(t.handles.Instance, t.handles.Device, garlicMemory)))
		if err != nil {
			return nil, fmt.Errorf("GpuTranslator: map garlic buffer: %w", err)
		}
	}

	if err = t.createCommandPool(); err != nil {
		return nil, fmt.Errorf("GpuTranslator: command pool: %w", err)
	}
	if err = t.createStubPipelineLayout(); err != nil {
		return nil, fmt.Errorf("GpuTranslator: pipeline layout: %w", err)
	}
	if err = t.createDescriptorPool(); err != nil {
		return nil, fmt.Errorf("GpuTranslator: descriptor pool: %w", err)
	}
	if err = t.CreateDiscoveryBuffers(); err != nil {
		return nil, fmt.Errorf("GpuTranslator: discovery buffers: %w", err)
	}
	t.updateDiscoveryDescriptorSet()
	t.createDummyTexture()

	return t, nil
}

// Destroy frees all Vulkan resources.
func (t *GpuTranslator) Destroy() {
	vk.DeviceWaitIdle(t.handles.Device)
	if t.descriptorPool != vk.NullDescriptorPool {
		vk.DestroyDescriptorPool(t.handles.Device, t.descriptorPool, nil)
	}
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
	t.userDataBuffersMutex.Lock()
	for h, b := range t.userDataBuffers {
		vk.DestroyBuffer(t.handles.Device, b, nil)
		vk.FreeMemory(t.handles.Device, t.userDataBufferMems[h], nil)
	}
	t.userDataBuffersMutex.Unlock()
	if t.discoveryMapBuffer != vk.NullBuffer {
		vk.DestroyBuffer(t.handles.Device, t.discoveryMapBuffer, nil)
		vk.FreeMemory(t.handles.Device, t.discoveryMapMem, nil)
	}
	if t.missingResourceBuffer != vk.NullBuffer {
		vk.DestroyBuffer(t.handles.Device, t.missingResourceBuffer, nil)
		vk.FreeMemory(t.handles.Device, t.missingResourceMem, nil)
	}
	t.shaderModulesMutex.Lock()
	for _, m := range t.shaderModules {
		vk.DestroyShaderModule(t.handles.Device, m, nil)
	}
	t.shaderModulesMutex.Unlock()
	if t.stubPipelineLayout != vk.NullPipelineLayout {
		vk.DestroyPipelineLayout(t.handles.Device, t.stubPipelineLayout, nil)
	}
	if t.texelDescriptorSetLayout != vk.NullDescriptorSetLayout {
		vk.DestroyDescriptorSetLayout(t.handles.Device, t.texelDescriptorSetLayout, nil)
	}
	if t.discoveryDescriptorSetLayout != vk.NullDescriptorSetLayout {
		vk.DestroyDescriptorSetLayout(t.handles.Device, t.discoveryDescriptorSetLayout, nil)
	}
	if t.stubDescriptorSetLayout != vk.NullDescriptorSetLayout {
		vk.DestroyDescriptorSetLayout(t.handles.Device, t.stubDescriptorSetLayout, nil)
	}
	if t.pool != vk.NullCommandPool {
		vk.DestroyCommandPool(t.handles.Device, t.pool, nil)
	}
}

// Translate translates Liverpool draw/compute commands into Vulkan commands and returns the command buffer.
func (t *GpuTranslator) Translate(frame uint64, draws []LiverpoolDrawCall, dispatches []LiverpoolComputeDispatch) *vk.CommandBuffer {
	if len(draws) == 0 && len(dispatches) == 0 {
		return nil
	}

	// Reset per-frame state.
	t.texelDescriptorSetIndex = 0
	t.FulfillResources(frame)

	// Update buffers holding user data.
	logger.Printf("[%s] updating buffers.\n",
		color.Blue.Sprintf("Frame %d", frame),
	)
	t.UpdateUserDataBuffers(draws)

	// Begin recording.
	commandBuffer := t.handles.AllocateCommandBuffer(t.pool)
	vk.BeginCommandBuffer(commandBuffer, &vk.CommandBufferBeginInfo{
		SType: vk.StructureTypeCommandBufferBeginInfo,
		Flags: vk.CommandBufferUsageFlags(vk.CommandBufferUsageOneTimeSubmitBit),
	})

	// Process compute dispatches.
	logger.Printf("[%s] processing %s compute dispatches.\n",
		color.Blue.Sprintf("Frame %d", frame),
		color.Blue.Sprint(len(dispatches)),
	)
	for i := range dispatches {
		t.dispatchCompute(frame, commandBuffer, &dispatches[i])
	}

	// Record draw calls.
	logger.Printf("[%s] recording %s draw calls.\n",
		color.Blue.Sprintf("Frame %d", frame),
		color.Blue.Sprint(len(draws)),
	)
	for i := range draws {
		t.recordDraw(frame, commandBuffer, &draws[i])
	}

	// Finish.
	vk.EndCommandBuffer(commandBuffer)

	return &commandBuffer
}

func (t *GpuTranslator) GetBufferAddress(buffer vk.Buffer) uint64 {
	return uint64(GetBufferDeviceAddress(t.handles.Instance, t.handles.Device, buffer))
}

func (t *GpuTranslator) FreeBuffer(commandBuffer vk.CommandBuffer) {
	vk.FreeCommandBuffers(t.handles.Device, t.pool, 1, []vk.CommandBuffer{commandBuffer})
}
