package translation

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"unsafe"

	"github.com/LamkasDev/cimgui-go-vulkan/backend"
	glfwvulkanbackend "github.com/LamkasDev/cimgui-go-vulkan/backend/glfwvulkan-backend"
	"github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gpu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/spirv"
	"github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan"
	vk "github.com/goki/vulkan"
	"github.com/gookit/color"
)

// GpuTranslator converts decoded DrawCalls into Vulkan commands.
type GpuTranslator struct {
	handles vulkan.VulkanHandles
	backend backend.Backend[glfwvulkanbackend.GLFWWindowFlags]

	// Vulkan surfaces mirroring guest framebuffers.
	surfacesMutex sync.Mutex
	surfaces      map[uintptr]*vulkan.VulkanSurface

	// Pipeline shared across all draws.
	staticDescriptorSetLayout vk.DescriptorSetLayout
	pipelineLayout            vk.PipelineLayout

	// Descriptor sets.
	descriptorPool         vk.DescriptorPool
	staticDescriptorSet    vk.DescriptorSet
	staticDescriptorSets   []vk.DescriptorSet
	staticDescriptorSetIdx int

	// Recompiled SPIR-V shaders mirroring Liverpool.LoadedShaders.
	shadersMutex sync.Mutex
	shaders      map[SpirvShaderKey]*spirv.SpirvShader

	// VkShaderModules created from SPIR-V shaders.
	shaderModulesMutex      sync.Mutex
	shaderModules           map[uintptr]vk.ShaderModule
	rectlistTcsShaderModule vk.ShaderModule
	rectlistTesShaderModule vk.ShaderModule

	// Per-draw compiled pipelines.
	pipelinesMutex sync.Mutex
	pipelines      map[vulkan.GraphicsPipelineKey]vk.Pipeline

	// Per-dispatch compiled compute pipelines.
	computePipelinesMutex sync.Mutex
	computePipelines      map[vulkan.ComputePipelineKey]vk.Pipeline

	// Per-draw framebuffers.
	framebuffersMutex sync.Mutex
	framebuffers      map[vulkan.FramebufferRequest]*vulkan.VulkanFramebuffer

	// Resources destroyed after the current frame's command buffer completes.
	deferredDestroyMutex sync.Mutex
	deferredDestroy      deferredDestroyQueue

	// Caches for images, image views and samplers.
	imagesMutex    sync.Mutex
	imagePages     map[uintptr]map[uintptr]struct{}
	images         map[uintptr]*vulkan.VulkanImage
	imageViews     map[uint64]*vulkan.VulkanImageView
	samplersMutex  sync.Mutex
	samplers       map[uint64]vk.Sampler
	defaultSampler vk.Sampler

	// Physical buffers for User Data snapshots.
	userDataBuffersMutex  sync.Mutex
	userDataBuffer        vk.Buffer
	userDataBufferMem     vk.DeviceMemory
	userDataBufferAddress uint64
	userDataOffsets       map[uint32]uint32

	// Static index buffer for QuadList drawing.
	quadListIndexBuffer    vk.Buffer
	quadListIndexBufferMem vk.DeviceMemory

	// Command pool/buffer for this frame's GPU work.
	pool          vk.CommandPool
	commandBuffer vk.CommandBuffer
	fenceChan     chan struct{}
	fenceMutex    sync.Mutex

	// Active state for chronological stream processing.
	lastColorRtAddress   uintptr
	activeSurface        *vulkan.VulkanSurface
	activePass           vk.RenderPass
	activePassNoClear    vk.RenderPass
	activeFramebuffer    vk.Framebuffer
	activePipeline       vk.Pipeline
	activeFragmentShader *spirv.SpirvShader
	activeVteControl     uint32
	activeClipControl    uint32
	activeDynamicState   *gpu.LiverpoolSetDynamicState

	// lastProcessedFrame tracks which guest frame surface lifetime state belongs to.
	lastProcessedFrame uint64
	currentGuestFrame  uint64

	// pendingComputeBarrier is set after a compute dispatch; the next draw waits for it.
	pendingComputeBarrier bool
}

// NewGpuTranslator creates a GpuTranslator, loads stub shaders and builds the stub pipeline layout.
func NewGpuTranslator(handles vulkan.VulkanHandles, bknd backend.Backend[glfwvulkanbackend.GLFWWindowFlags]) (*GpuTranslator, error) {
	if err := os.MkdirAll("temp/shaders", 0777); err != nil {
		return nil, fmt.Errorf("GpuTranslator: create temp/shaders directory: %w", err)
	}

	t := &GpuTranslator{
		handles: handles,
		backend: bknd,

		surfacesMutex: sync.Mutex{},
		surfaces:      map[uintptr]*vulkan.VulkanSurface{},

		deferredDestroyMutex: sync.Mutex{},
		shadersMutex:         sync.Mutex{},
		shaders:              map[SpirvShaderKey]*spirv.SpirvShader{},
		shaderModulesMutex:   sync.Mutex{},
		shaderModules:        map[uintptr]vk.ShaderModule{},

		pipelinesMutex: sync.Mutex{},
		pipelines:      map[vulkan.GraphicsPipelineKey]vk.Pipeline{},

		computePipelinesMutex: sync.Mutex{},
		computePipelines:      map[vulkan.ComputePipelineKey]vk.Pipeline{},

		framebuffersMutex: sync.Mutex{},
		framebuffers:      map[vulkan.FramebufferRequest]*vulkan.VulkanFramebuffer{},

		imagePages: map[uintptr]map[uintptr]struct{}{},

		imagesMutex:   sync.Mutex{},
		images:        map[uintptr]*vulkan.VulkanImage{},
		imageViews:    map[uint64]*vulkan.VulkanImageView{},
		samplersMutex: sync.Mutex{},
		samplers:      map[uint64]vk.Sampler{},

		userDataBuffersMutex: sync.Mutex{},
		userDataOffsets:      map[uint32]uint32{},

		fenceChan:  make(chan struct{}),
		fenceMutex: sync.Mutex{},
	}
	close(t.fenceChan)

	// Allocate user data buffer.
	userDataBuffer, userDataBufferMem, err := vulkan.AllocateBuffer(&t.handles, vk.DeviceSize(UserDataBufferSize),
		vk.BufferUsageFlags(vk.BufferUsageShaderDeviceAddressBit|vk.BufferUsageUniformBufferBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
	if err != nil {
		return nil, fmt.Errorf("GpuTranslator: user data buffer: %w", err)
	}
	t.userDataBuffer = userDataBuffer
	t.userDataBufferMem = userDataBufferMem
	t.userDataBufferAddress = t.GetBufferAddress(userDataBuffer)

	// Onion and garlic use separate device allocations - Linux dma-buf cannot be mmap'd
	// twice from one fd into two fixed guest VA windows.
	onionBuffer, onionMemory, err := vulkan.AllocateExternalBuffer(&t.handles, vk.DeviceSize(GlobalAllocator.Size),
		vk.BufferUsageFlags(vk.BufferUsageShaderDeviceAddressBit|vk.BufferUsageStorageBufferBit|vk.BufferUsageVertexBufferBit|vk.BufferUsageIndexBufferBit|vk.BufferUsageTransferSrcBit|vk.BufferUsageTransferDstBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit|vk.MemoryPropertyHostCachedBit))
	if err != nil {
		return nil, fmt.Errorf("GpuTranslator: onion buffer: %w", err)
	}
	GlobalAllocator.Buffer = onionBuffer
	GlobalAllocator.DeviceAddress = t.GetBufferAddress(onionBuffer)
	GuestBacking.Buffer = onionBuffer
	GuestBacking.Memory = onionMemory
	GuestBacking.DeviceAddress = GlobalAllocator.DeviceAddress

	garlicBuffer, garlicMemory, err := vulkan.AllocateExternalBuffer(&t.handles, vk.DeviceSize(GlobalGpuAllocator.Size),
		vk.BufferUsageFlags(vk.BufferUsageShaderDeviceAddressBit|vk.BufferUsageStorageBufferBit|vk.BufferUsageVertexBufferBit|vk.BufferUsageIndexBufferBit|vk.BufferUsageTransferSrcBit|vk.BufferUsageTransferDstBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
	if err != nil {
		return nil, fmt.Errorf("GpuTranslator: garlic buffer: %w", err)
	}
	GlobalGpuAllocator.Buffer = garlicBuffer
	GlobalGpuAllocator.DeviceAddress = t.GetBufferAddress(garlicBuffer)

	if runtime.GOOS == "windows" {
		err = MapVulkanMemory(GlobalAllocator.Base, GlobalAllocator.Size, vulkan.GetMemoryWin32Handle(t.handles.Instance, t.handles.Device, onionMemory), 0)
		if err != nil {
			return nil, fmt.Errorf("GpuTranslator: map onion buffer: %w", err)
		}
		err = MapVulkanMemory(GlobalGpuAllocator.Base, GlobalGpuAllocator.Size, vulkan.GetMemoryWin32Handle(t.handles.Instance, t.handles.Device, garlicMemory), 0)
		if err != nil {
			return nil, fmt.Errorf("GpuTranslator: map garlic buffer: %w", err)
		}
	} else {
		err = MapVulkanMemory(GlobalAllocator.Base, GlobalAllocator.Size, uintptr(vulkan.GetMemoryFd(t.handles.Instance, t.handles.Device, onionMemory)), 0)
		if err != nil {
			return nil, fmt.Errorf("GpuTranslator: map onion buffer: %w", err)
		}
		err = MapVulkanMemory(GlobalGpuAllocator.Base, GlobalGpuAllocator.Size, uintptr(vulkan.GetMemoryFd(t.handles.Instance, t.handles.Device, garlicMemory)), 0)
		if err != nil {
			return nil, fmt.Errorf("GpuTranslator: map garlic buffer: %w", err)
		}
	}
	t.pipelineLayout, t.staticDescriptorSetLayout, err = vulkan.CreateStubPipelineLayout(&t.handles)
	if err != nil {
		return nil, fmt.Errorf("GpuTranslator: pipeline layout: %w", err)
	}
	t.descriptorPool, t.staticDescriptorSet, err = vulkan.CreateDescriptorPool(&t.handles, t.staticDescriptorSetLayout)
	if err != nil {
		return nil, fmt.Errorf("GpuTranslator: descriptor pool: %w", err)
	}
	t.activePass = vk.NullRenderPass
	t.activePipeline = vk.NullPipeline

	// Allocate quad list index buffer.
	const maxQuads = 16384
	const quadListIndexCount = maxQuads * 6
	quadListIndexBuffer, quadListIndexBufferMem, err := vulkan.AllocateBuffer(&t.handles, vk.DeviceSize(quadListIndexCount*2),
		vk.BufferUsageFlags(vk.BufferUsageIndexBufferBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
	if err != nil {
		return nil, fmt.Errorf("GpuTranslator: allocate quad list index buffer: %w", err)
	}

	// Map and fill quad list index buffer.
	var indexData unsafe.Pointer
	result := vk.MapMemory(handles.Device, quadListIndexBufferMem, 0, vk.DeviceSize(quadListIndexCount*2), 0, &indexData)
	if err := vulkan.NewError(result); err != nil {
		vk.DestroyBuffer(handles.Device, quadListIndexBuffer, nil)
		vk.FreeMemory(handles.Device, quadListIndexBufferMem, nil)
		return nil, fmt.Errorf("GpuTranslator: map quad list index buffer: %w", err)
	}

	indices := (*[quadListIndexCount]uint16)(indexData)
	for i := 0; i < maxQuads; i++ {
		baseVertex := uint16(i * 4)
		indices[i*6+0] = baseVertex + 0
		indices[i*6+1] = baseVertex + 1
		indices[i*6+2] = baseVertex + 2
		indices[i*6+3] = baseVertex + 0
		indices[i*6+4] = baseVertex + 2
		indices[i*6+5] = baseVertex + 3
	}
	vk.UnmapMemory(handles.Device, quadListIndexBufferMem)

	t.quadListIndexBuffer = quadListIndexBuffer
	t.quadListIndexBufferMem = quadListIndexBufferMem

	structs.GlobalMemoryManager.SetRasterizer(t)
	structs.GlobalMemoryManager.Guest().RegisterPreMapped()
	structs.GlobalMemoryManager.OnMapGuest(lib_structs.GlobalAllocator.Base, uintptr(lib_structs.GlobalAllocator.Size))
	structs.GlobalMemoryManager.OnMapGuest(lib_structs.GlobalGpuAllocator.Base, uintptr(lib_structs.GlobalGpuAllocator.Size))

	go pprof.Do(context.Background(), pprof.Labels("name", "MemorySyncWorker"), func(ctx context.Context) {
		t.memorySyncWorker()
	})

	return t, nil
}

// Destroy frees all Vulkan resources.
func (t *GpuTranslator) Destroy() {
	vk.DeviceWaitIdle(t.handles.Device)
	if t.descriptorPool != vk.NullDescriptorPool {
		vk.DestroyDescriptorPool(t.handles.Device, t.descriptorPool, nil)
	}
	t.FlushDeferredDestruction()
	t.surfacesMutex.Lock()
	for _, s := range t.surfaces {
		s.Destroy(t.handles.Device)
	}
	t.surfacesMutex.Unlock()
	t.framebuffersMutex.Lock()
	for _, fb := range t.framebuffers {
		fb.Destroy(t.handles.Device)
	}
	t.framebuffersMutex.Unlock()
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
	if t.quadListIndexBuffer != vk.NullBuffer {
		vk.DestroyBuffer(t.handles.Device, t.quadListIndexBuffer, nil)
		vk.FreeMemory(t.handles.Device, t.quadListIndexBufferMem, nil)
	}
	t.shaderModulesMutex.Lock()
	for _, m := range t.shaderModules {
		vk.DestroyShaderModule(t.handles.Device, m, nil)
	}
	t.shaderModulesMutex.Unlock()
	if t.pipelineLayout != vk.NullPipelineLayout {
		vk.DestroyPipelineLayout(t.handles.Device, t.pipelineLayout, nil)
	}

	if t.staticDescriptorSetLayout != vk.NullDescriptorSetLayout {
		vk.DestroyDescriptorSetLayout(t.handles.Device, t.staticDescriptorSetLayout, nil)
	}
	if t.pool != vk.NullCommandPool {
		vk.DestroyCommandPool(t.handles.Device, t.pool, nil)
	}
	if t.handles.WorkerFence != vk.NullFence {
		vk.DestroyFence(t.handles.Device, t.handles.WorkerFence, nil)
	}
}

func (t *GpuTranslator) ResetFrameState(frame uint64) {
	t.currentGuestFrame = frame
	t.activeSurface = nil
	t.lastColorRtAddress = 0
	t.activePass = vk.NullRenderPass
	t.activePassNoClear = vk.NullRenderPass
	t.activeFramebuffer = vk.NullFramebuffer
	t.activePipeline = vk.NullPipeline
	t.activeVteControl = 0
	t.activeClipControl = 0
	t.activeDynamicState = nil
	t.pendingComputeBarrier = false

	if frame != t.lastProcessedFrame {
		t.surfacesMutex.Lock()
		for _, surface := range t.surfaces {
			surface.FrameUsed = 0
		}
		t.surfacesMutex.Unlock()
		t.lastProcessedFrame = frame
	}
}

// Translate translates Liverpool draw/compute commands into Vulkan commands and returns the command buffer.
func (t *GpuTranslator) Translate(frame uint64, stream *gpu.LiverpoolCommandStream) *vk.CommandBuffer {
	// Update buffers holding user data.
	t.UpdateUserDataBuffers(stream)

	// Reset per-frame descriptor sets.
	t.staticDescriptorSetIdx = 0

	// Begin recording.
	t.commandBuffer = t.handles.AllocateCommandBuffer()
	vk.BeginCommandBuffer(t.commandBuffer, &vk.CommandBufferBeginInfo{
		SType: vk.StructureTypeCommandBufferBeginInfo,
		Flags: vk.CommandBufferUsageFlags(vk.CommandBufferUsageOneTimeSubmitBit),
	})
	if frame == 0 {
		t.createDummyTexture()
	}

	// Process command stream.
	logger.Printf("[%s] processing %s commands in stream.\n",
		color.Blue.Sprintf("Frame %d", frame),
		color.Blue.Sprint(len(stream.Commands)),
	)
	for _, command := range stream.Commands {
		switch command.Type {
		case gpu.LiverpoolCommandTypeDraw:
			t.Draw(frame, &stream.Draws[command.Index])
		case gpu.LiverpoolCommandTypeDispatch:
			t.Dispatch(frame, &stream.Dispatches[command.Index])
		case gpu.LiverpoolCommandTypeDmaCopy:
			t.DmaCopy(frame, &stream.DmaCopies[command.Index])
		case gpu.LiverpoolCommandTypeBindPipeline:
			t.BindPipeline(frame, &stream.Pipelines[command.Index])
		case gpu.LiverpoolCommandTypeSetDynamicState:
			t.SetDynamicState(&stream.DynamicStates[command.Index])
		}
	}

	// Finish.
	t.EndRenderPass()
	vk.EndCommandBuffer(t.commandBuffer)

	return &t.commandBuffer
}

func (t *GpuTranslator) WaitOnFence() {
	t.fenceMutex.Lock()
	fenceChan := t.fenceChan
	t.fenceMutex.Unlock()
	<-fenceChan
}

func (t *GpuTranslator) SignalFence() {
	t.fenceMutex.Lock()
	defer t.fenceMutex.Unlock()
	select {
	case <-t.fenceChan:
	default:
		close(t.fenceChan)
	}
}

func (t *GpuTranslator) ResetFence() {
	t.fenceMutex.Lock()
	defer t.fenceMutex.Unlock()
	select {
	case <-t.fenceChan:
		t.fenceChan = make(chan struct{})
	default:
	}
}

func (t *GpuTranslator) GetWorkerFence() vk.Fence {
	return t.handles.WorkerFence
}

func (t *GpuTranslator) WaitOnWorkerFence() {
	vk.WaitForFences(t.handles.Device, 1, []vk.Fence{t.handles.WorkerFence}, vk.True, ^uint64(0))
}

func (t *GpuTranslator) ResetWorkerFence() {
	t.handles.QueueMutex.Lock()
	defer t.handles.QueueMutex.Unlock()
	status := vk.GetFenceStatus(t.handles.Device, t.handles.WorkerFence)
	if status == vk.Success {
		vk.ResetFences(t.handles.Device, 1, []vk.Fence{t.handles.WorkerFence})
	}
}

func (t *GpuTranslator) GetBufferAddress(buffer vk.Buffer) uint64 {
	return vulkan.GetBufferDeviceAddress(t.handles.Instance, t.handles.Device, buffer)
}

func (t *GpuTranslator) GetLinearBuffer(address uintptr) (vk.Buffer, uintptr, error) {
	if address >= GlobalGpuAllocator.Base && address < GlobalGpuAllocator.Base+uintptr(GlobalGpuAllocator.Size) {
		return GlobalGpuAllocator.Buffer, address - GlobalGpuAllocator.Base, nil
	}
	if address >= GlobalAllocator.Base && address < GlobalAllocator.Base+uintptr(GlobalAllocator.Size) {
		return GlobalAllocator.Buffer, address - GlobalAllocator.Base, nil
	}
	return vk.NullBuffer, 0, fmt.Errorf("address 0x%X not in any known allocator", address)
}

func (t *GpuTranslator) GetSurfaceByAddress(address uintptr) *vulkan.VulkanSurface {
	t.surfacesMutex.Lock()
	defer t.surfacesMutex.Unlock()
	return t.surfaces[address]
}

func regionsOverlap(addressA, sizeA, addressB, sizeB uintptr) bool {
	if addressA == 0 || addressB == 0 || sizeA == 0 || sizeB == 0 {
		return false
	}
	endA := addressA + sizeA
	endB := addressB + sizeB

	return addressA < endB && addressB < endA
}

func (t *GpuTranslator) CollectGpuResourcesInRange(address, size uintptr) []*vulkan.VulkanImage {
	seen := map[uintptr]struct{}{}
	var images []*vulkan.VulkanImage
	for _, image := range t.images {
		if !regionsOverlap(address, size, image.Address, image.GuestSize) {
			continue
		}
		if _, ok := seen[image.Address]; ok {
			continue
		}
		seen[image.Address] = struct{}{}
		images = append(images, image)
	}

	return images
}

func (t *GpuTranslator) DownloadRegionVkImages(address, size uintptr) error {
	for _, image := range t.CollectGpuResourcesInRange(address, size) {
		if err := image.DownloadFromVkImage(&t.handles); err != nil {
			return err
		}
		recordSyncDownload()
	}

	return nil
}

func (t *GpuTranslator) UploadRegionVkImages(address, size uintptr) error {
	for _, image := range t.CollectGpuResourcesInRange(address, size) {
		if err := image.UploadToVkImage(&t.handles, t.GetLinearBuffer); err != nil {
			return err
		}
	}

	return nil
}
