package vulkan

import (
	"fmt"

	"github.com/LamkasDev/cimgui-go-vulkan/imgui"
	vk "github.com/goki/vulkan"
)

// GpuSurface is a Vulkan-side render target that corresponds to a single
// GPU-address-identified framebuffer surface registered by the game.
type GpuSurface struct {
	GPUAddress uintptr
	Width      uint32
	Height     uint32
	Format     vk.Format
	TextureId  imgui.TextureRef

	// Vulkan objects.
	image             vk.Image
	imageMem          vk.DeviceMemory
	imageView         vk.ImageView
	sampler           vk.Sampler
	framebuffer       vk.Framebuffer
	renderPass        vk.RenderPass
	renderPassNoClear vk.RenderPass

	// firstUse tracks whether the image has been transitioned from UNDEFINED.
	firstUse bool

	// frameUsed tracks the last frame this surface was used in.
	frameUsed uint64
}

// Destroy frees all Vulkan resources owned by this surface.
func (s *GpuSurface) Destroy(dev vk.Device) {
	if s.framebuffer != vk.NullFramebuffer {
		vk.DestroyFramebuffer(dev, s.framebuffer, nil)
	}
	if s.renderPass != vk.NullRenderPass {
		vk.DestroyRenderPass(dev, s.renderPass, nil)
	}
	if s.renderPassNoClear != vk.NullRenderPass {
		vk.DestroyRenderPass(dev, s.renderPassNoClear, nil)
	}
	if s.imageView != vk.NullImageView {
		vk.DestroyImageView(dev, s.imageView, nil)
	}
	if s.image != vk.NullImage {
		vk.DestroyImage(dev, s.image, nil)
	}
	if s.imageMem != vk.NullDeviceMemory {
		vk.FreeMemory(dev, s.imageMem, nil)
	}
}

func (t *GpuTranslator) GetSurface(address uintptr, width, height uint32) (imgui.TextureRef, error) {
	// Check if it already exists.
	t.surfacesMutex.Lock()
	surface, ok := t.surfaces[address]
	t.surfacesMutex.Unlock()
	if ok {
		return surface.TextureId, nil
	}

	// Create the surface.
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

	// Transition to ShaderReadOnly so it's valid for sampling even before first draw.
	cb := t.handles.AllocateCommandBuffer(t.pool)
	vk.BeginCommandBuffer(cb, &vk.CommandBufferBeginInfo{
		SType: vk.StructureTypeCommandBufferBeginInfo,
		Flags: vk.CommandBufferUsageFlags(vk.CommandBufferUsageOneTimeSubmitBit),
	})
	t.imageBarrier(cb, surface.image,
		vk.ImageLayoutUndefined, vk.ImageLayoutShaderReadOnlyOptimal,
		0, vk.AccessFlags(vk.AccessShaderReadBit),
		vk.PipelineStageFlags(vk.PipelineStageTopOfPipeBit), vk.PipelineStageFlags(vk.PipelineStageAllGraphicsBit))
	vk.EndCommandBuffer(cb)
	vk.QueueSubmit(t.handles.GraphicsQueue, 1, []vk.SubmitInfo{{
		SType:              vk.StructureTypeSubmitInfo,
		CommandBufferCount: 1,
		PCommandBuffers:    []vk.CommandBuffer{cb},
	}}, vk.NullFence)
	vk.QueueWaitIdle(t.handles.GraphicsQueue)
	vk.FreeCommandBuffers(t.handles.Device, t.pool, 1, []vk.CommandBuffer{cb})

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
