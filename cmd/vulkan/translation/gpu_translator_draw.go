package translation

import (
	"math"
	"time"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/lib_structs/gpu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/spirv"
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
	userData, _ := gpu.GlobalUserDataSnapshots[draw.UserDataHash]
	t.userDataBuffersMutex.Unlock()

	// Get shaders.
	shaders := []*spirv.SpirvShader{t.activeVertexShader}
	if t.activeFragmentShader != nil {
		shaders = append(shaders, t.activeFragmentShader)
	}

	// Bind resources.
	staticSetToBind := t.staticDescriptorPool.DefaultSet(frame)
	_, _, activeStaticSet, err := t.BindResources(shaders, userData)
	if err != nil {
		panic(err)
	}
	if activeStaticSet != vk.NullDescriptorSet {
		staticSetToBind = activeStaticSet
	}

	// Resume render pass if needed.
	t.ResumeActiveRenderPass()
	vk.CmdBindDescriptorSets(
		t.commandBuffer.CommandBuffer, vk.PipelineBindPointGraphics,
		t.pipelineLayout, spirvStructs.DescriptorSetSlotStatic,
		1, []vk.DescriptorSet{staticSetToBind},
		0, nil,
	)

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

func (t *GpuTranslator) SetDynamicState(dynamicState *gpu.LiverpoolSetDynamicState) {
	t.activeVteControl = uint32(dynamicState.PaClVteCntl)
	t.activeClipControl = uint32(dynamicState.ClipControl)
	t.activeDynamicState = dynamicState

	t.setViewport(dynamicState)
	t.setScissor(dynamicState)

	// Setup blend constants.
	vk.CmdSetBlendConstants(t.commandBuffer.CommandBuffer, &[4]float32{
		math.Float32frombits(dynamicState.BlendRed),
		math.Float32frombits(dynamicState.BlendGreen),
		math.Float32frombits(dynamicState.BlendBlue),
		math.Float32frombits(dynamicState.BlendAlpha),
	})
}

func (t *GpuTranslator) setViewport(dynamicState *gpu.LiverpoolSetDynamicState) {
	// Derive viewport from GCN scale/offset/control registers.
	vpxScale := dynamicState.VpXScale
	if !dynamicState.PaClVteCntl.VpXScaleEnable() {
		vpxScale = 1.0
	}
	vpxOffset := dynamicState.VpXOffset
	if !dynamicState.PaClVteCntl.VpXOffsetEnable() {
		vpxOffset = 0.0
	}
	vpyScale := dynamicState.VpYScale
	if !dynamicState.PaClVteCntl.VpYScaleEnable() {
		vpyScale = 1.0
	}
	vpyOffset := dynamicState.VpYOffset
	if !dynamicState.PaClVteCntl.VpYOffsetEnable() {
		vpyOffset = 0.0
	}
	vpzScale := dynamicState.VpZScale
	if !dynamicState.PaClVteCntl.VpZScaleEnable() {
		vpzScale = 1.0
	}
	vpzOffset := dynamicState.VpZOffset
	if !dynamicState.PaClVteCntl.VpZOffsetEnable() {
		vpzOffset = 0.0
	}
	windowOffsetX := int32(int16(dynamicState.WindowOffset.WindowXOffset()))
	windowOffsetY := int32(int16(dynamicState.WindowOffset.WindowYOffset()))
	// hwOffsetX := float32(int32(int16(dynamicState.HardwareScreenOffset & 0xFFFF)))
	// hwOffsetY := float32(int32(int16((dynamicState.HardwareScreenOffset >> 16) & 0xFFFF)))

	// Process viewport transforms.
	vpWidth := vpxScale * 2
	vpHeight := vpyScale * 2
	vpX, vpY := vpxOffset-vpxScale, vpyOffset-vpyScale
	if dynamicState.PaSuScModeCntl.WindowOffsetEnable() {
		vpX += float32(windowOffsetX)
		vpY += float32(windowOffsetY)
	}
	// vpX += hwOffsetX
	// vpY += hwOffsetY

	// Apply fallback if zero sized.
	if vpWidth == 0 || vpHeight == 0 {
		vpWidth, vpHeight = float32(t.activeSurface.ImageView.Image.FirstDescriptor.Width), float32(t.activeSurface.ImageView.Image.FirstDescriptor.Height)
		vpX, vpY = 0, 0
		if dynamicState.PaSuScModeCntl.WindowOffsetEnable() {
			vpX, vpY = float32(windowOffsetX), float32(windowOffsetY)
		}
	}
	minDepth := max(0.0, min(1.0, vpzOffset))
	maxDepth := max(0.0, min(1.0, vpzOffset+vpzScale))

	vk.CmdSetViewport(t.commandBuffer.CommandBuffer, 0, 1, []vk.Viewport{{
		X: vpX, Y: vpY,
		Width: vpWidth, Height: vpHeight,
		MinDepth: minDepth,
		MaxDepth: maxDepth,
	}})
}

type ScissorRect struct {
	X1, Y1, X2, Y2 int32
}

func (s ScissorRect) Intersect(other ScissorRect) ScissorRect {
	return ScissorRect{
		X1: max(s.X1, other.X1),
		Y1: max(s.Y1, other.Y1),
		X2: min(s.X2, other.X2),
		Y2: min(s.Y2, other.Y2),
	}
}

func (t *GpuTranslator) setScissor(dynamicState *gpu.LiverpoolSetDynamicState) {
	windowOffsetX := int32(int16(dynamicState.WindowOffset.WindowXOffset()))
	windowOffsetY := int32(int16(dynamicState.WindowOffset.WindowYOffset()))

	// Helper to decode a GCN scissor register.
	decodeScissor := func(tl uint32, br uint32) ScissorRect {
		windowOffsetDisable := (tl >> 31) & 1
		x1 := int32(tl & 0x7FFF)
		y1 := int32((tl >> 16) & 0x7FFF)
		x2 := int32(br & 0x7FFF)
		y2 := int32((br >> 16) & 0x7FFF)
		if windowOffsetDisable == 0 {
			x1 += windowOffsetX
			y1 += windowOffsetY
			x2 += windowOffsetX
			y2 += windowOffsetY
		}

		return ScissorRect{X1: x1, Y1: y1, X2: x2, Y2: y2}
	}

	// Apply screen scissor (no offset).
	screenScissor := ScissorRect{
		X1: int32(int16(dynamicState.ScissorTl.TlX())),
		Y1: int32(int16(dynamicState.ScissorTl.TlY())),
		X2: int32(int16(dynamicState.ScissorBr & 0xFFFF)),
		Y2: int32(int16((dynamicState.ScissorBr >> 16) & 0xFFFF)),
	}
	finalScissor := screenScissor

	// Apply window scissor.
	if dynamicState.WindowScissorTl != 0 || dynamicState.WindowScissorBr != 0 {
		windowScissor := decodeScissor(uint32(dynamicState.WindowScissorTl), dynamicState.WindowScissorBr)
		finalScissor = screenScissor.Intersect(windowScissor)
	}

	// Apply optional viewport scissor.
	if dynamicState.PaScModeCntl0.VpScissorEnable() {
		vpScissor := decodeScissor(uint32(dynamicState.VpScissorTl), dynamicState.VpScissorBr)
		finalScissor = finalScissor.Intersect(vpScissor)
	}

	// Apply generic scissor.
	if dynamicState.GenericScissorTl != 0 || dynamicState.GenericScissorBr != 0 {
		genericScissor := decodeScissor(uint32(dynamicState.GenericScissorTl), dynamicState.GenericScissorBr)
		finalScissor = finalScissor.Intersect(genericScissor)
	}

	// Calculate width and height.
	width := uint32(max(0, finalScissor.X2-finalScissor.X1))
	height := uint32(max(0, finalScissor.Y2-finalScissor.Y1))
	if width == 0 || height == 0 {
		width = uint32(t.activeSurface.ImageView.Image.FirstDescriptor.Width)
		height = uint32(t.activeSurface.ImageView.Image.FirstDescriptor.Height)
		finalScissor.X1 = 0
		finalScissor.Y1 = 0
	}
	vk.CmdSetScissor(t.commandBuffer.CommandBuffer, 0, 1, []vk.Rect2D{{
		Offset: vk.Offset2D{X: finalScissor.X1, Y: finalScissor.Y1},
		Extent: vk.Extent2D{Width: width, Height: height},
	}})
}
