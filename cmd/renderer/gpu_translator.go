package renderer

import (
	"fmt"
	"math"
	"sync"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/structs/gpu"
	vk "github.com/vulkan-go/vulkan"
)

// GpuTranslator converts decoded DrawCalls into Vulkan commands.
type GpuTranslator struct {
	handles  VulkanHandles
	mutex    sync.Mutex
	surfaces map[uintptr]*GpuSurface

	// Stub pipeline shared across all draws until real shaders are available.
	stubPipelineLayout vk.PipelineLayout
	stubPipeline       vk.Pipeline
	stubVertShader     vk.ShaderModule
	stubFragShader     vk.ShaderModule

	// Command pool/buffer for this frame's GPU work.
	pool          vk.CommandPool
	commandBuffer vk.CommandBuffer

	// LastRenderedSurface is the last surface a draw was recorded into.
	LastRenderedSurface *GpuSurface
}

// NewGpuTranslator creates a GpuTranslator, loads stub shaders and builds the stub pipeline layout.
func NewGpuTranslator(handles VulkanHandles) (*GpuTranslator, error) {
	t := &GpuTranslator{
		handles:  handles,
		surfaces: make(map[uintptr]*GpuSurface),
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
	device := t.handles.Device
	vk.DeviceWaitIdle(device)
	for _, s := range t.surfaces {
		s.Destroy(device)
	}
	if t.stubPipeline != vk.NullPipeline {
		vk.DestroyPipeline(device, t.stubPipeline, nil)
	}
	if t.stubPipelineLayout != vk.NullPipelineLayout {
		vk.DestroyPipelineLayout(device, t.stubPipelineLayout, nil)
	}
	if t.stubVertShader != vk.NullShaderModule {
		vk.DestroyShaderModule(device, t.stubVertShader, nil)
	}
	if t.stubFragShader != vk.NullShaderModule {
		vk.DestroyShaderModule(device, t.stubFragShader, nil)
	}
	if t.pool != vk.NullCommandPool {
		vk.DestroyCommandPool(device, t.pool, nil)
	}
}

// RegisterSurface registers a GPU address as a Vulkan render target.
func (t *GpuTranslator) RegisterSurface(address uintptr, width, height uint32) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if _, exists := t.surfaces[address]; exists {
		return nil
	}

	surface := &GpuSurface{
		GPUAddress: address,
		Width:      width,
		Height:     height,
		Format:     vk.FormatR8g8b8a8Unorm,
		firstUse:   true,
	}
	if err := t.allocSurface(surface); err != nil {
		return fmt.Errorf("RegisterSurface 0x%X: %w", address, err)
	}
	t.surfaces[address] = surface

	return nil
}

// Submit translates a slice of DrawCalls into Vulkan commands and submits them.
func (t *GpuTranslator) Submit(draws []gpu.LiverpoolDrawCall) {
	if len(draws) == 0 {
		return
	}

	// Begin recording.
	t.commandBuffer = t.handles.AllocateCommandBuffer(t.pool)
	vk.BeginCommandBuffer(t.commandBuffer, &vk.CommandBufferBeginInfo{
		SType: vk.StructureTypeCommandBufferBeginInfo,
		Flags: vk.CommandBufferUsageFlags(vk.CommandBufferUsageOneTimeSubmitBit),
	})
	for i := range draws {
		t.recordDraw(&draws[i])
	}
	vk.EndCommandBuffer(t.commandBuffer)

	// Submit.
	vk.QueueSubmit(t.handles.GraphicsQueue, 1, []vk.SubmitInfo{{
		SType:              vk.StructureTypeSubmitInfo,
		CommandBufferCount: 1,
		PCommandBuffers:    []vk.CommandBuffer{t.commandBuffer},
	}}, vk.NullFence)
	vk.QueueWaitIdle(t.handles.GraphicsQueue)

	vk.FreeCommandBuffers(t.handles.Device, t.pool, 1, []vk.CommandBuffer{t.commandBuffer})
	// t.commandBuffer = vk.NullCommandBuffer
}

func (t *GpuTranslator) recordDraw(draw *gpu.LiverpoolDrawCall) {
	rtAddress := draw.RtGpuAddress()
	t.mutex.Lock()
	surface, ok := t.surfaces[rtAddress]
	t.mutex.Unlock()
	if !ok {
		return
	}

	// Lazily build the stub pipeline against this surface's render pass.
	if t.stubPipeline == vk.NullPipeline {
		if err := t.createStubPipeline(surface.renderPass, surface.Width, surface.Height); err != nil {
			return
		}
	}

	// Transition image layout on first use.
	if surface.firstUse {
		t.transitionImage(surface, vk.ImageLayoutUndefined, vk.ImageLayoutColorAttachmentOptimal)
		surface.firstUse = false
	}

	// Derive clear color from the stub.
	clearColor := vk.ClearValue{}
	clearColor.SetColor([]float32{0.8, 0.0, 0.8, 1.0})
	vk.CmdBeginRenderPass(t.commandBuffer, &vk.RenderPassBeginInfo{
		SType:           vk.StructureTypeRenderPassBeginInfo,
		RenderPass:      surface.renderPass,
		Framebuffer:     surface.framebuffer,
		RenderArea:      vk.Rect2D{Extent: vk.Extent2D{Width: surface.Width, Height: surface.Height}},
		ClearValueCount: 1,
		PClearValues:    []vk.ClearValue{clearColor},
	}, vk.SubpassContentsInline)

	vk.CmdBindPipeline(t.commandBuffer, vk.PipelineBindPointGraphics, t.stubPipeline)
	t.setDynamicState(draw)

	// Push a color constant.
	stubColor := [4]float32{0.8, 0.0, 0.8, 1.0}
	vk.CmdPushConstants(
		t.commandBuffer, t.stubPipelineLayout,
		vk.ShaderStageFlags(vk.ShaderStageFragmentBit),
		0, 16,
		unsafe.Pointer(&stubColor[0]),
	)

	// Draw a full-screen triangle; vertex count 3, instanced as needed.
	instanceCount := draw.InstanceCount
	if instanceCount == 0 {
		instanceCount = 1
	}
	vk.CmdDraw(t.commandBuffer, 3, instanceCount, 0, 0)

	vk.CmdEndRenderPass(t.commandBuffer)
	t.LastRenderedSurface = surface
}

func (t *GpuTranslator) setDynamicState(draw *gpu.LiverpoolDrawCall) {
	// Viewport: GCN stores XScale and XOffset separately.
	// VpXOffset = width/2, |VpYScale| = height/2.
	vpWidth := float32(math.Abs(float64(draw.VpXScale)) * 2)
	vpHeight := float32(math.Abs(float64(draw.VpYScale)) * 2)
	vpX := draw.VpXOffset - vpWidth/2
	vpY := draw.VpYOffset - vpHeight/2

	// GCN Y-scale is negative (flips Y), which Vulkan handles natively with
	// a negative-height viewport (VK_KHR_maintenance1).
	if draw.VpYScale < 0 {
		vpY = draw.VpYOffset + vpHeight/2
		vpHeight = -vpHeight
	}

	vk.CmdSetViewport(t.commandBuffer, 0, 1, []vk.Viewport{{
		X:        vpX,
		Y:        vpY,
		Width:    vpWidth,
		Height:   vpHeight,
		MinDepth: 0.0,
		MaxDepth: 1.0,
	}})

	sx, sy, sw, sh := draw.ScissorRect()
	vk.CmdSetScissor(t.commandBuffer, 0, 1, []vk.Rect2D{{
		Offset: vk.Offset2D{X: int32(sx), Y: int32(sy)},
		Extent: vk.Extent2D{Width: uint32(sw), Height: uint32(sh)},
	}})
}

// SurfaceImageView returns the VkImageView for a registered surface so the renderer can display it as a texture.
// Returns vk.NullImageView if unknown.
func (t *GpuTranslator) SurfaceImageView(gpuAddress uintptr) vk.ImageView {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if s, ok := t.surfaces[gpuAddress]; ok {
		return s.imageView
	}
	return vk.NullImageView
}
