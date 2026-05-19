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

func (t *GpuTranslator) DmaCopy(frame uint64, commandBuffer vk.CommandBuffer, dmaCopy *gpu.LiverpoolDmaCopy) {
	if t.activePass != vk.NullRenderPass {
		t.EndRenderPass(commandBuffer)
	}

	// Get buffers.
	srcBuffer, srcOffset, err1 := t.GetBufferFromAddress(dmaCopy.SrcAddress)
	dstBuffer, dstOffset, err2 := t.GetBufferFromAddress(dmaCopy.DstAddress)

	if err1 == nil && err2 == nil {
		// Perform raw linear copy between the buffers using Vulkan.
		vk.CmdCopyBuffer(commandBuffer, srcBuffer, dstBuffer, 1, []vk.BufferCopy{{
			SrcOffset: vk.DeviceSize(srcOffset),
			DstOffset: vk.DeviceSize(dstOffset),
			Size:      vk.DeviceSize(dmaCopy.Count * 4),
		}})

		// Add a barrier after the buffer copy.
		vk.CmdPipelineBarrier(commandBuffer,
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
