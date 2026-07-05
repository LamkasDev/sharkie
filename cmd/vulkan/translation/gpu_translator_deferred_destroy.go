package translation

import (
	"github.com/LamkasDev/sharkie/cmd/vulkan"
	vk "github.com/goki/vulkan"
)

type deferredDestroyQueue struct {
	surfaces     []*vulkan.VulkanSurface
	images       []*vulkan.VulkanImage
	imageViews   []*vulkan.VulkanImageView
	framebuffers []*vulkan.VulkanFramebuffer
}

// deferDestroySurface queues a surface for destruction after the current frame's GPU work completes.
func (t *GpuTranslator) deferDestroySurface(surface *vulkan.VulkanSurface) {
	if surface == nil {
		return
	}
	t.deferredDestroyMutex.Lock()
	t.deferredDestroy.surfaces = append(t.deferredDestroy.surfaces, surface)
	t.deferredDestroyMutex.Unlock()
}

func (t *GpuTranslator) deferDestroyImage(image *vulkan.VulkanImage) {
	if image == nil || image.Image == vk.NullImage {
		return
	}
	t.deferredDestroyMutex.Lock()
	t.deferredDestroy.images = append(t.deferredDestroy.images, image)
	t.deferredDestroyMutex.Unlock()
}

func (t *GpuTranslator) deferDestroyImageView(view *vulkan.VulkanImageView) {
	if view == nil || view.ImageView == vk.NullImageView {
		return
	}
	t.deferredDestroyMutex.Lock()
	t.deferredDestroy.imageViews = append(t.deferredDestroy.imageViews, view)
	t.deferredDestroyMutex.Unlock()
}

func (t *GpuTranslator) deferDestroyFramebuffer(fb *vulkan.VulkanFramebuffer) {
	if fb == nil {
		return
	}
	t.deferredDestroyMutex.Lock()
	t.deferredDestroy.framebuffers = append(t.deferredDestroy.framebuffers, fb)
	t.deferredDestroyMutex.Unlock()
}

func (t *GpuTranslator) invalidateFramebuffersForAddress(addr uintptr) {
	t.framebuffersMutex.Lock()
	var stale []*vulkan.VulkanFramebuffer
	for req, fb := range t.framebuffers {
		if req.GpuAddress != addr && req.DepthGpuAddress != addr {
			continue
		}
		delete(t.framebuffers, req)
		stale = append(stale, fb)
	}
	t.framebuffersMutex.Unlock()

	for _, fb := range stale {
		t.deferDestroyFramebuffer(fb)
	}
}

// FlushDeferredDestruction destroys resources queued during frame recording.
// Call only after the frame command buffer has finished executing on the GPU.
func (t *GpuTranslator) FlushDeferredDestruction() {
	t.deferredDestroyMutex.Lock()
	batch := t.deferredDestroy
	t.deferredDestroy = deferredDestroyQueue{}
	t.deferredDestroyMutex.Unlock()

	device := t.handles.Device
	for _, fb := range batch.framebuffers {
		fb.Destroy(device)
	}
	for _, view := range batch.imageViews {
		view.Destroy(device)
	}
	for _, image := range batch.images {
		image.Destroy(device)
	}
	for _, surface := range batch.surfaces {
		surface.Destroy(device)
	}
}

func (t *GpuTranslator) isOwnedSurfaceImageView(view *vulkan.VulkanImageView) bool {
	t.surfacesMutex.Lock()
	defer t.surfacesMutex.Unlock()
	for _, surface := range t.surfaces {
		if surface.ImageView == view {
			return true
		}
	}
	return false
}
