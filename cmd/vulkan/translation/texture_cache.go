package translation

import (
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan"
)

func (t *GpuTranslator) registerImage(image *vulkan.VulkanImage, isUpgrade bool) {
	t.imagesMutex.Lock()
	t.images[image.Address] = image
	t.imagesMutex.Unlock()
	structs.GlobalMemoryManager.Track(image.Address, image.GuestSize, image)
	if !isUpgrade {
		image.MarkCpuModified(t.currentGuestFrame)
	} else {
		if image.HasSync(vulkan.ImageSyncCpuModified) {
			image.MarkCpuModified(t.currentGuestFrame)
		} else if image.HasSync(vulkan.ImageSyncGpuModified) {
			image.MarkGpuModified(t.currentGuestFrame)
		} else {
			image.MarkSynced(t.currentGuestFrame)
		}
	}

	logger.Printf("registered image at 0x%X (%dx%d).\n",
		image.Address, image.FirstDescriptor.Width, image.FirstDescriptor.Height,
	)
}

func (t *GpuTranslator) unregisterImage(image *vulkan.VulkanImage) {
	structs.GlobalMemoryManager.Untrack(image.Address, image.GuestSize, image)
	t.imagesMutex.Lock()
	delete(t.images, image.Address)
	t.imagesMutex.Unlock()

	logger.Printf("deleted image at 0x%X (%dx%d).\n",
		image.Address, image.FirstDescriptor.Width, image.FirstDescriptor.Height,
	)
}
