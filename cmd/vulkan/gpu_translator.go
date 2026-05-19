package vulkan

import (
	"fmt"
	"os"
	"runtime"
	"sync"

	"github.com/LamkasDev/cimgui-go-vulkan/backend"
	glfwvulkanbackend "github.com/LamkasDev/cimgui-go-vulkan/backend/glfwvulkan-backend"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/spirv"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/LamkasDev/sharkie/cmd/structs/gpu"
	vk "github.com/goki/vulkan"
	"github.com/gookit/color"
)

// GpuTranslator converts decoded DrawCalls into Vulkan commands.
type GpuTranslator struct {
	handles VulkanHandles
	backend backend.Backend[glfwvulkanbackend.GLFWWindowFlags]

	// Vulkan surfaces mirroring guest framebuffers.
	surfacesMutex sync.Mutex
	surfaces      map[SurfaceKey]*GpuSurface

	// Pipeline shared across all draws.
	bindlessDescriptorSetLayout  vk.DescriptorSetLayout
	discoveryDescriptorSetLayout vk.DescriptorSetLayout
	pipelineLayout               vk.PipelineLayout

	// Descriptor sets.
	descriptorPool         vk.DescriptorPool
	discoveryDescriptorSet vk.DescriptorSet
	bindlessDescriptorSet  vk.DescriptorSet

	// Discovery buffers.
	discoveryMapBuffer         vk.Buffer
	discoveryMapMem            vk.DeviceMemory
	discoveryImageSamplerMap   map[spirvStructs.ImageSamplerKey]uint32
	discoveryImageNoSamplerMap map[spirvStructs.ImageNoSamplerKey]uint32
	discoveryNextVulkanIndex   uint32
	missingResourceBuffer      vk.Buffer
	missingResourceMem         vk.DeviceMemory

	// Recompiled SPIR-V shaders mirroring Liverpool.LoadedShaders.
	shadersMutex sync.Mutex
	shaders      map[SpirvShaderKey]*spirv.SpirvShader

	// VkShaderModules created from SPIR-V shaders.
	shaderModulesMutex   sync.Mutex
	shaderModules        map[uintptr]vk.ShaderModule
	rectlistShaderModule vk.ShaderModule

	// Per-draw compiled pipelines.
	pipelinesMutex sync.Mutex
	pipelines      map[GraphicsPipelineKey]vk.Pipeline

	// Per-dispatch compiled compute pipelines.
	computePipelinesMutex sync.Mutex
	computePipelines      map[ComputePipelineKey]vk.Pipeline

	// Per-draw framebuffers.
	framebuffersMutex sync.Mutex
	framebuffers      map[FramebufferRequest]*VulkanFramebuffer

	// Caches for images, image views and samplers.
	imagesMutex      sync.Mutex
	images           map[uintptr]vk.Image
	imageMems        map[uintptr]vk.DeviceMemory
	imageViews       map[uint64]vk.ImageView
	imageDescriptors map[uint64]spirvStructs.ImageDescriptor
	samplersMutex    sync.Mutex
	samplers         map[uint64]vk.Sampler
	defaultSampler   vk.Sampler

	// Physical buffers for User Data snapshots.
	userDataBuffersMutex  sync.Mutex
	userDataBuffer        vk.Buffer
	userDataBufferMem     vk.DeviceMemory
	userDataBufferAddress uint64
	userDataOffsets       map[uint32]uint32

	// Command pool/buffer for this frame's GPU work.
	pool        vk.CommandPool
	fence       vk.Fence
	workerFence vk.Fence
	QueueMutex  *sync.Mutex
}

// NewGpuTranslator creates a GpuTranslator, loads stub shaders and builds the stub pipeline layout.
func NewGpuTranslator(handles VulkanHandles, bknd backend.Backend[glfwvulkanbackend.GLFWWindowFlags], queueMutex *sync.Mutex) (*GpuTranslator, error) {
	if err := os.MkdirAll("temp/shaders", 0777); err != nil {
		return nil, fmt.Errorf("GpuTranslator: create temp/shaders directory: %w", err)
	}

	t := &GpuTranslator{
		handles: handles,
		backend: bknd,

		surfacesMutex:      sync.Mutex{},
		surfaces:           map[SurfaceKey]*GpuSurface{},
		shadersMutex:       sync.Mutex{},
		shaders:            map[SpirvShaderKey]*spirv.SpirvShader{},
		shaderModulesMutex: sync.Mutex{},
		shaderModules:      map[uintptr]vk.ShaderModule{},

		pipelinesMutex: sync.Mutex{},
		pipelines:      map[GraphicsPipelineKey]vk.Pipeline{},

		computePipelinesMutex: sync.Mutex{},
		computePipelines:      map[ComputePipelineKey]vk.Pipeline{},

		framebuffersMutex: sync.Mutex{},
		framebuffers:      map[FramebufferRequest]*VulkanFramebuffer{},

		discoveryImageSamplerMap:   map[spirvStructs.ImageSamplerKey]uint32{},
		discoveryImageNoSamplerMap: map[spirvStructs.ImageNoSamplerKey]uint32{},
		discoveryNextVulkanIndex:   1, // 0 is reserved for missing.

		imagesMutex:      sync.Mutex{},
		images:           map[uintptr]vk.Image{},
		imageMems:        map[uintptr]vk.DeviceMemory{},
		imageViews:       map[uint64]vk.ImageView{},
		imageDescriptors: map[uint64]spirvStructs.ImageDescriptor{},

		samplersMutex: sync.Mutex{},
		samplers:      map[uint64]vk.Sampler{},

		userDataBuffersMutex: sync.Mutex{},
		userDataOffsets:      map[uint32]uint32{},

		QueueMutex: queueMutex,
	}

	// Allocate user data buffer.
	userDataBuffer, userDataBufferMem, err := t.AllocBuffer(vk.DeviceSize(UserDataBufferSize),
		vk.BufferUsageFlags(vk.BufferUsageShaderDeviceAddressBit|vk.BufferUsageUniformBufferBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
	if err != nil {
		return nil, fmt.Errorf("GpuTranslator: user data buffer: %w", err)
	}
	t.userDataBuffer = userDataBuffer
	t.userDataBufferMem = userDataBufferMem
	t.userDataBufferAddress = t.GetBufferAddress(userDataBuffer)

	// Allocate memory buffers.
	onionBuffer, onionMemory, err := t.AllocExternalBuffer(vk.DeviceSize(GlobalAllocator.Size),
		vk.BufferUsageFlags(vk.BufferUsageShaderDeviceAddressBit|vk.BufferUsageStorageBufferBit|vk.BufferUsageVertexBufferBit|vk.BufferUsageIndexBufferBit|vk.BufferUsageTransferSrcBit|vk.BufferUsageTransferDstBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit|vk.MemoryPropertyHostCachedBit))
	if err != nil {
		return nil, fmt.Errorf("GpuTranslator: onion buffer: %w", err)
	}
	GlobalAllocator.Buffer = onionBuffer
	GlobalAllocator.DeviceAddress = t.GetBufferAddress(onionBuffer)

	garlicBuffer, garlicMemory, err := t.AllocExternalBuffer(vk.DeviceSize(GlobalGpuAllocator.Size),
		vk.BufferUsageFlags(vk.BufferUsageShaderDeviceAddressBit|vk.BufferUsageStorageBufferBit|vk.BufferUsageVertexBufferBit|vk.BufferUsageIndexBufferBit|vk.BufferUsageTransferSrcBit|vk.BufferUsageTransferDstBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
	if err != nil {
		return nil, fmt.Errorf("GpuTranslator: garlic buffer: %w", err)
	}
	GlobalGpuAllocator.Buffer = garlicBuffer
	GlobalGpuAllocator.DeviceAddress = t.GetBufferAddress(garlicBuffer)

	// Map memory buffers.
	if runtime.GOOS == "windows" {
		err = MapVulkanMemory(GlobalAllocator.Base, GlobalAllocator.Size, GetMemoryWin32Handle(t.handles.Instance, t.handles.Device, onionMemory))
		if err != nil {
			return nil, fmt.Errorf("GpuTranslator: map onion buffer: %w", err)
		}
		err = MapVulkanMemory(GlobalGpuAllocator.Base, GlobalGpuAllocator.Size, GetMemoryWin32Handle(t.handles.Instance, t.handles.Device, garlicMemory))
		if err != nil {
			return nil, fmt.Errorf("GpuTranslator: map garlic buffer: %w", err)
		}
	} else {
		err = MapVulkanMemory(GlobalAllocator.Base, GlobalAllocator.Size, uintptr(GetMemoryFd(t.handles.Instance, t.handles.Device, onionMemory)))
		if err != nil {
			return nil, fmt.Errorf("GpuTranslator: map onion buffer: %w", err)
		}
		err = MapVulkanMemory(GlobalGpuAllocator.Base, GlobalGpuAllocator.Size, uintptr(GetMemoryFd(t.handles.Instance, t.handles.Device, garlicMemory)))
		if err != nil {
			return nil, fmt.Errorf("GpuTranslator: map garlic buffer: %w", err)
		}
	}

	if err = t.createCommandPoolAndFence(); err != nil {
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
	t.computePipelinesMutex.Lock()
	for _, p := range t.computePipelines {
		vk.DestroyPipeline(t.handles.Device, p, nil)
	}
	t.computePipelinesMutex.Unlock()
	t.userDataBuffersMutex.Lock()
	if t.userDataBuffer != vk.NullBuffer {
		vk.DestroyBuffer(t.handles.Device, t.userDataBuffer, nil)
		vk.FreeMemory(t.handles.Device, t.userDataBufferMem, nil)
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
	if t.pipelineLayout != vk.NullPipelineLayout {
		vk.DestroyPipelineLayout(t.handles.Device, t.pipelineLayout, nil)
	}
	if t.discoveryDescriptorSetLayout != vk.NullDescriptorSetLayout {
		vk.DestroyDescriptorSetLayout(t.handles.Device, t.discoveryDescriptorSetLayout, nil)
	}
	if t.bindlessDescriptorSetLayout != vk.NullDescriptorSetLayout {
		vk.DestroyDescriptorSetLayout(t.handles.Device, t.bindlessDescriptorSetLayout, nil)
	}
	if t.pool != vk.NullCommandPool {
		vk.DestroyCommandPool(t.handles.Device, t.pool, nil)
	}
}

// Translate translates Liverpool draw/compute commands into Vulkan commands and returns the command buffer.
func (t *GpuTranslator) Translate(frame uint64, draws []gpu.LiverpoolDrawCall, dispatches []gpu.LiverpoolComputeDispatch, copies []gpu.LiverpoolDmaCopy) *vk.CommandBuffer {
	// Update buffers holding user data.
	t.UpdateUserDataBuffers(draws, dispatches)

	// Begin recording.
	commandBuffer := t.handles.AllocateCommandBuffer(t.pool)
	vk.BeginCommandBuffer(commandBuffer, &vk.CommandBufferBeginInfo{
		SType: vk.StructureTypeCommandBufferBeginInfo,
		Flags: vk.CommandBufferUsageFlags(vk.CommandBufferUsageOneTimeSubmitBit),
	})

	// Process DMA copies.
	if len(copies) > 0 {
		logger.Printf("[%s] processing %s DMA copies.\n",
			color.Blue.Sprintf("Frame %d", frame),
			color.Blue.Sprint(len(copies)),
		)
		for i := range copies {
			t.processDmaCopy(frame, commandBuffer, &copies[i])
		}
	}

	// Process compute dispatches.
	if len(dispatches) > 0 {
		logger.Printf("[%s] processing %s compute dispatches.\n",
			color.Blue.Sprintf("Frame %d", frame),
			color.Blue.Sprint(len(dispatches)),
		)
		for i := range dispatches {
			t.dispatchCompute(frame, commandBuffer, &dispatches[i])
		}
	}

	// Record draw calls.
	if len(draws) > 0 {
		logger.Printf("[%s] recording %s draw calls.\n",
			color.Blue.Sprintf("Frame %d", frame),
			color.Blue.Sprint(len(draws)),
		)
		for i := range draws {
			t.recordDraw(frame, commandBuffer, &draws[i])
		}
	}

	// Finish.
	vk.EndCommandBuffer(commandBuffer)

	return &commandBuffer
}

func (t *GpuTranslator) GetFence() vk.Fence {
	return t.fence
}

func (t *GpuTranslator) WaitOnFence() {
	vk.WaitForFences(t.handles.Device, 1, []vk.Fence{t.fence}, vk.True, ^uint64(0))
}

func (t *GpuTranslator) SignalFence() {
	t.QueueMutex.Lock()
	defer t.QueueMutex.Unlock()
	vk.QueueSubmit(t.handles.GraphicsQueue, 0, nil, t.fence)
}

func (t *GpuTranslator) ResetFence() {
	t.QueueMutex.Lock()
	defer t.QueueMutex.Unlock()
	vk.ResetFences(t.handles.Device, 1, []vk.Fence{t.fence})
}

func (t *GpuTranslator) GetWorkerFence() vk.Fence {
	return t.workerFence
}

func (t *GpuTranslator) WaitOnWorkerFence() {
	vk.WaitForFences(t.handles.Device, 1, []vk.Fence{t.workerFence}, vk.True, ^uint64(0))
}

func (t *GpuTranslator) ResetWorkerFence() {
	t.QueueMutex.Lock()
	defer t.QueueMutex.Unlock()
	status := vk.GetFenceStatus(t.handles.Device, t.workerFence)
	if status == vk.Success {
		vk.ResetFences(t.handles.Device, 1, []vk.Fence{t.workerFence})
	}
}

func (t *GpuTranslator) AllocateCommandBuffer() vk.CommandBuffer {
	t.QueueMutex.Lock()
	defer t.QueueMutex.Unlock()
	return t.handles.AllocateCommandBuffer(t.pool)
}

func (t *GpuTranslator) FreeCommandBuffer(commandBuffer vk.CommandBuffer) {
	t.QueueMutex.Lock()
	defer t.QueueMutex.Unlock()
	vk.FreeCommandBuffers(t.handles.Device, t.pool, 1, []vk.CommandBuffer{commandBuffer})
}

func (t *GpuTranslator) GetBufferAddress(buffer vk.Buffer) uint64 {
	return uint64(GetBufferDeviceAddress(t.handles.Instance, t.handles.Device, buffer))
}
