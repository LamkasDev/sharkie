package translation

import (
	"unsafe"

	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gpu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/spirv"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan"
	vk "github.com/goki/vulkan"
	"github.com/gookit/color"
)

func (t *GpuTranslator) Dispatch(frame uint64, dispatch *gpu.LiverpoolDispatch) {
	t.EndRenderPass()

	// Wait for all prior draw work before compute reads rendered attachments.
	vk.CmdPipelineBarrier(t.commandBuffer.CommandBuffer,
		vk.PipelineStageFlags(vk.PipelineStageAllGraphicsBit),
		vk.PipelineStageFlags(vk.PipelineStageComputeShaderBit),
		0, 1, []vk.MemoryBarrier{{
			SType:         vk.StructureTypeMemoryBarrier,
			SrcAccessMask: vk.AccessFlags(vk.AccessColorAttachmentWriteBit | vk.AccessDepthStencilAttachmentWriteBit | vk.AccessShaderWriteBit | vk.AccessIndirectCommandReadBit | vk.AccessIndexReadBit | vk.AccessVertexAttributeReadBit),
			DstAccessMask: vk.AccessFlags(vk.AccessShaderReadBit | vk.AccessShaderWriteBit),
		}}, 0, nil, 0, nil,
	)

	// Get scoped compute shader.
	csSpirv := t.GetShaderWithContext(dispatch.ComputeShader, spirv.SpirvShaderContext{
		ThreadX: dispatch.ThreadX,
		ThreadY: dispatch.ThreadY,
		ThreadZ: dispatch.ThreadZ,
	})
	csModule, err := t.GetShaderModule(csSpirv)
	if err != nil {
		panic(err)
	}

	// Get pipeline for compute module.
	pipeline, err := t.GetComputePipeline(vulkan.ComputePipelineRequest{
		ComputeModule: csModule,
		ComputePipelineKey: vulkan.ComputePipelineKey{
			ComputeModuleAddress: dispatch.ComputeShader.Address,
		},
	})
	if err != nil {
		panic(err)
	}

	// Bind pipeline.
	vk.CmdBindPipeline(t.commandBuffer.CommandBuffer, vk.PipelineBindPointCompute, pipeline)

	// Get buffer addresses.
	t.userDataBuffersMutex.Lock()
	userDataOffset := t.userDataOffsets[dispatch.UserDataHash]
	userData, _ := gpu.GlobalUserDataSnapshots[dispatch.UserDataHash]
	t.userDataBuffersMutex.Unlock()

	// Bind resources.
	staticSetToBind := t.staticDescriptorSet
	storeTargets, activeStaticSet, err := t.BindResources([]*spirv.SpirvShader{csSpirv}, userData)
	if err != nil {
		panic(err)
	}
	if activeStaticSet != vk.NullDescriptorSet {
		staticSetToBind = activeStaticSet
	}
	vk.CmdBindDescriptorSets(
		t.commandBuffer.CommandBuffer, vk.PipelineBindPointCompute,
		t.pipelineLayout, spirvStructs.DescriptorSetSlotStatic,
		1, []vk.DescriptorSet{staticSetToBind},
		0, nil,
	)

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
			color.Yellow.Sprintf("0x%X", dispatch.ComputeShader.Address),
			color.Yellow.Sprintf("0x%X", dispatch.UserDataHash),
			color.Green.Sprint(pushData.UserSgprCount),
		)
	}
	vk.CmdDispatch(t.commandBuffer.CommandBuffer, dispatch.DimX, dispatch.DimY, dispatch.DimZ)

	// Global memory barrier to make compute writes visible.
	vk.CmdPipelineBarrier(t.commandBuffer.CommandBuffer,
		vk.PipelineStageFlags(vk.PipelineStageComputeShaderBit),
		vk.PipelineStageFlags(vk.PipelineStageAllCommandsBit),
		0, 1, []vk.MemoryBarrier{{
			SType:         vk.StructureTypeMemoryBarrier,
			SrcAccessMask: vk.AccessFlags(vk.AccessShaderWriteBit),
			DstAccessMask: vk.AccessFlags(vk.AccessShaderReadBit | vk.AccessTransferReadBit | vk.AccessTransferWriteBit),
		}}, 0, nil, 0, nil,
	)

	// Mark storage-written addresses as GPU modified so CPU doesn't overwrite them.
	for _, image := range storeTargets {
		image.MarkGpuModified(t.currentGuestFrame)
	}

	t.FlushPendingResourceBarriers(t.commandBuffer, 0)
	t.pendingComputeBarrier = true
}
