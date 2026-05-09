package vulkan

import (
	vk "github.com/goki/vulkan"
)

func (t *GpuTranslator) uploadDataToImage(data []uint8, image vk.Image, width, height, bpp uint32) {
	size := vk.DeviceSize(len(data))
	stagingBuffer, stagingMem, err := t.AllocBuffer(size,
		vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
	if err != nil {
		return
	}

	ptr := t.handles.MapMemory(stagingMem, size)
	copy(ptr, data)
	vk.UnmapMemory(t.handles.Device, stagingMem)

	cb := t.handles.AllocateCommandBuffer(t.pool)
	vk.BeginCommandBuffer(cb, &vk.CommandBufferBeginInfo{
		SType: vk.StructureTypeCommandBufferBeginInfo,
		Flags: vk.CommandBufferUsageFlags(vk.CommandBufferUsageOneTimeSubmitBit),
	})

	t.imageBarrier(cb, image,
		vk.ImageLayoutUndefined, vk.ImageLayoutTransferDstOptimal,
		0, vk.AccessFlags(vk.AccessTransferWriteBit),
		vk.PipelineStageFlags(vk.PipelineStageTopOfPipeBit), vk.PipelineStageFlags(vk.PipelineStageTransferBit))

	vk.CmdCopyBufferToImage(cb, stagingBuffer, image, vk.ImageLayoutTransferDstOptimal, 1, []vk.BufferImageCopy{{
		BufferOffset:      0,
		BufferRowLength:   0,
		BufferImageHeight: 0,
		ImageSubresource: vk.ImageSubresourceLayers{
			AspectMask: vk.ImageAspectFlags(vk.ImageAspectColorBit),
			LayerCount: 1,
		},
		ImageOffset: vk.Offset3D{X: 0, Y: 0, Z: 0},
		ImageExtent: vk.Extent3D{Width: width, Height: height, Depth: 1},
	}})

	t.imageBarrier(cb, image,
		vk.ImageLayoutTransferDstOptimal, vk.ImageLayoutShaderReadOnlyOptimal,
		vk.AccessFlags(vk.AccessTransferWriteBit), vk.AccessFlags(vk.AccessShaderReadBit),
		vk.PipelineStageFlags(vk.PipelineStageTransferBit), vk.PipelineStageFlags(vk.PipelineStageAllGraphicsBit))

	vk.EndCommandBuffer(cb)

	vk.QueueSubmit(t.handles.GraphicsQueue, 1, []vk.SubmitInfo{{
		SType:              vk.StructureTypeSubmitInfo,
		CommandBufferCount: 1,
		PCommandBuffers:    []vk.CommandBuffer{cb},
	}}, vk.NullFence)
	vk.QueueWaitIdle(t.handles.GraphicsQueue)

	vk.FreeCommandBuffers(t.handles.Device, t.pool, 1, []vk.CommandBuffer{cb})
	vk.DestroyBuffer(t.handles.Device, stagingBuffer, nil)
	vk.FreeMemory(t.handles.Device, stagingMem, nil)
}

func (t *GpuTranslator) imageBarrier(commandBuffer vk.CommandBuffer, image vk.Image,
	oldLayout, newLayout vk.ImageLayout,
	srcAccess, dstAccess vk.AccessFlags,
	srcStage, dstStage vk.PipelineStageFlags,
) {
	vk.CmdPipelineBarrier(commandBuffer,
		srcStage, dstStage,
		0, 0, nil, 0, nil,
		1, []vk.ImageMemoryBarrier{{
			SType:               vk.StructureTypeImageMemoryBarrier,
			OldLayout:           oldLayout,
			NewLayout:           newLayout,
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
			DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image:               image,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				BaseMipLevel:   0,
				LevelCount:     vk.RemainingMipLevels,
				BaseArrayLayer: 0,
				LayerCount:     vk.RemainingArrayLayers,
			},
			SrcAccessMask: srcAccess,
			DstAccessMask: dstAccess,
		}},
	)
}
