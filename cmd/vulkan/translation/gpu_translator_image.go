package translation

import (
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
		if image.NeedsRecreate(descriptor, format, isSurface) {
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
			t.imagesMutex.Lock()
			t.registerImage(newImage, false)
			t.imagesMutex.Unlock()

			if newImage.ShouldUploadToVkImage() {
				_ = newImage.UploadToVkImage(&t.handles, t.GetLinearBuffer)
			}
			return newImage, nil, true
		}

		// Upgrade texture-only VkImages to surfaces (adds COLOR_ATTACHMENT usage).
		if !image.IsSurface && isSurface {
			gen := image.Generation
			syncFlags := image.SyncFlags

			t.EvictResourcesAtAddress(descriptor.BaseAddress)

			newImage, err := vulkan.CreateImage(&t.handles, vulkan.VulkanImageRequest{
				Descriptor: descriptor,
				Format:     format,
				IsSurface:  true,
			}, t.commandBuffer)
			if err != nil {
				return nil, err, false
			}
			newImage.Generation = gen + 1
			newImage.SyncFlags = syncFlags

			t.imagesMutex.Lock()
			t.registerImage(newImage, true)
			t.imagesMutex.Unlock()
			err = image.CopyToImage(&t.handles, newImage)
			if err != nil {
				return nil, err, false
			}

			return newImage, nil, true
		}

		if image.ShouldUploadToVkImage() {
			_ = image.UploadToVkImage(&t.handles, t.GetLinearBuffer)
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
	t.imagesMutex.Lock()
	t.registerImage(image, false)
	t.imagesMutex.Unlock()

	if image.ShouldUploadToVkImage() {
		_ = image.UploadToVkImage(&t.handles, t.GetLinearBuffer)
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
	if image, ok := t.images[address]; ok {
		t.unregisterImage(address)
		t.deferDestroyImage(image)
	}

	for hash, view := range t.imageViews {
		if view.Image != nil && view.Image.Address == address {
			delete(t.imageViews, hash)
			t.deferDestroyImageView(view)
		}
	}
	t.imagesMutex.Unlock()
}
