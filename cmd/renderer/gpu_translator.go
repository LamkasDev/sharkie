package renderer

import (
	"fmt"
	"math"
	"sync"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/structs/gpu"
	"github.com/elokore/cimgui-go-vulkan/backend"
	glfwvulkanbackend "github.com/elokore/cimgui-go-vulkan/backend/glfwvulkan-backend"
	"github.com/elokore/cimgui-go-vulkan/imgui"
	vk "github.com/vulkan-go/vulkan"
)

// GpuTranslator converts decoded DrawCalls into Vulkan commands.
type GpuTranslator struct {
	handles  VulkanHandles
	backend  backend.Backend[glfwvulkanbackend.GLFWWindowFlags]
	mutex    sync.Mutex
	surfaces map[uintptr]*GpuSurface

	// Stub pipeline shared across all draws until real shaders are available.
	stubPipelineLayout vk.PipelineLayout
	stubPipeline       vk.Pipeline
	stubVertShader     vk.ShaderModule
	stubFragShader     vk.ShaderModule

	// Command pool/buffer for this frame's GPU work.
	pool vk.CommandPool
}

// NewGpuTranslator creates a GpuTranslator, loads stub shaders and builds the stub pipeline layout.
func NewGpuTranslator(handles VulkanHandles, bknd backend.Backend[glfwvulkanbackend.GLFWWindowFlags]) (*GpuTranslator, error) {
	t := &GpuTranslator{
		handles:  handles,
		backend:  bknd,
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
	t.mutex.Lock()
	vk.DeviceWaitIdle(device)
	for _, s := range t.surfaces {
		s.Destroy(device)
	}
	t.mutex.Unlock()
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
func (t *GpuTranslator) RegisterSurface(address uintptr, width, height uint32) (imgui.TextureRef, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if surface, exists := t.surfaces[address]; exists {
		return surface.TextureId, nil
	}

	surface := &GpuSurface{
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
	t.surfaces[address] = surface

	return surface.TextureId, nil
}

// Translate translates a slice of DrawCalls into Vulkan commands and returns the command buffer.
func (t *GpuTranslator) Translate(draws []gpu.LiverpoolDrawCall) *vk.CommandBuffer {
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

func (t *GpuTranslator) recordDraw(commandBuffer vk.CommandBuffer, draw *gpu.LiverpoolDrawCall) {
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
	if !surface.firstUse {
		t.imageBarrier(commandBuffer, surface.image,
			vk.ImageLayoutShaderReadOnlyOptimal, vk.ImageLayoutColorAttachmentOptimal,
			vk.AccessFlags(vk.AccessShaderReadBit), vk.AccessFlags(vk.AccessColorAttachmentWriteBit),
			vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit),
			vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit),
		)
	} else {
		surface.firstUse = false
	}

	// Derive clear color from the stub.
	clearColor := vk.ClearValue{}
	clearColor.SetColor([]float32{0.8, 0.0, 0.8, 1.0})
	vk.CmdBeginRenderPass(commandBuffer, &vk.RenderPassBeginInfo{
		SType:           vk.StructureTypeRenderPassBeginInfo,
		RenderPass:      surface.renderPass,
		Framebuffer:     surface.framebuffer,
		RenderArea:      vk.Rect2D{Extent: vk.Extent2D{Width: surface.Width, Height: surface.Height}},
		ClearValueCount: 1,
		PClearValues:    []vk.ClearValue{clearColor},
	}, vk.SubpassContentsInline)

	vk.CmdBindPipeline(commandBuffer, vk.PipelineBindPointGraphics, t.stubPipeline)
	t.setDynamicState(commandBuffer, draw, surface)

	// Push a color constant.
	stubColor := [4]float32{0.8, 0.0, 0.8, 1.0}
	vk.CmdPushConstants(
		commandBuffer, t.stubPipelineLayout,
		vk.ShaderStageFlags(vk.ShaderStageFragmentBit),
		0, 16,
		unsafe.Pointer(&stubColor[0]),
	)

	// Draw a full-screen triangle; vertex count 3, instanced as needed.
	instanceCount := draw.InstanceCount
	if instanceCount == 0 {
		instanceCount = 1
	}
	vk.CmdDraw(commandBuffer, 3, instanceCount, 0, 0)

	vk.CmdEndRenderPass(commandBuffer)
}

func (t *GpuTranslator) setDynamicState(commandBuffer vk.CommandBuffer, draw *gpu.LiverpoolDrawCall, surface *GpuSurface) {
	// Derive viewport from GCN scale/offset registers.
	// XScale = width/2, YScale = -height/2 (GCN NDC => screen, Y is flipped).
	vpWidth := float32(math.Abs(float64(draw.VpXScale)) * 2)
	vpHeight := float32(math.Abs(float64(draw.VpYScale)) * 2)
	vpX := draw.VpXOffset - vpWidth/2
	vpY := draw.VpYOffset - vpHeight/2

	// Negative height = Vulkan's built-in Y-flip (VK_KHR_maintenance1).
	if draw.VpYScale < 0 {
		vpY = draw.VpYOffset + vpHeight/2
		vpHeight = -vpHeight
	}
	if vpWidth <= 0 || vpHeight == 0 {
		vpWidth = float32(surface.Width)
		vpHeight = float32(surface.Height)
		vpX, vpY = 0, 0
	}
	vk.CmdSetViewport(commandBuffer, 0, 1, []vk.Viewport{{
		X:        vpX,
		Y:        vpY,
		Width:    vpWidth,
		Height:   vpHeight,
		MinDepth: 0.0,
		MaxDepth: 1.0,
	}})

	sx, sy, sw, sh := draw.ScissorRect()
	if sw <= 0 || sh <= 0 {
		sw = int(surface.Width)
		sh = int(surface.Height)
		sx, sy = 0, 0
	}
	vk.CmdSetScissor(commandBuffer, 0, 1, []vk.Rect2D{{
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
