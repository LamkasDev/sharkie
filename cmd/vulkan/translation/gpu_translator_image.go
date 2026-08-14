package translation

import (
	"github.com/LamkasDev/sharkie/cmd/logger"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan"
)

func (t *GpuTranslator) GetImageByAddress(address uintptr) *vulkan.VulkanImageGroup {
	t.imageGroupsMutex.Lock()
	group, _ := t.imageGroups[address]
	t.imageGroupsMutex.Unlock()

	return group
}

func (t *GpuTranslator) GetImage(descriptor spirvStructs.ImageDescriptor, compSwap uint32, isSurface bool) (*vulkan.VulkanImage, error, bool) {
	t.imageGroupsMutex.Lock()
	group, ok := t.imageGroups[descriptor.BaseAddress]
	if !ok {
		guestSize := vulkan.DescriptorGuestSize(descriptor)
		group = vulkan.NewVulkanImageGroup(descriptor.BaseAddress, guestSize)
	}
	t.imageGroupsMutex.Unlock()

	// Get image from the group.
	oldSize := group.GuestSize
	image, err, created := group.GetImage(t.handles, descriptor, compSwap, isSurface, t.commandBuffer, t.currentGuestFrame, t.GetLinearBuffer)
	if err != nil {
		return nil, err, false
	}

	// Expand memory tracking region.
	if !ok {
		t.imageGroups[descriptor.BaseAddress] = group
		structs.GlobalMemoryManager.Track(group.Address, group.GuestSize, group)
		if logger.LogRenderer {
			logger.Printf("registered image group at 0x%X (0x%X bytes).\n", group.Address, group.GuestSize)
		}
	}
	if group.GuestSize > oldSize {
		structs.GlobalMemoryManager.Untrack(group.Address, oldSize, group)
		structs.GlobalMemoryManager.Track(group.Address, group.GuestSize, group)
		if logger.LogRenderer {
			logger.Printf("expanded image group at 0x%X (0x%X bytes -> 0x%X bytes).\n", group.Address, oldSize, group.GuestSize)
		}
	}

	return image, nil, created
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
