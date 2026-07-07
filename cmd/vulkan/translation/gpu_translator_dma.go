package translation

import (
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gpu"
	vk "github.com/goki/vulkan"
)

func (t *GpuTranslator) DmaCopy(frame uint64, dmaCopy *gpu.LiverpoolDmaCopy) {
	t.EndRenderPass()

	copySize := uintptr(dmaCopy.Count * 4)
	if err := t.DownloadRegionVkImages(dmaCopy.SrcAddress, copySize); err != nil {
		panic(err)
	}

	srcBuffer, srcOffset, err1 := t.GetBufferFromAddress(dmaCopy.SrcAddress)
	dstBuffer, dstOffset, err2 := t.GetBufferFromAddress(dmaCopy.DstAddress)
	if err1 != nil {
		panic(err1)
	}
	if err2 != nil {
		panic(err2)
	}

	vk.CmdCopyBuffer(t.commandBuffer, srcBuffer, dstBuffer, 1, []vk.BufferCopy{{
		SrcOffset: vk.DeviceSize(srcOffset),
		DstOffset: vk.DeviceSize(dstOffset),
		Size:      vk.DeviceSize(copySize),
	}})

	vk.CmdPipelineBarrier(t.commandBuffer,
		vk.PipelineStageFlags(vk.PipelineStageTransferBit),
		vk.PipelineStageFlags(vk.PipelineStageTransferBit),
		0, 1, []vk.MemoryBarrier{{
			SType:         vk.StructureTypeMemoryBarrier,
			SrcAccessMask: vk.AccessFlags(vk.AccessTransferWriteBit),
			DstAccessMask: vk.AccessFlags(vk.AccessTransferReadBit),
		}}, 0, nil, 0, nil,
	)

	// Upload DMA destination into any overlapping VkImages (guest buffer is now fresh in-GPU-order).
	if err := t.UploadRegionVkImages(dmaCopy.DstAddress, copySize); err != nil {
		panic(err)
	}

	vk.CmdPipelineBarrier(t.commandBuffer,
		vk.PipelineStageFlags(vk.PipelineStageTransferBit),
		vk.PipelineStageFlags(vk.PipelineStageTransferBit|vk.PipelineStageComputeShaderBit|vk.PipelineStageAllGraphicsBit),
		0, 1, []vk.MemoryBarrier{{
			SType:         vk.StructureTypeMemoryBarrier,
			SrcAccessMask: vk.AccessFlags(vk.AccessTransferWriteBit),
			DstAccessMask: vk.AccessFlags(vk.AccessTransferReadBit | vk.AccessShaderReadBit | vk.AccessShaderWriteBit | vk.AccessIndexReadBit | vk.AccessVertexAttributeReadBit),
		}}, 0, nil, 0, nil,
	)
}
