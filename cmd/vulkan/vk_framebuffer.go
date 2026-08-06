package vulkan

import (
	vk "github.com/goki/vulkan"
)

type VulkanFramebuffer struct {
	ColorView   vk.ImageView
	DepthView   vk.ImageView
	ColorFormat vk.Format
	DepthFormat vk.Format
	Width       uint32
	Height      uint32
}

type FramebufferRequest struct {
	ImageView      *VulkanImageView
	DepthImageView *VulkanImageView

	FramebufferKey
}

type FramebufferKey struct {
	GpuAddress      uintptr
	DepthGpuAddress uintptr

	Format      vk.Format
	DepthFormat vk.Format
	Width       uint32
	Height      uint32
}

func CreateFramebuffer(handles *VulkanHandles, request FramebufferRequest) (*VulkanFramebuffer, error) {
	fb := &VulkanFramebuffer{
		ColorFormat: request.Format,
		DepthFormat: request.DepthFormat,
		Width:       request.Width,
		Height:      request.Height,
	}
	if request.Format != vk.FormatUndefined && request.ImageView != nil {
		fb.ColorView = request.ImageView.ImageView
	}
	if request.DepthFormat != vk.FormatUndefined && request.DepthImageView != nil {
		fb.DepthView = request.DepthImageView.ImageView
	}

	return fb, nil
}
