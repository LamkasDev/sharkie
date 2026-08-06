package vulkan

import (
	vk "github.com/goki/vulkan"
)

type deferredDestroyQueue struct {
	functions      []func()
	surfaces       []*VulkanSurface
	images         []*VulkanImage
	imageViews     []*VulkanImageView
	framebuffers   []*VulkanFramebuffer
	bufferViews    []vk.BufferView
	buffers        []vk.Buffer
	bufferMemories []vk.DeviceMemory
}

func (vkh *VulkanHandles) DeferDestroyFunction(function func()) {
	vkh.DeferredDestroyMutex.Lock()
	vkh.DeferredDestroy.functions = append(vkh.DeferredDestroy.functions, function)
	vkh.DeferredDestroyMutex.Unlock()
}

func (vkh *VulkanHandles) DeferDestroySurface(surface *VulkanSurface) {
	if surface == nil {
		return
	}
	vkh.DeferredDestroyMutex.Lock()
	vkh.DeferredDestroy.surfaces = append(vkh.DeferredDestroy.surfaces, surface)
	vkh.DeferredDestroyMutex.Unlock()
}

func (vkh *VulkanHandles) DeferDestroyImage(image *VulkanImage) {
	if image == nil || image.Image == vk.NullImage {
		return
	}
	vkh.DeferredDestroyMutex.Lock()
	vkh.DeferredDestroy.images = append(vkh.DeferredDestroy.images, image)
	vkh.DeferredDestroyMutex.Unlock()
}

func (vkh *VulkanHandles) DeferDestroyImageView(view *VulkanImageView) {
	if view == nil || view.ImageView == vk.NullImageView {
		return
	}
	vkh.DeferredDestroyMutex.Lock()
	vkh.DeferredDestroy.imageViews = append(vkh.DeferredDestroy.imageViews, view)
	vkh.DeferredDestroyMutex.Unlock()
}

func (vkh *VulkanHandles) DeferDestroyFramebuffer(fb *VulkanFramebuffer) {
	if fb == nil {
		return
	}
	vkh.DeferredDestroyMutex.Lock()
	vkh.DeferredDestroy.framebuffers = append(vkh.DeferredDestroy.framebuffers, fb)
	vkh.DeferredDestroyMutex.Unlock()
}

func (vkh *VulkanHandles) DeferDestroyBufferView(view vk.BufferView) {
	if view == vk.NullBufferView {
		return
	}
	vkh.DeferredDestroyMutex.Lock()
	vkh.DeferredDestroy.bufferViews = append(vkh.DeferredDestroy.bufferViews, view)
	vkh.DeferredDestroyMutex.Unlock()
}

func (vkh *VulkanHandles) DeferDestroyBuffer(buffer vk.Buffer, mem vk.DeviceMemory) {
	if buffer == vk.NullBuffer {
		return
	}
	vkh.DeferredDestroyMutex.Lock()
	vkh.DeferredDestroy.buffers = append(vkh.DeferredDestroy.buffers, buffer)
	vkh.DeferredDestroy.bufferMemories = append(vkh.DeferredDestroy.bufferMemories, mem)
	vkh.DeferredDestroyMutex.Unlock()
}

func (vkh *VulkanHandles) FlushDeferredDestruction() {
	vkh.DeferredDestroyMutex.Lock()
	batch := vkh.DeferredDestroy
	vkh.DeferredDestroy = deferredDestroyQueue{}
	vkh.DeferredDestroyMutex.Unlock()

	device := vkh.Device
	for _, function := range batch.functions {
		function()
	}
	for _, view := range batch.imageViews {
		view.Destroy(device)
	}
	for _, view := range batch.bufferViews {
		vk.DestroyBufferView(device, view, nil)
	}
	for _, buffer := range batch.buffers {
		vk.DestroyBuffer(device, buffer, nil)
	}
	for _, mem := range batch.bufferMemories {
		vk.FreeMemory(device, mem, nil)
	}
	for _, image := range batch.images {
		image.Destroy(device)
	}
	for _, surface := range batch.surfaces {
		surface.Destroy(device)
	}
}
