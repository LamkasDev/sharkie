package translation

import (
	"math"
	"time"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/lib_structs/gpu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	vk "github.com/goki/vulkan"
	"github.com/gookit/color"
)

var startTime time.Time

func init() {
	startTime = time.Now()
}

func (t *GpuTranslator) Draw(frame uint64, draw *gpu.LiverpoolDraw) {
	// Get buffer addresses.
	t.userDataBuffersMutex.Lock()
	userDataOffset := t.userDataOffsets[draw.UserDataHash]
	t.userDataBuffersMutex.Unlock()

	// Perform depth/stencil clears before rendering the draw
	if draw.DbRenderControl.DepthClearEnable() || draw.DbRenderControl.StencilClearEnable() {
		var aspectMask vk.ImageAspectFlags
		if draw.DbRenderControl.DepthClearEnable() {
			aspectMask |= vk.ImageAspectFlags(vk.ImageAspectDepthBit)
		}
		if draw.DbRenderControl.StencilClearEnable() {
			aspectMask |= vk.ImageAspectFlags(vk.ImageAspectStencilBit)
		}

		clearValue := vk.ClearValue{}
		clearValue.SetDepthStencil(math.Float32frombits(draw.DbDepthClearValue), draw.DbStencilClearValue)
		clearAttachments := []vk.ClearAttachment{{
			AspectMask: aspectMask,
			ClearValue: clearValue,
		}}
		clearRects := []vk.ClearRect{{
			Rect: vk.Rect2D{
				Extent: vk.Extent2D{
					Width:  uint32(t.activeSurface.ImageView.Image.FirstDescriptor.Width),
					Height: uint32(t.activeSurface.ImageView.Image.FirstDescriptor.Height),
				},
			},
			BaseArrayLayer: 0,
			LayerCount:     1,
		}}
		vk.CmdClearAttachments(t.commandBuffer.CommandBuffer, 1, clearAttachments, 1, clearRects)
	}

	// Push constants to vertex shader.
	pushDataVs := spirvStructs.PushConstants{
		UserDataAddress:         t.userDataBufferAddress + uint64(userDataOffset),
		OnionMemoryBaseAddress:  GlobalAllocator.DeviceAddress,
		GarlicMemoryBaseAddress: GlobalGpuAllocator.DeviceAddress,
		UserSgprCount:           gpu.DecodeUserSgprCount(draw.VertexShRsrc2),
		VteControl:              t.activeVteControl,
		ClipControl:             t.activeClipControl,
	}
	vk.CmdPushConstants(
		t.commandBuffer.CommandBuffer, t.pipelineLayout,
		vk.ShaderStageFlags(vk.ShaderStageVertexBit|vk.ShaderStageComputeBit), 0,
		spirvStructs.PushConstantsSize, unsafe.Pointer(&pushDataVs),
	)

	// Push constants to fragment shader.
	pushDataFs := spirvStructs.PushConstants{
		UserDataAddress:         t.userDataBufferAddress + uint64(userDataOffset),
		OnionMemoryBaseAddress:  GlobalAllocator.DeviceAddress,
		GarlicMemoryBaseAddress: GlobalGpuAllocator.DeviceAddress,
		UserSgprCount:           gpu.DecodeUserSgprCount(draw.PixelShRsrc2),
		ShaderRsrc2:             draw.PixelShRsrc2,
		VteControl:              t.activeVteControl,
		ClipControl:             t.activeClipControl,
	}
	vk.CmdPushConstants(
		t.commandBuffer.CommandBuffer, t.pipelineLayout,
		vk.ShaderStageFlags(vk.ShaderStageFragmentBit), spirvStructs.PushConstantsSize,
		spirvStructs.PushConstantsSize, unsafe.Pointer(&pushDataFs),
	)

	// Draw.
	if logger.LogRenderer {
		logger.Printf("[%s] Drawing %s vertices (userData=%s, topology=%s, indexed=%s).\n",
			color.Blue.Sprintf("Frame %d", frame),
			color.Green.Sprint(draw.IndexCount),
			color.Yellow.Sprintf("0x%X", draw.UserDataHash),
			color.Green.Sprint(draw.PrimType),
			color.Green.Sprint(draw.IsIndexed),
		)
	}
	if draw.IsIndexed {
		if draw.PrimType == 19 {
			panic("Indexed QuadList drawing is not implemented")
		}
		targetBuffer, relativeOffset, err := t.GetLinearBuffer(draw.IndexBase)
		if err != nil {
			panic(err)
		}
		indexType := vk.IndexTypeUint16
		if draw.IndexType == 1 {
			indexType = vk.IndexTypeUint32
		}
		vk.CmdBindIndexBuffer(t.commandBuffer.CommandBuffer, targetBuffer, vk.DeviceSize(relativeOffset), indexType)
		vk.CmdDrawIndexed(t.commandBuffer.CommandBuffer, draw.IndexCount, draw.InstanceCount, draw.IndexOffset, 0, 0)
	} else {
		if draw.PrimType == 19 {
			quadCount := draw.IndexCount / 4
			vk.CmdBindIndexBuffer(t.commandBuffer.CommandBuffer, t.quadListIndexBuffer, 0, vk.IndexTypeUint16)
			vk.CmdDrawIndexed(t.commandBuffer.CommandBuffer, quadCount*6, draw.InstanceCount, 0, 0, 0)
		} else {
			vk.CmdDraw(t.commandBuffer.CommandBuffer, draw.IndexCount, draw.InstanceCount, 0, 0)
		}
	}

	// Mark surface as modified.
	t.activeSurface.ContentValid = true
	t.activeSurface.ImageView.Image.MarkGpuModified(t.currentGuestFrame)
}
