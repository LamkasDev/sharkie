package renderer

import (
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
	image       vk.Image
	imageMem    vk.DeviceMemory
	imageView   vk.ImageView
	sampler     vk.Sampler
	framebuffer vk.Framebuffer
	renderPass  vk.RenderPass

	// firstUse tracks whether the image has been transitioned from UNDEFINED.
	firstUse bool
}

// Destroy frees all Vulkan resources owned by this surface.
func (s *GpuSurface) Destroy(dev vk.Device) {
	if s.framebuffer != vk.NullFramebuffer {
		vk.DestroyFramebuffer(dev, s.framebuffer, nil)
	}
	if s.renderPass != vk.NullRenderPass {
		vk.DestroyRenderPass(dev, s.renderPass, nil)
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
