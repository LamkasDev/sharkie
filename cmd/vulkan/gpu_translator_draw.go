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

func (t *GpuTranslator) recordDraw(frame uint64, commandBuffer vk.CommandBuffer, draw *gpu.LiverpoolDrawCall) {
	rtAddress := draw.RtGpuAddress()
	if rtAddress == 0 {
		return
	}

	// Get or create surface.
	width := draw.RtPitchPixels()
	height := draw.VpHeight()
	if height == 0 {
		height = 1080 // Default to something if viewport height is 0
	}
	surface, err := t.GetSurface(SurfaceRequest{
		SurfaceKey: SurfaceKey{
			GpuAddress: rtAddress,
		},
		Format: translateColorFormat(draw.RtFormat, draw.RtNumberType, draw.RtCompSwap),
		Width:  width,
		Height: height,
	})
	if err != nil {
		return
	}

	// Handle depth surface.
	var depthSurface *GpuSurface
	depthFormat := vk.FormatUndefined
	/* depthFormat := TranslateGcnDepthFormat(draw.DbZFormat)
	if depthFormat != vk.FormatUndefined && draw.DbZWriteBase != 0 {
		depthSurface, err = t.GetSurface(SurfaceRequest{
			SurfaceKey: SurfaceKey{
				GpuAddress: uintptr(draw.DbZWriteBase) << 8,
			},
			Format: depthFormat,
			Width:  width,
			Height: height,
		})
		if err != nil {
			return
		}
	} */

	// Get or create framebuffer.
	fbRequest := FramebufferRequest{
		ImageView: surface.Value.imageView,
		FramebufferKey: FramebufferKey{
			GpuAddress:  rtAddress,
			Format:      surface.Value.format,
			DepthFormat: depthFormat,
			Width:       surface.Value.width,
			Height:      surface.Value.height,
		},
	}
	if depthSurface != nil {
		fbRequest.DepthImageView = depthSurface.Value.imageView
	}
	fb, err := t.GetFramebuffer(fbRequest)
	if err != nil {
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

	var gsModule vk.ShaderModule
	if draw.PrimType == 17 { // RECTLIST
		gsModule, err = t.GetRectlistShader()
		if err != nil {
			return
		}
	} else if draw.GeometryShader != nil {
		gsSpirv := t.GetShader(draw.GeometryShader)
		gsModule, err = t.GetShaderModule(gsSpirv)
		if err != nil {
			return
		}
	}

	// Get pipeline for shader modules and parameters.
	colorWriteMask := draw.RtTargetMask
	if (draw.RtColorControl>>4)&0x7 == 0 { // CB_DISABLE
		colorWriteMask = 0
	}

	logicOpEnable := vk.Bool32(vk.False)
	logicOp := vk.LogicOpCopy
	if (draw.RtBlendControl>>30)&1 == 0 { // Blending disabled
		rop3 := (draw.RtColorControl >> 16) & 0xFF
		if rop3 != 0xCC {
			logicOpEnable = vk.Bool32(vk.True)
			logicOp = translateLogicOp(rop3)
		}
	}

	pipeline, err := t.GetPipeline(GraphicsPipelineRequest{
		VertexModule:   vsModule,
		GeometryModule: gsModule,
		FragmentModule: psModule,
		RenderPass:     fb.RenderPass,
		GraphicsPipelineKey: GraphicsPipelineKey{
			VertexModuleAddress:   draw.VertexShader.Address,
			FragmentModuleAddress: draw.PixelShader.Address,
			RenderTargetAddress:   rtAddress,
			RenderPass:            fb.RenderPass,

			Width:             surface.Value.width,
			Height:            surface.Value.height,
			PrimType:          draw.PrimType,
			BlendAttachment:   translateBlendControl(draw.RtBlendControl, colorWriteMask, draw.RtBlendBypass),
			DepthStencilState: translateDepthControl(draw.DbDepthControl, draw.DbStencilControl),
			LogicOpEnable:     logicOpEnable,
			LogicOp:           logicOp,

			DbKillEnable:           draw.DbKillEnable,
			DbCoverageToMaskEnable: draw.DbCoverageToMaskEnable,
			DbAlphaToMaskDisable:   draw.DbAlphaToMaskDisable,
		},
	})
	if err != nil {
		return
	}

	// Select render pass and clear on first use in frame or if explicitly requested.
	var renderPass vk.RenderPass
	var clearValues []vk.ClearValue
	shouldClear := surface.FrameUsed < frame
	if depthSurface != nil && depthSurface.FrameUsed < frame {
		shouldClear = true
	}
	/* if draw.DbDepthClearEnable || draw.DbStencilClearEnable {
		shouldClear = true
	} */
	if shouldClear {
		renderPass = fb.RenderPass
		clearColor := vk.ClearValue{}
		clearColor.SetColor([]float32{0.8, 0.8, 0.8, 1.0})
		clearValues = []vk.ClearValue{clearColor}
		if depthSurface != nil {
			clearDepth := vk.ClearValue{}
			clearDepth.SetDepthStencil(math.Float32frombits(draw.DbDepthClearValue), draw.DbStencilClearValue)
			clearValues = append(clearValues, clearDepth)
		}

		surface.FrameUsed = frame
		if depthSurface != nil {
			depthSurface.FrameUsed = frame
		}

		// Transition surface to General if it's the first use.
		if surface.FirstUse {
			t.imageBarrier(commandBuffer, surface.Value.image,
				vk.ImageLayoutUndefined, vk.ImageLayoutGeneral,
				0, vk.AccessFlags(vk.AccessColorAttachmentWriteBit|vk.AccessShaderReadBit|vk.AccessShaderWriteBit),
				vk.PipelineStageFlags(vk.PipelineStageTopOfPipeBit), vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit|vk.PipelineStageAllGraphicsBit|vk.PipelineStageComputeShaderBit))
			surface.FirstUse = false
		}
		if depthSurface != nil && depthSurface.FirstUse {
			t.imageBarrier(commandBuffer, depthSurface.Value.image,
				vk.ImageLayoutUndefined, vk.ImageLayoutGeneral,
				0, vk.AccessFlags(vk.AccessDepthStencilAttachmentWriteBit|vk.AccessShaderReadBit|vk.AccessShaderWriteBit),
				vk.PipelineStageFlags(vk.PipelineStageTopOfPipeBit), vk.PipelineStageFlags(vk.PipelineStageEarlyFragmentTestsBit|vk.PipelineStageLateFragmentTestsBit|vk.PipelineStageAllGraphicsBit|vk.PipelineStageComputeShaderBit))
			depthSurface.FirstUse = false
		}
	} else {
		renderPass = fb.RenderPassNoClear
	}

	vk.CmdBeginRenderPass(commandBuffer, &vk.RenderPassBeginInfo{
		SType:           vk.StructureTypeRenderPassBeginInfo,
		RenderPass:      renderPass,
		Framebuffer:     fb.Framebuffer,
		RenderArea:      vk.Rect2D{Extent: vk.Extent2D{Width: surface.Value.width, Height: surface.Value.height}},
		ClearValueCount: uint32(len(clearValues)),
		PClearValues:    clearValues,
	}, vk.SubpassContentsInline)

	// Bind pipeline and setup scissor/viewport.
	vk.CmdBindPipeline(commandBuffer, vk.PipelineBindPointGraphics, pipeline)
	t.setDynamicState(commandBuffer, draw, surface)

	// Bind bindless descriptor set.
	vk.CmdBindDescriptorSets(commandBuffer, vk.PipelineBindPointGraphics, t.pipelineLayout, spirvStructs.DescriptorSetSlotBindless, 1, []vk.DescriptorSet{t.bindlessDescriptorSet}, 0, nil)

	// Bind discovery descriptor set.
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
		VteControl:              draw.VteControl,
	}
	vk.CmdPushConstants(
		commandBuffer, t.pipelineLayout,
		vk.ShaderStageFlags(vk.ShaderStageVertexBit|vk.ShaderStageComputeBit),
		0, spirvStructs.PushConstantsSize,
		unsafe.Pointer(&pushDataVs),
	)

	// Push constants to fragment shader.
	pushDataFs := spirvStructs.PushConstants{
		UserDataAddress:         t.userDataBufferAddress + uint64(userDataOffset),
		OnionMemoryBaseAddress:  GlobalAllocator.DeviceAddress,
		GarlicMemoryBaseAddress: GlobalGpuAllocator.DeviceAddress,
		UserSgprCount:           gpu.DecodeUserSgprCount(draw.PixelShRsrc2),
		ShaderRsrc2:             draw.PixelShRsrc2,
		VteControl:              draw.VteControl,
	}
	vk.CmdPushConstants(
		commandBuffer, t.pipelineLayout,
		vk.ShaderStageFlags(vk.ShaderStageFragmentBit),
		spirvStructs.PushConstantsSize, spirvStructs.PushConstantsSize,
		unsafe.Pointer(&pushDataFs),
	)

	// Draw.
	logger.Printf("[%s] Drawing %s vertices (vertex=%s, fragment=%s, userData=%s, topology=%s, indexed=%s).\n",
		color.Blue.Sprintf("Frame %d", frame),
		color.Green.Sprint(draw.VertexCount),
		color.Yellow.Sprintf("0x%X", draw.VertexShader.Address),
		color.Yellow.Sprintf("0x%X", draw.PixelShader.Address),
		color.Yellow.Sprintf("0x%X", draw.UserDataHash),
		color.Green.Sprint(draw.PrimType),
		color.Green.Sprint(nstd.Btoi(draw.IsIndexed)),
	)
	if draw.IsIndexed {
		targetBuffer, relativeOffset, err := t.GetBufferFromAddress(draw.IndexBaseAddress)
		if err != nil {
			panic(err)
		}

		if targetBuffer != vk.NullBuffer {
			indexType := vk.IndexTypeUint16
			if draw.IndexType == 1 {
				indexType = vk.IndexTypeUint32
			}
			vk.CmdBindIndexBuffer(commandBuffer, targetBuffer, vk.DeviceSize(relativeOffset), indexType)
			vk.CmdDrawIndexed(commandBuffer, draw.IndexCount, draw.InstanceCount, 0, int32(draw.BaseVertexOffset), 0)
		} else {
			vk.CmdDraw(commandBuffer, draw.VertexCount, draw.InstanceCount, 0, 0)
		}
	} else {
		vk.CmdDraw(commandBuffer, draw.VertexCount, draw.InstanceCount, 0, 0)
	}

	vk.CmdEndRenderPass(commandBuffer)
}

func (t *GpuTranslator) setDynamicState(commandBuffer vk.CommandBuffer, draw *gpu.LiverpoolDrawCall, surface *GpuSurface) {
	// Derive viewport from GCN scale/offset/control registers.
	vpxScale := draw.VpXScale
	if !draw.VpXScaleEnable {
		vpxScale = 1.0
	}
	vpxOffset := draw.VpXOffset
	if !draw.VpXOffsetEnable {
		vpxOffset = 0.0
	}
	vpyScale := draw.VpYScale
	if !draw.VpYScaleEnable {
		vpyScale = 1.0
	}
	vpyOffset := draw.VpYOffset
	if !draw.VpYOffsetEnable {
		vpyOffset = 0.0
	}
	vpZScale := draw.VpZScale
	if !draw.VpZScaleEnable {
		vpZScale = 1.0
	}
	vpZOffset := draw.VpZOffset
	if !draw.VpZOffsetEnable {
		vpZOffset = 0.0
	}

	// Process viewport transforms.
	vpWidth := vpxScale * 2
	vpHeight := vpyScale * 2
	vpX, vpY := vpxOffset-vpxScale, vpyOffset-vpyScale
	if vpWidth == 0 || vpHeight == 0 {
		vpWidth, vpHeight = float32(surface.Value.width), float32(surface.Value.height)
		vpX, vpY = 0, 0
	}
	vk.CmdSetViewport(commandBuffer, 0, 1, []vk.Viewport{{
		X: vpX, Y: vpY,
		Width: vpWidth, Height: vpHeight,
		MinDepth: vpZOffset,
		MaxDepth: vpZOffset + vpZScale,
	}})

	// Setup the scissor from TL/BR registers.
	scissorX, scissorY, scissorW, scissorH := draw.ScissorRect()
	if scissorW <= 0 || scissorH <= 0 {
		scissorW = int(surface.Value.width)
		scissorH = int(surface.Value.height)
		scissorX, scissorY = 0, 0
	}
	vk.CmdSetScissor(commandBuffer, 0, 1, []vk.Rect2D{{
		Offset: vk.Offset2D{X: int32(scissorX), Y: int32(scissorY)},
		Extent: vk.Extent2D{Width: uint32(scissorW), Height: uint32(scissorH)},
	}})

	// Setup blend constants.
	vk.CmdSetBlendConstants(commandBuffer, &[4]float32{
		math.Float32frombits(draw.BlendRed),
		math.Float32frombits(draw.BlendGreen),
		math.Float32frombits(draw.BlendBlue),
		math.Float32frombits(draw.BlendAlpha),
	})
}
