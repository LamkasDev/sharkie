package translation

import (
	"github.com/LamkasDev/sharkie/cmd/logger"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan"
	vk "github.com/goki/vulkan"
)

func (t *GpuTranslator) GetImageByAddress(address uintptr) *vulkan.VulkanImage {
	t.imagesMutex.Lock()
	image, _ := t.images[address]
	t.imagesMutex.Unlock()

	return image
}

func (t *GpuTranslator) GetImage(descriptor spirvStructs.ImageDescriptor, format vk.Format, isSurface bool) (*vulkan.VulkanImage, error, bool) {
	t.imagesMutex.Lock()
	image, ok := t.images[descriptor.BaseAddress]
	t.imagesMutex.Unlock()

	if ok {
		// Check if we need to recreate the existing image.
		recreate := image.NeedsRecreate(descriptor, format, isSurface)
		if recreate {
			gen := image.Generation
			t.EvictResourcesAtAddress(descriptor.BaseAddress)

			newImage, err := vulkan.CreateImage(t.handles, vulkan.VulkanImageRequest{
				Descriptor: descriptor,
				Format:     format,
				IsSurface:  isSurface,
			}, t.commandBuffer, t.currentGuestFrame)
			if err != nil {
				return nil, err, false
			}
			newImage.Generation = gen + 1
			t.registerImage(newImage)

			if newImage.ShouldUploadToVkImage(t.currentGuestFrame) {
				t.EndRenderPass()
				if err = newImage.UploadToVkImage(t.handles, t.commandBuffer, t.GetLinearBuffer, t.currentGuestFrame); err != nil {
					logger.Printf("failed to upload image: %v\n", err)
				}
			}
			return newImage, nil, true
		}

		if isSurface {
			image.IsSurface = true
		}
		if image.ShouldUploadToVkImage(t.currentGuestFrame) {
			t.EndRenderPass()
			if err := image.UploadToVkImage(t.handles, t.commandBuffer, t.GetLinearBuffer, t.currentGuestFrame); err != nil {
				logger.Printf("failed to upload image: %v\n", err)
			}
		}
		return image, nil, false
	}

	image, err := vulkan.CreateImage(t.handles, vulkan.VulkanImageRequest{
		Descriptor: descriptor,
		Format:     format,
		IsSurface:  isSurface,
	}, t.commandBuffer, t.currentGuestFrame)
	if err != nil {
		return nil, err, false
	}
	t.registerImage(image)

	if image.ShouldUploadToVkImage(t.currentGuestFrame) {
		t.EndRenderPass()
		if err = image.UploadToVkImage(t.handles, t.commandBuffer, t.GetLinearBuffer, t.currentGuestFrame); err != nil {
			logger.Printf("failed to upload image: %v\n", err)
		}
	}
	return image, nil, true
}

func (t *GpuTranslator) EvictResourcesAtAddress(address uintptr) {
	t.surfacesMutex.Lock()
	if surface, ok := t.surfaces[address]; ok {
		delete(t.surfaces, address)
		if t.activeSurface != nil && t.activeSurface.Address == address {
			t.activeSurface = nil
		}
		t.handles.DeferDestroySurface(surface)
	}
	t.surfacesMutex.Unlock()
	t.InvalidateFramebuffersForAddress(address)

	t.imagesMutex.Lock()
	image, ok := t.images[address]
	t.imagesMutex.Unlock()
	if ok {
		t.unregisterImage(image)
		t.handles.DeferDestroyImage(image)
	}

	t.imagesMutex.Lock()
	for hash, view := range t.imageViews {
		if view.Image != nil && view.Image.Address == address {
			delete(t.imageViews, hash)
			t.handles.DeferDestroyImageView(view)
		}
	}
	t.imagesMutex.Unlock()
}

func (t *GpuTranslator) InvalidateFramebuffersForAddress(addr uintptr) {
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
		t.handles.DeferDestroyFramebuffer(fb)
	}
}

func (t *GpuTranslator) IsOwnedSurfaceImageView(view *vulkan.VulkanImageView) bool {
	t.surfacesMutex.Lock()
	defer t.surfacesMutex.Unlock()
	for _, surface := range t.surfaces {
		if surface.ImageView == view {
			return true
		}
	}

	return false
}

func (t *GpuTranslator) ClearAllResources() {
	t.imagesMutex.Lock()
	addresses := make([]uintptr, 0, len(t.images))
	for address := range t.images {
		addresses = append(addresses, address)
	}
	t.imagesMutex.Unlock()

	for _, address := range addresses {
		t.EvictResourcesAtAddress(address)
	}
}
