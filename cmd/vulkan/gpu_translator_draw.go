package vulkan

import (
	"math"
	"time"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/logger"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/LamkasDev/sharkie/cmd/structs/gpu"
	vk "github.com/goki/vulkan"
	"github.com/gookit/color"
	"go101.org/nstd"
)

var startTime time.Time

func init() {
	startTime = time.Now()
}

func (t *GpuTranslator) Draw(frame uint64, commandBuffer vk.CommandBuffer, draw *gpu.LiverpoolDraw) {
	if t.activePass == vk.NullRenderPass || t.activePipeline == vk.NullPipeline {
		return
	}

	// Handle explicit clears.
	/* if draw.DbDepthClearEnable || draw.DbStencilClearEnable {
		clearDepth := vk.ClearAttachment{
			AspectMask: vk.ImageAspectFlags(vk.ImageAspectDepthBit | vk.ImageAspectStencilBit),
			ClearValue: vk.ClearValue{},
		}
		clearDepth.ClearValue.SetDepthStencil(math.Float32frombits(draw.DbDepthClearValue), draw.DbStencilClearValue)
		vk.CmdClearAttachments(commandBuffer, 1, []vk.ClearAttachment{clearDepth}, 1, []vk.ClearRect{{
			Rect:           vk.Rect2D{Extent: vk.Extent2D{Width: t.activeSurface.Value.width, Height: t.activeSurface.Value.height}},
			BaseArrayLayer: 0,
			LayerCount:     1,
		}})
	} */

	// Bind bindless/discovery descriptor sets.
	vk.CmdBindDescriptorSets(commandBuffer, vk.PipelineBindPointGraphics, t.pipelineLayout, spirvStructs.DescriptorSetSlotBindless, 1, []vk.DescriptorSet{t.bindlessDescriptorSet}, 0, nil)
	vk.CmdBindDescriptorSets(commandBuffer, vk.PipelineBindPointGraphics, t.pipelineLayout, spirvStructs.DescriptorSetSlotDiscovery, 1, []vk.DescriptorSet{t.discoveryDescriptorSet}, 0, nil)

	// Get buffer addresses.
	t.userDataBuffersMutex.Lock()
	userDataOffset := t.userDataOffsets[draw.UserDataHash]
	t.userDataBuffersMutex.Unlock()

	// Push constants to vertex shader.
	pushDataVs := spirvStructs.PushConstants{
		UserDataAddress:         t.userDataBufferAddress + uint64(userDataOffset),
		OnionMemoryBaseAddress:  GlobalAllocator.DeviceAddress,
		GarlicMemoryBaseAddress: GlobalGpuAllocator.DeviceAddress,
		UserSgprCount:           gpu.DecodeUserSgprCount(draw.VertexShRsrc2),
		VteControl:              t.activeVteControl,
	}
	vk.CmdPushConstants(commandBuffer, t.pipelineLayout, vk.ShaderStageFlags(vk.ShaderStageVertexBit|vk.ShaderStageComputeBit), 0, spirvStructs.PushConstantsSize, unsafe.Pointer(&pushDataVs))

	// Push constants to fragment shader.
	pushDataFs := spirvStructs.PushConstants{
		UserDataAddress:         t.userDataBufferAddress + uint64(userDataOffset),
		OnionMemoryBaseAddress:  GlobalAllocator.DeviceAddress,
		GarlicMemoryBaseAddress: GlobalGpuAllocator.DeviceAddress,
		UserSgprCount:           gpu.DecodeUserSgprCount(draw.PixelShRsrc2),
		ShaderRsrc2:             draw.PixelShRsrc2,
		VteControl:              t.activeVteControl,
	}
	vk.CmdPushConstants(commandBuffer, t.pipelineLayout, vk.ShaderStageFlags(vk.ShaderStageFragmentBit), spirvStructs.PushConstantsSize, spirvStructs.PushConstantsSize, unsafe.Pointer(&pushDataFs))

	// Draw.
	logger.Printf("[%s] Drawing %s vertices (userData=%s, topology=%s, indexed=%s).\n",
		color.Blue.Sprintf("Frame %d", frame),
		color.Green.Sprint(draw.VertexCount),
		color.Yellow.Sprintf("0x%X", draw.UserDataHash),
		color.Green.Sprint(draw.PrimType),
		color.Green.Sprint(nstd.Btoi(draw.IsIndexed)),
	)
	if draw.IsIndexed {
		targetBuffer, relativeOffset, err := t.GetBufferFromAddress(draw.IndexBaseAddress)
		if err != nil {
			panic(err)
		}
		indexType := vk.IndexTypeUint16
		if draw.IndexType == 1 {
			indexType = vk.IndexTypeUint32
		}
		vk.CmdBindIndexBuffer(commandBuffer, targetBuffer, vk.DeviceSize(relativeOffset), indexType)
		vk.CmdDrawIndexed(commandBuffer, draw.IndexCount, draw.InstanceCount, 0, int32(draw.BaseVertexOffset), 0)
	} else {
		vk.CmdDraw(commandBuffer, draw.VertexCount, draw.InstanceCount, 0, 0)
	}
}

func (t *GpuTranslator) SetDynamicState(commandBuffer vk.CommandBuffer, dynamicState *gpu.LiverpoolSetDynamicState) {
	t.activeVteControl = dynamicState.VteControl

	// Derive viewport from GCN scale/offset/control registers.
	vpxScale := dynamicState.VpXScale
	if !dynamicState.VpXScaleEnable {
		vpxScale = 1.0
	}
	vpxOffset := dynamicState.VpXOffset
	if !dynamicState.VpXOffsetEnable {
		vpxOffset = 0.0
	}
	vpyScale := dynamicState.VpYScale
	if !dynamicState.VpYScaleEnable {
		vpyScale = 1.0
	}
	vpyOffset := dynamicState.VpYOffset
	if !dynamicState.VpYOffsetEnable {
		vpyOffset = 0.0
	}
	vpZScale := dynamicState.VpZScale
	if !dynamicState.VpZScaleEnable {
		vpZScale = 1.0
	}
	vpZOffset := dynamicState.VpZOffset
	if !dynamicState.VpZOffsetEnable {
		vpZOffset = 0.0
	}

	// Process viewport transforms.
	vpWidth := vpxScale * 2
	vpHeight := vpyScale * 2
	vpX, vpY := vpxOffset-vpxScale, vpyOffset-vpyScale
	if vpWidth == 0 || vpHeight == 0 {
		vpWidth, vpHeight = float32(t.activeSurface.Value.width), float32(t.activeSurface.Value.height)
		vpX, vpY = 0, 0
	}
	vk.CmdSetViewport(commandBuffer, 0, 1, []vk.Viewport{{
		X: vpX, Y: vpY,
		Width: vpWidth, Height: vpHeight,
		MinDepth: vpZOffset,
		MaxDepth: vpZOffset + vpZScale,
	}})

	// Setup the scissor from TL/BR registers.
	scissorX := int32(dynamicState.ScissorTl & 0x7FFF)
	scissorY := int32((dynamicState.ScissorTl >> 16) & 0x7FFF)
	scissorW := uint32((dynamicState.ScissorBr & 0x7FFF) - uint32(scissorX))
	scissorH := uint32(((dynamicState.ScissorBr >> 16) & 0x7FFF) - uint32(scissorY))
	if scissorW <= 0 || scissorH <= 0 {
		scissorW = t.activeSurface.Value.width
		scissorH = t.activeSurface.Value.height
		scissorX, scissorY = 0, 0
	}
	vk.CmdSetScissor(commandBuffer, 0, 1, []vk.Rect2D{{
		Offset: vk.Offset2D{X: scissorX, Y: scissorY},
		Extent: vk.Extent2D{Width: scissorW, Height: scissorH},
	}})

	// Setup blend constants.
	vk.CmdSetBlendConstants(commandBuffer, &[4]float32{
		math.Float32frombits(dynamicState.BlendRed),
		math.Float32frombits(dynamicState.BlendGreen),
		math.Float32frombits(dynamicState.BlendBlue),
		math.Float32frombits(dynamicState.BlendAlpha),
	})
}
