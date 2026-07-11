package translation

import (
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan"
	"go101.org/nstd"
)

func (t *GpuTranslator) registerImage(image *vulkan.VulkanImage, isUpgrade bool) {
	t.imagesMutex.Lock()
	t.images[image.Address] = image
	t.imagesMutex.Unlock()
	structs.GlobalMemoryManager.Track(image.Address, image.GuestSize, image)
	if !isUpgrade {
		image.MarkCpuModified()
	} else {
		if image.HasSync(vulkan.ImageSyncCpuModified) {
			image.MarkCpuModified()
		} else if image.HasSync(vulkan.ImageSyncGpuModified) {
			image.MarkGpuModified()
		} else {
			image.MarkSynced()
		}
	}

	logger.Printf("registered image at 0x%X (%dx%d, surface=%d).\n",
		image.Address, image.FirstDescriptor.Width, image.FirstDescriptor.Height,
		nstd.Btoi(image.IsSurface),
	)
}

func (t *GpuTranslator) unregisterImage(address uintptr) {
	t.imagesMutex.Lock()
	image, ok := t.images[address]
	if !ok {
		t.imagesMutex.Unlock()
		return
	}
	structs.GlobalMemoryManager.Untrack(image.Address, image.GuestSize, image)
	delete(t.images, address)
	t.imagesMutex.Unlock()

	logger.Printf("deleted image at 0x%X (%dx%d).\n",
		image.Address, image.FirstDescriptor.Width, image.FirstDescriptor.Height,
	)
}
