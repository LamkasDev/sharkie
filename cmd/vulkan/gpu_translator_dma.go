package vulkan

import (
	"github.com/LamkasDev/sharkie/cmd/structs/gpu"
	vk "github.com/goki/vulkan"
)

func dmaRangesOverlap(aStart, aSize, bStart, bSize uintptr) bool {
	if aSize == 0 || bSize == 0 {
		return false
	}
	aEnd := aStart + aSize
	bEnd := bStart + bSize
	return aStart < bEnd && bStart < aEnd
}

func (t *GpuTranslator) processDmaCopy(frame uint64, cb vk.CommandBuffer, copy *gpu.LiverpoolDmaCopy) {
	srcBuffer, srcOffset, err1 := t.GetBufferFromAddress(copy.SrcAddress)
	dstBuffer, dstOffset, err2 := t.GetBufferFromAddress(copy.DstAddress)

	if err1 == nil && err2 == nil {
		// Perform raw linear copy between the buffers using Vulkan.
		vk.CmdCopyBuffer(cb, srcBuffer, dstBuffer, 1, []vk.BufferCopy{{
			SrcOffset: vk.DeviceSize(srcOffset),
			DstOffset: vk.DeviceSize(dstOffset),
			Size:      vk.DeviceSize(copy.Count * 4),
		}})

		// Add a barrier after the buffer copy.
		vk.CmdPipelineBarrier(cb,
			vk.PipelineStageFlags(vk.PipelineStageTransferBit),
			vk.PipelineStageFlags(vk.PipelineStageTransferBit|vk.PipelineStageComputeShaderBit|vk.PipelineStageAllGraphicsBit),
			0, 1, []vk.MemoryBarrier{{
				SType:         vk.StructureTypeMemoryBarrier,
				SrcAccessMask: vk.AccessFlags(vk.AccessTransferWriteBit),
				DstAccessMask: vk.AccessFlags(vk.AccessTransferReadBit | vk.AccessShaderReadBit | vk.AccessShaderWriteBit),
			}}, 0, nil, 0, nil,
		)
	}
}
