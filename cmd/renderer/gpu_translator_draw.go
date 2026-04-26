package renderer

import (
	"math"
	"time"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs/gpu"
	vk "github.com/goki/vulkan"
	"github.com/gookit/color"
)

var startTime time.Time

func init() {
	startTime = time.Now()
}

type StubPushConstants struct {
	UserDataAddress uint64
}

func (t *GpuTranslator) recordDraw(frame uint64, commandBuffer vk.CommandBuffer, draw *LiverpoolDrawCall) {
	rtAddress := draw.RtGpuAddress()
	t.surfacesMutex.Lock()
	surface, ok := t.surfaces[rtAddress]
	t.surfacesMutex.Unlock()
	if !ok {
		return
	}

	// Force load SPIR-V shaders.
	vsSpirv := t.GetShader(draw.VertexShader)
	if draw.EvalShader != nil {
		t.GetShader(draw.EvalShader)
	}
	if draw.HullShader != nil {
		t.GetShader(draw.HullShader)
	}
	if draw.GeometryShader != nil {
		t.GetShader(draw.GeometryShader)
	}
	psSpirv := t.GetShader(draw.PixelShader)

	// Get shader modules.
	vsModule, err := t.GetShaderModule(vsSpirv)
	if err != nil {
		return
	}
	psModule, err := t.GetShaderModule(psSpirv)
	if err != nil {
		return
	}

	// Get pipeline for defined shader modules.
	key := GpuTranslatorPipelineKey{
		VertexShaderAddress: draw.VertexShader.Address,
		PixelShaderAddress:  draw.PixelShader.Address,
		SurfaceAddress:      rtAddress,
	}
	pipeline, err := t.GetPipeline(key, vsModule, psModule, surface.renderPass, surface.Width, surface.Height)
	if err != nil {
		return
	}

	// Transition image layout on first use.
	if !surface.firstUse {
		t.imageBarrier(commandBuffer, surface.image,
			vk.ImageLayoutShaderReadOnlyOptimal, vk.ImageLayoutColorAttachmentOptimal,
			vk.AccessFlags(vk.AccessShaderReadBit), vk.AccessFlags(vk.AccessColorAttachmentWriteBit),
			vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit),
			vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit),
		)
	} else {
		surface.firstUse = false
	}

	// Derive clear color from the stub.
	clearColor := vk.ClearValue{}
	clearColor.SetColor([]float32{0.8, 0.8, 0.8, 1.0})
	vk.CmdBeginRenderPass(commandBuffer, &vk.RenderPassBeginInfo{
		SType:           vk.StructureTypeRenderPassBeginInfo,
		RenderPass:      surface.renderPass,
		Framebuffer:     surface.framebuffer,
		RenderArea:      vk.Rect2D{Extent: vk.Extent2D{Width: surface.Width, Height: surface.Height}},
		ClearValueCount: 1,
		PClearValues:    []vk.ClearValue{clearColor},
	}, vk.SubpassContentsInline)

	vk.CmdBindPipeline(commandBuffer, vk.PipelineBindPointGraphics, pipeline)
	t.setDynamicState(commandBuffer, draw, surface)

	// Get buffer addresses.
	t.userDataBuffersMutex.Lock()
	userDataBuffer := t.userDataBuffers[draw.UserDataHash]
	userDataBufferDebug := t.userDataBuffersDebug[draw.UserDataHash]
	t.userDataBuffersMutex.Unlock()

	// Push constants to shader.
	pushData := StubPushConstants{
		UserDataAddress: t.GetBufferAddress(userDataBuffer),
	}
	vk.CmdPushConstants(
		commandBuffer, t.stubPipelineLayout,
		vk.ShaderStageFlags(vk.ShaderStageVertexBit|vk.ShaderStageFragmentBit),
		0, uint32(unsafe.Sizeof(pushData)),
		unsafe.Pointer(&pushData),
	)
	t.BindTexelBuffers(commandBuffer, draw, userDataBufferDebug)

	logger.Printf("[%s] Drawing %s vertices (vertex=%s, fragment=%s, userData=%s).\n",
		color.Blue.Sprintf("Frame %d", frame),
		color.Green.Sprint(draw.VertexCount),
		color.Yellow.Sprintf("0x%X", draw.VertexShader.Address),
		color.Yellow.Sprintf("0x%X", draw.PixelShader.Address),
		color.Yellow.Sprintf("0x%X", draw.UserDataHash),
	)
	vk.CmdDraw(commandBuffer, draw.VertexCount, draw.InstanceCount, 0, 0)

	vk.CmdEndRenderPass(commandBuffer)
}

func (t *GpuTranslator) setDynamicState(commandBuffer vk.CommandBuffer, draw *LiverpoolDrawCall, surface *GpuSurface) {
	// Derive viewport from GCN scale/offset registers.
	// XScale = width/2, YScale = -height/2 (GCN NDC => screen, Y is flipped).
	vpWidth := float32(math.Abs(float64(draw.VpXScale)) * 2)
	vpHeight := float32(math.Abs(float64(draw.VpYScale)) * 2)
	vpX, vpY := draw.VpXOffset-vpWidth/2, draw.VpYOffset-vpHeight/2

	// Negative height (Vulkan's built-in Y-flip from VK_KHR_maintenance1).
	if draw.VpYScale < 0 {
		vpY = draw.VpYOffset + vpHeight/2
		vpHeight = -vpHeight
	}
	if vpWidth <= 0 || vpHeight == 0 {
		vpWidth, vpHeight = float32(surface.Width), float32(surface.Height)
		vpX, vpY = 0, 0
	}
	vk.CmdSetViewport(commandBuffer, 0, 1, []vk.Viewport{{
		X: vpX, Y: vpY,
		Width: vpWidth, Height: vpHeight,
		MinDepth: 0.0, MaxDepth: 1.0,
	}})

	scissorX, scissorY, scissorW, scissorH := draw.ScissorRect()
	if scissorW <= 0 || scissorH <= 0 {
		scissorW = int(surface.Width)
		scissorH = int(surface.Height)
		scissorX, scissorY = 0, 0
	}
	vk.CmdSetScissor(commandBuffer, 0, 1, []vk.Rect2D{{
		Offset: vk.Offset2D{X: int32(scissorX), Y: int32(scissorY)},
		Extent: vk.Extent2D{Width: uint32(scissorW), Height: uint32(scissorH)},
	}})
}
