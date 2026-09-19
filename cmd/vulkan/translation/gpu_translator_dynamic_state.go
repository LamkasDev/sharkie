package translation

import (
	"math"

	"github.com/LamkasDev/sharkie/cmd/lib_structs/gpu"
	vk "github.com/goki/vulkan"
)

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
	isClipDisabled := dynamicState.ClipControl.ClipDisable() || t.activePrimType == 17
	var vpX, vpY, vpWidth, vpHeight float32
	if isClipDisabled {
		vpX = 0.0
		vpY = 0.0
		vpWidth = 16384.0
		vpHeight = 16384.0
	} else {
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

		// Process viewport transforms.
		vpWidth = vpxScale * 2
		vpHeight = vpyScale * 2
		vpX, vpY = vpxOffset-vpxScale, vpyOffset-vpyScale

		// Apply fallback if zero sized.
		if vpWidth == 0 || vpHeight == 0 {
			if t.activeSurface != nil {
				vpWidth, vpHeight = float32(t.activeSurface.ImageView.Image.FirstDescriptor.Width), float32(t.activeSurface.ImageView.Image.FirstDescriptor.Height)
			} else if t.activeDepthSurface != nil {
				vpWidth, vpHeight = float32(t.activeDepthSurface.ImageView.Image.FirstDescriptor.Width), float32(t.activeDepthSurface.ImageView.Image.FirstDescriptor.Height)
			}
			vpX, vpY = 0, 0
		}
	}
	vpzScale := dynamicState.VpZScale
	if !dynamicState.PaClVteCntl.VpZScaleEnable() {
		vpzScale = 1.0
	}
	vpzOffset := dynamicState.VpZOffset
	if !dynamicState.PaClVteCntl.VpZOffsetEnable() {
		vpzOffset = 0.0
	}

	var minDepth, maxDepth float32
	if !dynamicState.ClipControl.DxClipSpaceDef() {
		// When clip space is [-W, W] (MinusWToW), zoffset is centered.
		minDepth = vpzOffset - vpzScale
		maxDepth = vpzOffset + vpzScale
	} else {
		minDepth = vpzOffset
		maxDepth = vpzOffset + vpzScale
	}
	minDepth = max(0.0, min(1.0, minDepth))
	maxDepth = max(0.0, min(1.0, maxDepth))

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

	// Clamp scissor coordinates to non-negative screen space, matching Vulkan requirements and shadps4.
	var surfaceWidth, surfaceHeight int32 = 16384, 16384
	if t.activeSurface != nil {
		surfaceWidth = int32(t.activeSurface.ImageView.Image.FirstDescriptor.Width)
		surfaceHeight = int32(t.activeSurface.ImageView.Image.FirstDescriptor.Height)
	} else if t.activeDepthSurface != nil {
		surfaceWidth = int32(t.activeDepthSurface.ImageView.Image.FirstDescriptor.Width)
		surfaceHeight = int32(t.activeDepthSurface.ImageView.Image.FirstDescriptor.Height)
	}

	clampedX1 := max(0, finalScissor.X1)
	clampedY1 := max(0, finalScissor.Y1)
	clampedX2 := min(surfaceWidth, max(0, finalScissor.X2))
	clampedY2 := min(surfaceHeight, max(0, finalScissor.Y2))

	width := uint32(max(0, clampedX2-clampedX1))
	height := uint32(max(0, clampedY2-clampedY1))
	if width == 0 || height == 0 {
		width = uint32(surfaceWidth)
		height = uint32(surfaceHeight)
		clampedX1 = 0
		clampedY1 = 0
	}
	vk.CmdSetScissor(t.commandBuffer.CommandBuffer, 0, 1, []vk.Rect2D{{
		Offset: vk.Offset2D{X: clampedX1, Y: clampedY1},
		Extent: vk.Extent2D{Width: width, Height: height},
	}})
}
