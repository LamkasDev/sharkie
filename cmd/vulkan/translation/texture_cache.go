package translation

import (
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan"
)

func (t *GpuTranslator) registerImage(image *vulkan.VulkanImage) {
	t.imagesMutex.Lock()
	t.images[image.Address] = image
	t.imagesMutex.Unlock()
	structs.GlobalMemoryManager.Track(image.Address, image.GuestSize, image)
	image.MarkCpuModified(t.currentGuestFrame)

	logger.Printf("registered image at 0x%X (%dx%d, 0x%X bytes).\n",
		image.Address, image.FirstDescriptor.Width, image.FirstDescriptor.Height,
		image.GuestSize,
	)
}

func (t *GpuTranslator) unregisterImage(image *vulkan.VulkanImage) {
	structs.GlobalMemoryManager.Untrack(image.Address, image.GuestSize, image)
	t.imagesMutex.Lock()
	delete(t.images, image.Address)
	t.imagesMutex.Unlock()
	/* deferFunc, err := image.DownloadFromVkImage(t.handles, t.commandBuffer, t.currentGuestFrame)
	if err != nil {
		panic(err)
	}
	t.handles.DeferDestroyFunction(deferFunc) */

	logger.Printf("deleted image at 0x%X (%dx%d).\n",
		image.Address, image.FirstDescriptor.Width, image.FirstDescriptor.Height,
	)
}
