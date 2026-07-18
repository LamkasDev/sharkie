package vulkan

import (
	vk "github.com/goki/vulkan"
)

const (
	tileModeDisplayLinearAligned = 8
	tileModeDisplayLinearGeneral = 31
)

func isLinearTileMode(tilingIndex uint8) bool {
	return tilingIndex == tileModeDisplayLinearAligned || tilingIndex == tileModeDisplayLinearGeneral
}

func is1DTiledMode(tilingIndex uint8) bool {
	switch tilingIndex {
	case 5, 9, 13, 19: // Depth1DThin, Display1DThin, Thin1DThin, Thick1DThick
		return true
	default:
		return false
	}
}

func isMacroTiledMode(tilingIndex uint8) bool {
	return !isLinearTileMode(tilingIndex) && !is1DTiledMode(tilingIndex)
}

func usesDisplayMicroTiling(tilingIndex uint8) bool {
	switch tilingIndex {
	case 6, 7, 8, 9, 10, 11, 12, 31:
		return true
	default:
		return false
	}
}

func ImageBarrier(commandBuffer *VulkanCommandBuffer, image *VulkanImage,
	newLayout vk.ImageLayout,
	dstAccess vk.AccessFlags,
	dstStage vk.PipelineStageFlags,
	aspectMask vk.ImageAspectFlags,
) {
	vk.CmdPipelineBarrier(commandBuffer.CommandBuffer,
		image.ImageStage, dstStage,
		0, 0, nil, 0, nil,
		1, []vk.ImageMemoryBarrier{{
			SType:               vk.StructureTypeImageMemoryBarrier,
			OldLayout:           image.ImageLayout,
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
			SrcAccessMask: image.ImageAccess,
			DstAccessMask: dstAccess,
		}},
	)
	image.ImageLayout = newLayout
	image.ImageAccess = dstAccess
	image.ImageStage = dstStage
}
