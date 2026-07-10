package translation

import (
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gpu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	vk "github.com/goki/vulkan"
	"github.com/gookit/color"
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
	if logger.LogRenderer {
		logger.Printf("[%s] DMA copy of %s bytes from %s to %s.\n",
			color.Blue.Sprintf("Frame %d", frame),
			color.Yellow.Sprintf("0x%X", dmaCopy.Count),
			color.Yellow.Sprintf("0x%X", dmaCopy.SrcAddress),
			color.Yellow.Sprintf("0x%X", dmaCopy.DstAddress),
		)
	}
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
