package translation

import (
	"github.com/LamkasDev/sharkie/cmd/lib_structs"
	"github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan"
)

func (t *GpuTranslator) registerImage(image *vulkan.VulkanImage) {
	size := vulkan.DescriptorRegionSize(image.FirstDescriptor)
	prevGpuModified := image.SyncFlags & vulkan.ImageSyncGpuModified
	image.GuestSize = size
	image.SyncFlags = image.SyncFlags &^ vulkan.ImageSyncGpuModified
	image.SyncFlags = image.SyncFlags | prevGpuModified
	image.FrameTouched = t.currentGuestFrame
	t.images[image.Address] = image
	t.indexImagePages(image.Address, size)
	structs.GlobalMemoryManager.TrackRegion(image.Address, size)
}

func (t *GpuTranslator) indexImagePages(addr, size uintptr) {
	end := addr + size
	for page := addr >> lib_structs.SystemPageShift; (page << lib_structs.SystemPageShift) < end; page++ {
		set := t.imagePages[page]
		if set == nil {
			set = map[uintptr]struct{}{}
			t.imagePages[page] = set
		}
		set[addr] = struct{}{}
	}
}

func (t *GpuTranslator) unindexImagePages(address, size uintptr) {
	end := address + size
	for page := address >> lib_structs.SystemPageShift; (page << lib_structs.SystemPageShift) < end; page++ {
		if set, ok := t.imagePages[page]; ok {
			delete(set, address)
			if len(set) == 0 {
				delete(t.imagePages, page)
			}
		}
	}
}

func (t *GpuTranslator) forEachOverlap(address, size uintptr, fn func(*vulkan.VulkanImage)) {
	seen := map[uintptr]struct{}{}
	end := address + size
	for page := address >> lib_structs.SystemPageShift; (page << lib_structs.SystemPageShift) < end; page++ {
		for imageAddr := range t.imagePages[page] {
			if _, ok := seen[imageAddr]; ok {
				continue
			}
			seen[imageAddr] = struct{}{}
			image, ok := t.images[imageAddr]
			if !ok {
				continue
			}
			imageEnd := image.Address + image.GuestSize
			if address < imageEnd && image.Address < end {
				fn(image)
			}
		}
	}
}

// markGpuModifiedRegion flags images overlapping the written guest range.
// Does not install CPU read traps - RTs/textures live in VkImages; trapping guest
// pages caused a SIGSEGV storm (every CPU read -> downloadVkImageToGuest + GPU stall).
func (t *GpuTranslator) markGpuModifiedRegion(address, size uintptr) {
	t.forEachOverlap(address, size, func(image *vulkan.VulkanImage) {
		image.SyncFlags = image.SyncFlags | vulkan.ImageSyncGpuModified | vulkan.ImageSyncGuestUploaded
		image.SyncFlags = image.SyncFlags &^ vulkan.ImageSyncCpuDirty
	})
}

// unregisterImage removes page watchers, map references, and page indexing for an image.
func (t *GpuTranslator) unregisterImage(address uintptr) {
	image, ok := t.images[address]
	if !ok {
		return
	}
	if image.GuestSize > 0 {
		structs.GlobalMemoryManager.UntrackRegion(image.Address, image.GuestSize)
		t.unindexImagePages(image.Address, image.GuestSize)
		image.Destroy(t.handles.Device)
	}
	delete(t.images, address)
}

// invalidateImageMemory handles CPU write faults overlapping cached images.
// Write traps are cleared lazily: unregisterImage runs at the next guest upload (RefreshImageFromGuest).
func (t *GpuTranslator) invalidateImageMemory(address, size uintptr) {
	t.forEachOverlap(address, size, func(image *vulkan.VulkanImage) {
		image.SetSync(vulkan.ImageSyncCpuDirty)
		image.ClearSync(vulkan.ImageSyncGpuModified)
	})
}

// unmapImageMemory removes page index entries for images overlapping a guest unmap.
func (t *GpuTranslator) unmapImageMemory(address, size uintptr) {
	// Collect images first to avoid modifying t.imagePages during iteration
	var imagesToUnmap []*vulkan.VulkanImage
	t.forEachOverlap(address, size, func(image *vulkan.VulkanImage) {
		imagesToUnmap = append(imagesToUnmap, image)
	})

	for _, image := range imagesToUnmap {
		t.unregisterImage(image.Address)
	}
}
