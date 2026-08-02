package translation

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/lib_structs/gpu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	vk "github.com/goki/vulkan"
	"github.com/gookit/color"
)

func (t *GpuTranslator) Dispatch(frame uint64, dispatch *gpu.LiverpoolDispatch) {
	// Wait for all prior draw work before compute reads rendered attachments.
	vk.CmdPipelineBarrier(t.commandBuffer.CommandBuffer,
		vk.PipelineStageFlags(vk.PipelineStageAllCommandsBit),
		vk.PipelineStageFlags(vk.PipelineStageComputeShaderBit),
		0, 1, []vk.MemoryBarrier{{
			SType: vk.StructureTypeMemoryBarrier,
			SrcAccessMask: vk.AccessFlags(vk.AccessColorAttachmentWriteBit |
				vk.AccessDepthStencilAttachmentWriteBit |
				vk.AccessShaderWriteBit |
				vk.AccessTransferWriteBit),
			DstAccessMask: vk.AccessFlags(vk.AccessShaderReadBit | vk.AccessShaderWriteBit),
		}}, 0, nil, 0, nil,
	)

	// Get buffer addresses.
	t.userDataBuffersMutex.Lock()
	userDataOffset := t.userDataOffsets[dispatch.UserDataHash]
	t.userDataBuffersMutex.Unlock()

	// Push constants to shader.
	pushData := spirvStructs.PushConstants{
		UserDataAddress:         t.userDataBufferAddress + uint64(userDataOffset),
		OnionMemoryBaseAddress:  GlobalAllocator.DeviceAddress,
		GarlicMemoryBaseAddress: GlobalGpuAllocator.DeviceAddress,
		UserSgprCount:           gpu.DecodeUserSgprCount(dispatch.ComputeShRsrc2),
		ShaderRsrc2:             dispatch.ComputeShRsrc2,
	}
	vk.CmdPushConstants(
		t.commandBuffer.CommandBuffer, t.pipelineLayout,
		vk.ShaderStageFlags(vk.ShaderStageVertexBit|vk.ShaderStageComputeBit),
		0, spirvStructs.PushConstantsSize,
		unsafe.Pointer(&pushData),
	)

	// Dispatch.
	if logger.LogRenderer {
		logger.Printf("[%s] Dispatching %s/%s/%s (compute=%s, userData=%s, userReg=%s).\n",
			color.Blue.Sprintf("Frame %d", frame),
			color.Yellow.Sprintf("0x%X", dispatch.DimX),
			color.Yellow.Sprintf("0x%X", dispatch.DimY),
			color.Yellow.Sprintf("0x%X", dispatch.DimZ),
			color.Yellow.Sprintf("0x%X", t.activeComputeShader.GcnShader.Address),
			color.Yellow.Sprintf("0x%X", dispatch.UserDataHash),
			color.Green.Sprint(pushData.UserSgprCount),
		)
	}
	dispatchDimX := dispatch.DimX
	if dispatch.ThreadX > 1024 {
		splitFactor := (dispatch.ThreadX + 1023) / 1024
		dispatchDimX *= splitFactor
	}
	vk.CmdDispatch(t.commandBuffer.CommandBuffer, dispatchDimX, dispatch.DimY, dispatch.DimZ)

	// Global memory barrier to make compute writes visible.
	vk.CmdPipelineBarrier(t.commandBuffer.CommandBuffer,
		vk.PipelineStageFlags(vk.PipelineStageComputeShaderBit),
		vk.PipelineStageFlags(vk.PipelineStageAllCommandsBit|vk.PipelineStageHostBit),
		0, 1, []vk.MemoryBarrier{{
			SType:         vk.StructureTypeMemoryBarrier,
			SrcAccessMask: vk.AccessFlags(vk.AccessShaderWriteBit),
			DstAccessMask: vk.AccessFlags(vk.AccessColorAttachmentWriteBit | vk.AccessShaderReadBit | vk.AccessIndirectCommandReadBit |
				vk.AccessIndexReadBit | vk.AccessVertexAttributeReadBit | vk.AccessTransferReadBit | vk.AccessHostReadBit),
		}}, 0, nil, 0, nil,
	)

	// Mark image store op images as GPU modified.
	if t.activeComputeStoreTargets != nil {
		for _, image := range t.activeComputeStoreTargets {
			image.MarkGpuModified(t.currentGuestFrame)
		}
	}

	// Mark buffer store op images as CPU modified.
	if t.activeComputeStoreBufferTargets != nil {
		for _, image := range t.activeComputeStoreBufferTargets {
			image.MarkCpuModified(t.currentGuestFrame)
		}
	}
}
