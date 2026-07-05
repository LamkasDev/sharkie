package translation

import (
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/LamkasDev/sharkie/cmd/structs"
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
			t.registerImage(newImage)
			t.imagesMutex.Unlock()

			t.RefreshImageFromGuest(newImage)
			return newImage, nil, true
		}

		// Upgrade texture-only VkImages to surfaces (adds COLOR_ATTACHMENT usage).
		if !image.IsSurface && isSurface {
			gen := image.Generation
			syncFlags := image.SyncFlags
			syncMarkedFrame := image.SyncMarkedFrame
			mirrorSynced := image.MirrorSynced
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
			newImage.SyncMarkedFrame = syncMarkedFrame
			newImage.MirrorSynced = mirrorSynced
			t.imagesMutex.Lock()
			t.registerImage(newImage)
			t.imagesMutex.Unlock()
			err = image.CopyToImage(&t.handles, newImage)
			if err != nil {
				return nil, err, false
			}

			return newImage, nil, true
		}

		t.RefreshImageFromGuest(image)
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
	t.registerImage(image)
	t.imagesMutex.Unlock()

	t.RefreshImageFromGuest(image)
	return image, nil, true
}

func (t *GpuTranslator) RefreshImageFromGuest(image *vulkan.VulkanImage) {
	if !image.ShouldUploadToVkImage() {
		if image.HasSync(vulkan.ImageSyncGpuModified) {
			image.SetSync(vulkan.ImageSyncGuestUploaded)
			image.ClearSync(vulkan.ImageSyncCpuDirty)
		}
		return
	}
	if err := image.UploadToVkImage(&t.handles, t.GetBufferFromAddress); err != nil {
		return
	}
	image.SetSync(vulkan.ImageSyncGuestUploaded)
	image.ClearSync(vulkan.ImageSyncCpuDirty)
	structs.ClearRegionDirty(image.FirstDescriptor.BaseAddress, vulkan.DescriptorRegionSize(image.FirstDescriptor))
}

func (t *GpuTranslator) MarkGpuModified(image *vulkan.VulkanImage) {
	image.SetSync(vulkan.ImageSyncNeedsReadBarrier)
	image.MirrorSynced = false

	if image.SyncMarkedFrame == t.currentGuestFrame {
		return
	}
	image.SyncMarkedFrame = t.currentGuestFrame

	regionSize := image.GuestSize
	size := regionSize
	if size == 0 {
		image.SetSync(vulkan.ImageSyncGuestUploaded)
		image.ClearSync(vulkan.ImageSyncCpuDirty)
		structs.ClearRegionDirty(image.Address, vulkan.DescriptorRegionSize(image.FirstDescriptor))
		size = structs.SystemPageSize
	}
	t.imagesMutex.Lock()
	t.markGpuModifiedRegion(image.Address, size)
	t.imagesMutex.Unlock()
	structs.GlobalMemoryManager.MarkRegionGpuModified(image.Address, size)
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
