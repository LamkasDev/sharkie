package vulkan

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/spirv"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/LamkasDev/sharkie/cmd/structs/gpu"
	vk "github.com/goki/vulkan"
	"github.com/gookit/color"
)

func (t *GpuTranslator) dispatchCompute(frame uint64, commandBuffer vk.CommandBuffer, dispatch *gpu.LiverpoolComputeDispatch) {
	// Get shader module.
	csSpirv := t.GetShaderWithContext(dispatch.ComputeShader, spirv.SpirvShaderContext{
		ThreadX: dispatch.ThreadX,
		ThreadY: dispatch.ThreadY,
		ThreadZ: dispatch.ThreadZ,
	})
	csModule, err := t.GetShaderModule(csSpirv)
	if err != nil {
		return
	}

	// Get pipeline for compute module.
	request := ComputePipelineRequest{
		ComputeModule: csModule,
		ComputePipelineKey: ComputePipelineKey{
			ComputeModuleAddress: dispatch.ComputeShader.Address,
		},
	}
	pipeline, err := t.GetComputePipeline(request)
	if err != nil {
		return
	}

	// Bind pipeline.
	vk.CmdBindPipeline(commandBuffer, vk.PipelineBindPointCompute, pipeline)

	// Bind bindless descriptor set.
	vk.CmdBindDescriptorSets(commandBuffer, vk.PipelineBindPointCompute, t.pipelineLayout, spirvStructs.DescriptorSetSlotBindless, 1, []vk.DescriptorSet{t.bindlessDescriptorSet}, 0, nil)

	// Bind discovery descriptor set.
	vk.CmdBindDescriptorSets(commandBuffer, vk.PipelineBindPointCompute, t.pipelineLayout, spirvStructs.DescriptorSetSlotDiscovery, 1, []vk.DescriptorSet{t.discoveryDescriptorSet}, 0, nil)

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
		commandBuffer, t.pipelineLayout,
		vk.ShaderStageFlags(vk.ShaderStageVertexBit|vk.ShaderStageComputeBit),
		0, spirvStructs.PushConstantsSize,
		unsafe.Pointer(&pushData),
	)

	// Dispatch.
	logger.Printf("[%s] Dispatching %s/%s/%s (compute=%s, userData=%s, userReg=%s).\n",
		color.Blue.Sprintf("Frame %d", frame),
		color.Yellow.Sprintf("0x%X", dispatch.DimX),
		color.Yellow.Sprintf("0x%X", dispatch.DimY),
		color.Yellow.Sprintf("0x%X", dispatch.DimZ),
		color.Yellow.Sprintf("0x%X", dispatch.ComputeShader.Address),
		color.Yellow.Sprintf("0x%X", dispatch.UserDataHash),
		color.Green.Sprint(pushData.UserSgprCount),
	)
	vk.CmdDispatch(commandBuffer, dispatch.DimX, dispatch.DimY, dispatch.DimZ)

	// Global memory barrier to make compute writes visible.
	vk.CmdPipelineBarrier(commandBuffer,
		vk.PipelineStageFlags(vk.PipelineStageComputeShaderBit),
		vk.PipelineStageFlags(vk.PipelineStageAllCommandsBit),
		0, 1, []vk.MemoryBarrier{{
			SType:         vk.StructureTypeMemoryBarrier,
			SrcAccessMask: vk.AccessFlags(vk.AccessShaderWriteBit),
			DstAccessMask: vk.AccessFlags(vk.AccessShaderReadBit | vk.AccessTransferReadBit | vk.AccessTransferWriteBit),
		}}, 0, nil, 0, nil,
	)
}
