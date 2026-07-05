package vulkan

import (
	"fmt"
	"unsafe"

	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/LamkasDev/sharkie/cmd/structs"
	vk "github.com/goki/vulkan"
)

// detileGuestTexture reads a tiled mip from guest RAM and returns row-major pixels.
func DetileGuestTexture(descriptor spirvStructs.ImageDescriptor) ([]byte, MipLayout, error) {
	bpp := int(structs.GetBytesPerPixel(descriptor.DataFormat))
	if descriptor.Width <= 0 || descriptor.Height <= 0 || bpp <= 0 {
		return nil, MipLayout{}, fmt.Errorf("invalid texture dimensions %dx%d bpp=%d", descriptor.Width, descriptor.Height, bpp)
	}

	mipLevel := int(descriptor.BaseLevel)
	layouts := computeMipLayouts(descriptor, mipLevelCount(descriptor))
	if mipLevel >= len(layouts) {
		return nil, MipLayout{}, fmt.Errorf("mip level %d out of range", mipLevel)
	}
	layout := layouts[mipLevel]

	srcBase := descriptor.BaseAddress + uintptr(layout.Offset)
	src := unsafe.Slice((*byte)(unsafe.Pointer(srcBase)), layout.Size)

	linearSize := layout.Width * layout.Height * bpp
	linear := make([]byte, linearSize)
	DetileToLinear(src, linear, layout.Width, layout.Height, layout.Pitch, descriptor.TilingIndex, bpp)
	return linear, layout, nil
}

func ImageBarrier(commandBuffer vk.CommandBuffer, image *VulkanImage,
	newLayout vk.ImageLayout,
	dstAccess vk.AccessFlags,
	dstStage vk.PipelineStageFlags,
	aspectMask vk.ImageAspectFlags,
) {
	vk.CmdPipelineBarrier(commandBuffer,
		image.imageStage, dstStage,
		0, 0, nil, 0, nil,
		1, []vk.ImageMemoryBarrier{{
			SType:               vk.StructureTypeImageMemoryBarrier,
			OldLayout:           image.imageLayout,
			NewLayout:           newLayout,
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
			DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image:               image.Image,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     aspectMask,
				BaseMipLevel:   0,
				LevelCount:     vk.RemainingMipLevels,
				BaseArrayLayer: 0,
				LayerCount:     vk.RemainingArrayLayers,
			},
			SrcAccessMask: image.imageAccess,
			DstAccessMask: dstAccess,
		}},
	)
	image.imageLayout = newLayout
	image.imageAccess = dstAccess
	image.imageStage = dstStage
}
