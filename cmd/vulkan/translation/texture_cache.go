package translation

import (
	"fmt"

	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan"
	"go101.org/nstd"
)

func (t *GpuTranslator) registerImage(image *vulkan.VulkanImage, isUpgrade bool) {
	size := vulkan.DescriptorRegionSize(image.FirstDescriptor)
	image.GuestSize = size
	t.images[image.Address] = image
	structs.GlobalMemoryManager.Track(image.Address, size, image)

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
	if image.Address == 0xFE4DC9400 {
		fmt.Print()
	}
}

// unregisterImage removes page watchers, map references, and page indexing for an image.
func (t *GpuTranslator) unregisterImage(address uintptr) {
	image, ok := t.images[address]
	if !ok {
		return
	}
	if image.GuestSize > 0 {
		structs.GlobalMemoryManager.Untrack(image.Address, image.GuestSize, image)
		image.Destroy(t.handles.Device)
	}
	delete(t.images, address)
	logger.Printf("deleted image at 0x%X (%dx%d).\n",
		image.Address, image.FirstDescriptor.Width, image.FirstDescriptor.Height,
	)
}
