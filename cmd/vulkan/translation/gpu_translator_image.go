package translation

import (
	"github.com/LamkasDev/sharkie/cmd/logger"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan"
	vk "github.com/goki/vulkan"
)

func (t *GpuTranslator) GetImage(descriptor spirvStructs.ImageDescriptor, format vk.Format, isSurface bool) (*vulkan.VulkanImage, error, bool) {
	t.imagesMutex.Lock()
	image, ok := t.images[descriptor.BaseAddress]
	t.imagesMutex.Unlock()

	if ok {
		// Check if we need to recreate the existing image.
		recreate, copyOld := image.NeedsRecreate(descriptor, format, isSurface)
		if recreate {
			gen := image.Generation
			t.EvictResourcesAtAddress(descriptor.BaseAddress)

			newImage, err := vulkan.CreateImage(&t.handles, vulkan.VulkanImageRequest{
				Descriptor: descriptor,
				Format:     format,
				IsSurface:  isSurface,
			}, t.commandBuffer)
			if err != nil {
				return nil, err, false
			}
			newImage.Generation = gen + 1
			t.registerImage(newImage, false)

			if copyOld {
				_ = image.CopyToImage(&t.handles, newImage, t.currentGuestFrame)
			}
			if newImage.ShouldUploadToVkImage(t.currentGuestFrame) {
				if err = newImage.UploadToVkImage(&t.handles, t.GetLinearBuffer, t.currentGuestFrame); err != nil {
					logger.Printf("failed to upload image: %v\n", err)
				}
			}
			return newImage, nil, true
		}

		if isSurface {
			image.IsSurface = true
		}
		if image.ShouldUploadToVkImage(t.currentGuestFrame) {
			if err := image.UploadToVkImage(&t.handles, t.GetLinearBuffer, t.currentGuestFrame); err != nil {
				logger.Printf("failed to upload image: %v\n", err)
			}
		}
		return image, nil, false
	}

	image, err := vulkan.CreateImage(&t.handles, vulkan.VulkanImageRequest{
		Descriptor: descriptor,
		Format:     format,
		IsSurface:  isSurface,
	}, t.commandBuffer)
	if err != nil {
		return nil, err, false
	}
	t.registerImage(image, false)

	if image.ShouldUploadToVkImage(t.currentGuestFrame) {
		_ = image.UploadToVkImage(&t.handles, t.GetLinearBuffer, t.currentGuestFrame)
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
		t.deferDestroySurface(surface)
	}
	t.surfacesMutex.Unlock()

	t.invalidateFramebuffersForAddress(address)

	t.imagesMutex.Lock()
	image, ok := t.images[address]
	t.imagesMutex.Unlock()
	if ok {
		t.unregisterImage(image)
		t.deferDestroyImage(image)
	}

	t.imagesMutex.Lock()
	for hash, view := range t.imageViews {
		if view.Image != nil && view.Image.Address == address {
			delete(t.imageViews, hash)
			t.deferDestroyImageView(view)
		}
	}
	t.imagesMutex.Unlock()
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
