package vulkan

import (
	"runtime"

	as "github.com/LamkasDev/asche"
	"github.com/LamkasDev/sharkie/cmd/logger"
	vk "github.com/goki/vulkan"
)

func (t *GpuTranslator) uploadDataToImage(data []uint8, image vk.Image, width, height, bpp uint32) {
	size := vk.DeviceSize(len(data))
	stagingBuffer, stagingMem, err := t.AllocBuffer(size,
		vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit),
		vk.MemoryPropertyFlags(vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit))
	defer vk.DestroyBuffer(t.handles.Device, stagingBuffer, nil)
	defer vk.FreeMemory(t.handles.Device, stagingMem, nil)
	if err != nil {
		return
	}

	ptr := t.handles.MapMemory(stagingMem, size)
	copy(ptr, data)
	vk.UnmapMemory(t.handles.Device, stagingMem)

	cb := t.AllocateCommandBuffer()
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
		vk.ImageLayoutTransferDstOptimal, vk.ImageLayoutGeneral,
		vk.AccessFlags(vk.AccessTransferWriteBit), vk.AccessFlags(vk.AccessShaderReadBit|vk.AccessShaderWriteBit),
		vk.PipelineStageFlags(vk.PipelineStageTransferBit), vk.PipelineStageFlags(vk.PipelineStageAllCommandsBit))
	vk.EndCommandBuffer(cb)
	defer t.FreeCommandBuffer(cb)

	// Submit and wait for completion.
	commandBuffers := []vk.CommandBuffer{cb}
	submitInfos := []vk.SubmitInfo{{
		SType:              vk.StructureTypeSubmitInfo,
		CommandBufferCount: 1,
		PCommandBuffers:    commandBuffers,
	}}

	pinner := &runtime.Pinner{}
	pinner.Pin(&commandBuffers)
	pinner.Pin(&submitInfos)
	defer pinner.Unpin()

	t.ResetWorkerFence()
	t.QueueMutex.Lock()
	result := vk.QueueSubmit(t.handles.GraphicsQueue, 1, submitInfos, t.workerFence)
	t.QueueMutex.Unlock()

	if err = as.NewError(result); err != nil {
		logger.Printf("uploadDataToImage: QueueSubmit failed: %s\n", err.Error())
		t.FreeCommandBuffer(cb)
		vk.DestroyBuffer(t.handles.Device, stagingBuffer, nil)
		vk.FreeMemory(t.handles.Device, stagingMem, nil)
		return
	}
	t.WaitOnWorkerFence()
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
