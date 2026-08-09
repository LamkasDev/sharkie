package translation

import (
	"context"
	"fmt"
	"runtime"
	"runtime/pprof"
	"sync"
	"unsafe"

	"github.com/LamkasDev/cimgui-go-vulkan/backend"
	glfwvulkanbackend "github.com/LamkasDev/cimgui-go-vulkan/backend/glfwvulkan-backend"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gpu"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/irq"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/spirv"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan"
	vk "github.com/goki/vulkan"
	"github.com/gookit/color"
)

// GpuTranslator converts decoded DrawCalls into Vulkan commands.
type GpuTranslator struct {
	handles *vulkan.VulkanHandles
	backend backend.Backend[glfwvulkanbackend.GLFWWindowFlags]

	// Vulkan surfaces mirroring guest framebuffers.
	surfacesMutex sync.Mutex
	surfaces      map[uintptr]*vulkan.VulkanSurface

	// Pipeline layout shared across all draws.
	pipelineLayout vk.PipelineLayout

	// Descriptor sets.
	globalDescriptorPool *vulkan.VulkanDescriptorPool2
	imageDescriptorPool  *vulkan.VulkanDescriptorPool2

	// Recompiled SPIR-V shaders mirroring Liverpool.LoadedShaders.
	shadersMutex sync.Mutex
	shaders      map[SpirvShaderKey]*spirv.SpirvShader

	// VkShaderModules created from SPIR-V shaders.
	shaderModulesMutex       sync.Mutex
	shaderModules            map[SpirvShaderKey]vk.ShaderModule
	rectlistTescShaderModule vk.ShaderModule
	rectlistTeseShaderModule vk.ShaderModule

	// Per-draw compiled pipelines.
	pipelinesMutex sync.Mutex
	pipelines      map[vulkan.GraphicsPipelineKey]vk.Pipeline

	// Per-dispatch compiled compute pipelines.
	computePipelinesMutex sync.Mutex
	computePipelines      map[vulkan.ComputePipelineKey]vk.Pipeline

	// Per-draw framebuffers.
	framebuffersMutex sync.Mutex
	framebuffers      map[vulkan.FramebufferRequest]*vulkan.VulkanFramebuffer

	// Caches for images, image views and samplers.
	imagesMutex sync.Mutex
	images      map[uintptr]*vulkan.VulkanImage
	imageViews  map[uint64]*vulkan.VulkanImageView

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
	pool                  vk.CommandPool
	commandBuffer         *vulkan.VulkanCommandBuffer
	pendingCommandBuffers []*vulkan.VulkanCommandBuffer

	// Fence for signaling kernel.
	fenceChan          chan struct{}
	fenceMutex         sync.Mutex
	OnFinishedRingWork func()

	// Active state for chronological stream processing.
	lastColorRtAddress              uintptr
	activeSurface                   *vulkan.VulkanSurface
	activeDepthSurface              *vulkan.VulkanSurface
	insideRenderPass                bool
	activePipeline                  vk.Pipeline
	activeFragmentShader            *spirv.SpirvShader
	activeGeometryShader            *spirv.SpirvShader
	activeComputeShader             *spirv.SpirvShader
	activeComputeStoreTargets       []*vulkan.VulkanImage
	activeComputeStoreBufferTargets []*vulkan.VulkanImage
	activeVertexShader              *spirv.SpirvShader
	activeVteControl                uint32
	activeClipControl               uint32
	activeDynamicState              *gpu.LiverpoolSetDynamicState

	activeFragmentShaderKey SpirvShaderKey
	activeGeometryShaderKey SpirvShaderKey
	activeComputeShaderKey  SpirvShaderKey
	activeVertexShaderKey   SpirvShaderKey

	// Direct allocations for Option A translation.
	directAllocationsMutex      sync.Mutex
	directAllocations           map[uintptr]DirectAllocation
	addressTranslationBuffer    vk.Buffer
	addressTranslationBufferMem vk.DeviceMemory
	addressTranslationMap       []AddressTranslationEntry

	// lastProcessedFrame tracks which guest frame surface lifetime state belongs to.
	lastProcessedFrame uint64
	currentGuestFrame  uint64
}

// NewGpuTranslator creates a GpuTranslator, loads stub shaders and builds the stub pipeline layout.
func NewGpuTranslator(handles *vulkan.VulkanHandles, bknd backend.Backend[glfwvulkanbackend.GLFWWindowFlags]) (*GpuTranslator, error) {
	t := &GpuTranslator{
		handles: handles,
		backend: bknd,

		surfacesMutex: sync.Mutex{},
		surfaces:      map[uintptr]*vulkan.VulkanSurface{},

		shadersMutex:       sync.Mutex{},
		shaders:            map[SpirvShaderKey]*spirv.SpirvShader{},
		shaderModulesMutex: sync.Mutex{},
		shaderModules:      map[SpirvShaderKey]vk.ShaderModule{},

		pipelinesMutex: sync.Mutex{},
		pipelines:      map[vulkan.GraphicsPipelineKey]vk.Pipeline{},

		computePipelinesMutex: sync.Mutex{},
		computePipelines:      map[vulkan.ComputePipelineKey]vk.Pipeline{},

		framebuffersMutex: sync.Mutex{},
		framebuffers:      map[vulkan.FramebufferRequest]*vulkan.VulkanFramebuffer{},

		imagesMutex: sync.Mutex{},
		images:      map[uintptr]*vulkan.VulkanImage{},
		imageViews:  map[uint64]*vulkan.VulkanImageView{},

		samplersMutex: sync.Mutex{},
		samplers:      map[uint64]vk.Sampler{},

		userDataBuffersMutex: sync.Mutex{},
		userDataOffsets:      map[uint32]uint32{},

		directAllocationsMutex: sync.Mutex{},
		directAllocations:      map[uintptr]DirectAllocation{},

		pendingCommandBuffers: []*vulkan.VulkanCommandBuffer{},
		fenceChan:             make(chan struct{}),
		fenceMutex:            sync.Mutex{},
	}
	close(t.fenceChan)

	// Allocate user data buffer.
	userDataBuffer, userDataBufferMem, err := vulkan.AllocateBuffer(t.handles, vk.DeviceSize(UserDataBufferSize),
		vk.BufferUsageFlags(vk.BufferUsageShaderDeviceAddressBit|vk.BufferUsageUniformBufferBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
	if err != nil {
		return nil, fmt.Errorf("GpuTranslator: user data buffer: %w", err)
	}
	t.userDataBuffer = userDataBuffer
	t.userDataBufferMem = userDataBufferMem
	t.userDataBufferAddress = t.GetBufferAddress(userDataBuffer)

	// Allocate SSBO for address translation.
	addressTranslationBuffer, addressTranslationBufferMem, err := vulkan.AllocateBuffer(t.handles, vk.DeviceSize(spirvStructs.AddressTranslationCount*32),
		vk.BufferUsageFlags(vk.BufferUsageStorageBufferBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
	if err != nil {
		return nil, fmt.Errorf("GpuTranslator: address translation buffer: %w", err)
	}
	t.addressTranslationBuffer = addressTranslationBuffer
	t.addressTranslationBufferMem = addressTranslationBufferMem

	var addressTranslationData unsafe.Pointer
	result := vk.MapMemory(handles.Device, addressTranslationBufferMem, 0, vk.DeviceSize(spirvStructs.AddressTranslationCount*32), 0, &addressTranslationData)
	if err := vulkan.NewError(result); err != nil {
		return nil, fmt.Errorf("GpuTranslator: map address translation buffer: %w", err)
	}
	t.addressTranslationMap = unsafe.Slice((*AddressTranslationEntry)(addressTranslationData), spirvStructs.AddressTranslationCount)
	t.addressTranslationMap[0].GuestBase = ^uint64(0)

	// Setup hooks for dynamically allocating direct memory.
	HookAllocateMemoryVulkan = func(offset uintptr, length uint64) {
		t.directAllocationsMutex.Lock()
		defer t.directAllocationsMutex.Unlock()

		buffer, mem, err := vulkan.AllocateExternalBuffer(t.handles, vk.DeviceSize(length),
			vk.BufferUsageFlags(vk.BufferUsageShaderDeviceAddressBit|vk.BufferUsageStorageBufferBit|vk.BufferUsageVertexBufferBit|vk.BufferUsageIndexBufferBit|vk.BufferUsageTransferSrcBit|vk.BufferUsageTransferDstBit),
			vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit|vk.MemoryPropertyHostCachedBit))
		if err != nil {
			panic(fmt.Errorf("GpuTranslator: allocate external buffer: %w", err))
		}

		t.directAllocations[offset] = DirectAllocation{
			Buffer:        buffer,
			Memory:        mem,
			DeviceAddress: t.GetBufferAddress(buffer),
			Length:        length,
		}
		t.updateAddressTranslationSSBO()
	}

	HookMapMemoryVulkan = func(addr uintptr, length uint64, offset uintptr) {
		t.directAllocationsMutex.Lock()
		defer t.directAllocationsMutex.Unlock()

		if alloc, ok := t.directAllocations[offset]; ok {
			var err error
			if runtime.GOOS == "windows" {
				err = MapVulkanMemory(addr, length, vulkan.GetMemoryWin32Handle(t.handles.Instance, t.handles.Device, alloc.Memory), 0)
			} else {
				err = MapVulkanMemory(addr, length, uintptr(vulkan.GetMemoryFd(t.handles.Instance, t.handles.Device, alloc.Memory)), 0)
			}
			if err != nil {
				panic(fmt.Errorf("GpuTranslator: map external buffer: %w", err))
			}
			if addr != offset {
				t.directAllocations[addr] = alloc
				delete(t.directAllocations, offset)
			}
		} else {
			fmt.Printf("GpuTranslator: HookMapMemoryVulkan failed to find direct allocation for offset 0x%X\n", offset)
		}
		t.updateAddressTranslationSSBO()
	}

	HookFreeMemoryVulkan = func(offset uintptr) {
		t.directAllocationsMutex.Lock()
		defer t.directAllocationsMutex.Unlock()

		if alloc, ok := t.directAllocations[offset]; ok {
			vk.DestroyBuffer(t.handles.Device, alloc.Buffer, nil)
			vk.FreeMemory(t.handles.Device, alloc.Memory, nil)
			delete(t.directAllocations, offset)
			t.updateAddressTranslationSSBO()
		}
	}

	var globalDescriptorSetLayout, imageDescriptorSetLayout vk.DescriptorSetLayout
	t.pipelineLayout, globalDescriptorSetLayout, imageDescriptorSetLayout, err = vulkan.CreateStubPipelineLayout(t.handles)
	if err != nil {
		return nil, fmt.Errorf("GpuTranslator: pipeline layout: %w", err)
	}

	t.globalDescriptorPool, err = vulkan.CreateDescriptorPool2(t.handles, globalDescriptorSetLayout, []vk.DescriptorPoolSize{
		{
			Type:            vk.DescriptorTypeStorageBuffer,
			DescriptorCount: 256,
		},
	}, 256)
	if err != nil {
		return nil, fmt.Errorf("GpuTranslator: global descriptor pool: %w", err)
	}
	t.globalDescriptorPool.SetCopyTemplate([]vk.CopyDescriptorSet{
		{
			SType:           vk.StructureTypeCopyDescriptorSet,
			SrcBinding:      spirvStructs.GlobalBindingAddressTranslation,
			DstBinding:      spirvStructs.GlobalBindingAddressTranslation,
			DescriptorCount: 1,
		},
	})

	t.imageDescriptorPool, err = vulkan.CreateDescriptorPool2(t.handles, imageDescriptorSetLayout, []vk.DescriptorPoolSize{
		{
			Type:            vk.DescriptorTypeCombinedImageSampler,
			DescriptorCount: 32768,
		},
		{
			Type:            vk.DescriptorTypeStorageImage,
			DescriptorCount: 8192,
		},
	}, 8192)
	if err != nil {
		return nil, fmt.Errorf("GpuTranslator: image descriptor pool: %w", err)
	}
	t.imageDescriptorPool.SetCopyTemplate([]vk.CopyDescriptorSet{
		{
			SType:           vk.StructureTypeCopyDescriptorSet,
			SrcBinding:      spirvStructs.ImageBindingSampledImages1D,
			DstBinding:      spirvStructs.ImageBindingSampledImages1D,
			DescriptorCount: spirvStructs.MaxStaticBindings,
		},
		{
			SType:           vk.StructureTypeCopyDescriptorSet,
			SrcBinding:      spirvStructs.ImageBindingStorageImages1D,
			DstBinding:      spirvStructs.ImageBindingStorageImages1D,
			DescriptorCount: spirvStructs.MaxStaticBindings,
		},
		{
			SType:           vk.StructureTypeCopyDescriptorSet,
			SrcBinding:      spirvStructs.ImageBindingSampledImages2D,
			DstBinding:      spirvStructs.ImageBindingSampledImages2D,
			DescriptorCount: spirvStructs.MaxStaticBindings,
		},
		{
			SType:           vk.StructureTypeCopyDescriptorSet,
			SrcBinding:      spirvStructs.ImageBindingStorageImages2D,
			DstBinding:      spirvStructs.ImageBindingStorageImages2D,
			DescriptorCount: spirvStructs.MaxStaticBindings,
		},
		{
			SType:           vk.StructureTypeCopyDescriptorSet,
			SrcBinding:      spirvStructs.ImageBindingSampledImages3D,
			DstBinding:      spirvStructs.ImageBindingSampledImages3D,
			DescriptorCount: spirvStructs.MaxStaticBindings,
		},
		{
			SType:           vk.StructureTypeCopyDescriptorSet,
			SrcBinding:      spirvStructs.ImageBindingStorageImages3D,
			DstBinding:      spirvStructs.ImageBindingStorageImages3D,
			DescriptorCount: spirvStructs.MaxStaticBindings,
		},
		{
			SType:           vk.StructureTypeCopyDescriptorSet,
			SrcBinding:      spirvStructs.ImageBindingSampledImages2DArray,
			DstBinding:      spirvStructs.ImageBindingSampledImages2DArray,
			DescriptorCount: spirvStructs.MaxStaticBindings,
		},
		{
			SType:           vk.StructureTypeCopyDescriptorSet,
			SrcBinding:      spirvStructs.ImageBindingStorageImages2DArray,
			DstBinding:      spirvStructs.ImageBindingStorageImages2DArray,
			DescriptorCount: spirvStructs.MaxStaticBindings,
		},
	})

	t.activePipeline = vk.NullPipeline

	// Allocate quad list index buffer.
	const maxQuads = 16384
	const quadListIndexCount = maxQuads * 6
	quadListIndexBuffer, quadListIndexBufferMem, err := vulkan.AllocateBuffer(t.handles, vk.DeviceSize(quadListIndexCount*2),
		vk.BufferUsageFlags(vk.BufferUsageIndexBufferBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
	if err != nil {
		return nil, fmt.Errorf("GpuTranslator: allocate quad list index buffer: %w", err)
	}

	// Map and fill quad list index buffer.
	var indexData unsafe.Pointer
	result = vk.MapMemory(handles.Device, quadListIndexBufferMem, 0, vk.DeviceSize(quadListIndexCount*2), 0, &indexData)
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

	go pprof.Do(context.Background(), pprof.Labels("name", "MemorySyncWorker"), func(ctx context.Context) {
		t.memorySyncWorker()
	})

	return t, nil
}

// Destroy frees all Vulkan resources.
func (t *GpuTranslator) Destroy() {
	vk.DeviceWaitIdle(t.handles.Device)
	t.globalDescriptorPool.Destroy(t.handles)
	t.imageDescriptorPool.Destroy(t.handles)
	t.handles.FlushDeferredDestruction()
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
	if t.pool != vk.NullCommandPool {
		vk.DestroyCommandPool(t.handles.Device, t.pool, nil)
	}
	if t.addressTranslationBuffer != vk.NullBuffer {
		vk.DestroyBuffer(t.handles.Device, t.addressTranslationBuffer, nil)
		vk.FreeMemory(t.handles.Device, t.addressTranslationBufferMem, nil)
	}
	t.directAllocationsMutex.Lock()
	for _, alloc := range t.directAllocations {
		vk.DestroyBuffer(t.handles.Device, alloc.Buffer, nil)
		vk.FreeMemory(t.handles.Device, alloc.Memory, nil)
	}
	t.directAllocationsMutex.Unlock()
	t.handles.FencePool.Destroy(t.handles)
}

func (t *GpuTranslator) ResetFrameState(frame uint64) {
	t.currentGuestFrame = frame
	t.activeSurface = nil
	t.activeDepthSurface = nil
	t.lastColorRtAddress = 0
	t.insideRenderPass = false
	t.activePipeline = vk.NullPipeline
	t.activeVteControl = 0
	t.activeClipControl = 0
	t.activeDynamicState = nil
	t.activeVertexShader = nil
	t.activeFragmentShader = nil
	t.activeGeometryShader = nil
	t.activeComputeShader = nil

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
func (t *GpuTranslator) Translate(frame uint64, stream *gpu.LiverpoolCommandStream) {
	remaining := len(stream.Commands) - stream.CommandIndex
	if remaining == 0 {
		return
	}

	// Process command stream.
	if logger.LogRenderer {
		logger.Printf("[%s] processing %s remaining commands in %s stream.\n",
			color.Blue.Sprintf("Frame %d", frame),
			color.Green.Sprint(remaining),
			color.Blue.Sprint(stream.Name),
		)
	}
	for stream.CommandIndex < len(stream.Commands) {
		command := stream.Commands[stream.CommandIndex]
		switch command.Type {
		case gpu.LiverpoolCommandTypeDraw:
			t.Draw(frame, &stream.Draws[command.Index])
		case gpu.LiverpoolCommandTypeDispatch:
			t.Dispatch(frame, &stream.Dispatches[command.Index])
		case gpu.LiverpoolCommandTypeDmaCopy:
			t.DmaCopy(frame, &stream.DmaCopies[command.Index])
		case gpu.LiverpoolCommandTypeBindPipeline:
			t.BindPipeline(frame, &stream.Pipelines[command.Index])
		case gpu.LiverpoolCommandTypeBindResources:
			t.BindResources(frame, &stream.BindResources[command.Index])
		case gpu.LiverpoolCommandTypeBindComputePipeline:
			t.BindComputePipeline(frame, &stream.ComputePipelines[command.Index])
		case gpu.LiverpoolCommandTypeSetDynamicState:
			t.SetDynamicState(&stream.DynamicStates[command.Index])
		case gpu.LiverpoolCommandTypeWriteData:
			t.WriteData(&stream.WriteDatas[command.Index])
		case gpu.LiverpoolCommandTypeWaitRegMemory:
			t.WaitRegMemory(&stream.WaitRegMems[command.Index])
		case gpu.LiverpoolCommandTypeFlip:
			irq.GlobalInterruptHandler.Signal(irq.InterruptIdGraphicsFlip)
		}
		stream.CommandIndex++
	}
}

func (t *GpuTranslator) BeforeTranslate() {
	if t.currentGuestFrame == 0 {
		t.createDummyTextures()
	}
	t.globalDescriptorPool.Reset(t.currentGuestFrame)
	t.imageDescriptorPool.Reset(t.currentGuestFrame)
	vulkan.ResetDetilePipelines(t.currentGuestFrame)
}

func (t *GpuTranslator) StartCommandBuffer() {
	var err error
	t.commandBuffer, err = vulkan.CreateCommandBuffer(t.handles)
	if err != nil {
		panic(err)
	}
	vk.BeginCommandBuffer(t.commandBuffer.CommandBuffer, &vk.CommandBufferBeginInfo{
		SType: vk.StructureTypeCommandBufferBeginInfo,
		Flags: vk.CommandBufferUsageFlags(vk.CommandBufferUsageOneTimeSubmitBit),
	})
}

func (t *GpuTranslator) EndRenderPass() {
	if t.insideRenderPass {
		vulkan.CmdEndRendering(t.handles.Instance, t.commandBuffer.CommandBuffer)
		t.insideRenderPass = false
		t.activePipeline = vk.NullPipeline
	}
}

func (t *GpuTranslator) EndCommandBuffer() {
	if t.commandBuffer == nil {
		return
	}
	t.EndRenderPass()
	t.commandBuffer.End(t.handles)
	t.pendingCommandBuffers = append(t.pendingCommandBuffers, t.commandBuffer)
	t.commandBuffer = nil
}

func (t *GpuTranslator) FlushCommandBuffers() bool {
	for _, commandBuffer := range t.pendingCommandBuffers {
		if commandBuffer.Submitted {
			continue
		}
		if !commandBuffer.CanSubmit(t.currentGuestFrame) {
			return false
		}
		for _, write := range commandBuffer.Writes {
			dstSlice := unsafe.Slice((*uint32)(unsafe.Pointer(write.Address)), len(write.Data))
			copy(dstSlice, write.Data)
			if logger.LogRendererInternal {
				logger.Printf("[%s] wrote %s bytes to %s.\n",
					color.Blue.Sprintf("Frame %d", t.currentGuestFrame),
					color.Green.Sprintf("%+v", write.Data),
					color.Yellow.Sprintf("0x%X", write.Address),
				)
			}
			if write.Eop {
				irq.GlobalInterruptHandler.Signal(irq.InterruptIdGraphicsEop)
			}
		}
		fence, err := t.handles.FencePool.Get(t.handles, t.currentGuestFrame)
		if err != nil {
			panic(err)
		}
		err = vulkan.NewError(t.handles.GraphicsQueue.Submit([]vk.SubmitInfo{{
			SType:              vk.StructureTypeSubmitInfo,
			CommandBufferCount: 1,
			PCommandBuffers:    []vk.CommandBuffer{commandBuffer.CommandBuffer},
		}}, fence))
		if err != nil {
			panic(err)
		}
		vk.WaitForFences(t.handles.Device, 1, []vk.Fence{fence}, vk.True, vk.MaxUint64)
		t.handles.FencePool.Put(t.handles, fence, t.currentGuestFrame)
		commandBuffer.Destroy(t.handles)
		commandBuffer.Submitted = true
	}

	return true
}

func (t *GpuTranslator) SubmitCommandBuffers() {
	for !t.FlushCommandBuffers() {
		runtime.Gosched()
	}
	t.pendingCommandBuffers = []*vulkan.VulkanCommandBuffer{}
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

func (t *GpuTranslator) CollectGpuResourcesInRange(address, size uintptr) []*vulkan.VulkanImage {
	var images []*vulkan.VulkanImage
	end := address + size
	seen := map[uintptr]struct{}{}

	structs.GlobalMemoryManager.Lock.Lock()
	for addr := address >> SystemPageShift; (addr << SystemPageShift) < end; addr++ {
		if page, ok := structs.GlobalMemoryManager.Pages[addr]; ok {
			for _, resource := range page.Resources {
				if resource == nil {
					continue
				}
				image := resource.(*vulkan.VulkanImage)
				if _, ok := seen[image.Address]; !ok {
					images = append(images, image)
					seen[image.Address] = struct{}{}
				}
			}
		}
	}
	structs.GlobalMemoryManager.Lock.Unlock()

	return images
}

func (t *GpuTranslator) DownloadRegionVkImages(address, size uintptr, commandBuffer *vulkan.VulkanCommandBuffer) error {
	for _, image := range t.CollectGpuResourcesInRange(address, size) {
		if !image.ShouldDownloadFromVkImage() {
			continue
		}
		err := image.DownloadFromVkImage(t.handles, commandBuffer, t.GetLinearBuffer, t.currentGuestFrame)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *GpuTranslator) UploadRegionVkImages(address, size uintptr, commandBuffer *vulkan.VulkanCommandBuffer) error {
	for _, image := range t.CollectGpuResourcesInRange(address, size) {
		if !image.ShouldUploadToVkImage(t.currentGuestFrame) {
			continue
		}
		if err := image.UploadToVkImage(t.handles, commandBuffer, t.GetLinearBuffer, t.currentGuestFrame); err != nil {
			return err
		}
	}

	return nil
}
