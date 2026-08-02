package vulkan

import (
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan/gcn"
)

func DescriptorGuestSize(descriptor spirvStructs.ImageDescriptor) uintptr {
	bpp := gcn.GetBytesPerPixel(descriptor.DataFormat)
	layouts := computeMipLayouts(descriptor, mipLevelCount(descriptor))
	if len(layouts) > 0 {
		last := layouts[len(layouts)-1]
		return uintptr(last.Offset + last.Size)
	}

	return uintptr(descriptor.Pitch) * uintptr(descriptor.Height) * uintptr(bpp)
}
