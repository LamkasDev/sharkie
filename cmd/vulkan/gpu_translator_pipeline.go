package vulkan

import (
	"fmt"
	"math"

	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/spirv"
	"github.com/LamkasDev/sharkie/cmd/structs/gpu"
	vk "github.com/goki/vulkan"
	"github.com/gookit/color"
)

func (t *GpuTranslator) BindPipeline(frame uint64, commandBuffer vk.CommandBuffer, bind *gpu.LiverpoolBindPipeline) {
	if t.activePass != vk.NullRenderPass {
		vk.CmdEndRenderPass(commandBuffer)
		t.activePass = vk.NullRenderPass
	}
	rtAddress := uintptr(bind.RtBase) << 8
	if rtAddress == 0 {
		return
	}

	// Get or create surface.
	surface, err := t.GetSurface(SurfaceRequest{
		SurfaceKey: SurfaceKey{
			GpuAddress: rtAddress,
		},
		Format: translateColorFormat(bind.RtFormat, bind.RtNumberType, bind.RtCompSwap),
		Width:  ((bind.RtPitch & 0x7FF) + 1) * 8,
		Height: 1080,
	})
	if err != nil {
		return
	}
	t.activeSurface = surface

	// Handle depth surface.
	var depthSurface *GpuSurface
	dbAddress := uintptr(bind.DbZWriteBase) << 8
	if dbAddress != 0 {
		depthSurface, err = t.GetSurface(SurfaceRequest{
			SurfaceKey: SurfaceKey{
				GpuAddress: dbAddress,
			},
			Format: TranslateGcnDepthFormat(bind.DbZFormat),
			Width:  surface.Value.width,
			Height: surface.Value.height,
		})
		if err != nil {
			return
		}
	}

	// Get or create framebuffer.
	fbRequest := FramebufferRequest{
		ImageView: surface.Value.imageView,
		FramebufferKey: FramebufferKey{
			GpuAddress:      rtAddress,
			DepthGpuAddress: dbAddress,
			Format:          surface.Value.format,
			Width:           surface.Value.width,
			Height:          surface.Value.height,
		},
	}
	if depthSurface != nil {
		fbRequest.DepthFormat = depthSurface.Value.format
		fbRequest.DepthImageView = depthSurface.Value.imageView
	}
	fb, err := t.GetFramebuffer(fbRequest)
	if err != nil {
		return
	}

	// Get shader modules.
	vsSpirv := t.GetShader(bind.VertexShader)
	vsModule, err := t.GetShaderModule(vsSpirv)
	if err != nil {
		return
	}
	psSpirv := t.GetShaderWithContext(bind.PixelShader, spirv.SpirvShaderContext{
		PsInputAddress:  bind.PsInputAddress,
		PsInputControls: bind.PsInputControls,
	})
	psModule, err := t.GetShaderModule(psSpirv)
	if err != nil {
		return
	}

	var gsModule vk.ShaderModule
	if bind.PrimType == 17 { // RECTLIST
		gsModule, err = t.GetRectlistShader()
		if err != nil {
			return
		}
	} else if bind.GeometryShader != nil {
		gsSpirv := t.GetShader(bind.GeometryShader)
		gsModule, err = t.GetShaderModule(gsSpirv)
		if err != nil {
			return
		}
	}

	// Translate pipeline parameters.
	colorWriteMask := bind.RtTargetMask
	if (bind.RtColorControl>>4)&0x7 == 0 { // CB_DISABLE
		colorWriteMask = 0
	}

	logicOpEnable := vk.Bool32(vk.False)
	logicOp := vk.LogicOpCopy
	if (bind.RtBlendControl>>30)&1 == 0 { // Blending disabled
		rop3 := (bind.RtColorControl >> 16) & 0xFF
		if rop3 != 0xCC {
			logicOpEnable = vk.Bool32(vk.True)
			logicOp = translateLogicOp(rop3)
		}
	}

	// Get pipeline.
	pipeline, err := t.GetPipeline(GraphicsPipelineRequest{
		VertexModule:   vsModule,
		GeometryModule: gsModule,
		FragmentModule: psModule,
		RenderPass:     fb.RenderPass,
		GraphicsPipelineKey: GraphicsPipelineKey{
			VertexModuleAddress:   bind.VertexShader.Address,
			FragmentModuleAddress: bind.PixelShader.Address,
			RenderTargetAddress:   rtAddress,

			Width:    surface.Value.width,
			Height:   surface.Value.height,
			PrimType: bind.PrimType,

			CullFront:             bind.CullFront,
			CullBack:              bind.CullBack,
			Face:                  bind.Face,
			PolyMode:              bind.PolyMode,
			PolyModeFrontPtype:    bind.PolyModeFrontPtype,
			PolyModeBackPtype:     bind.PolyModeBackPtype,
			PolyOffsetFrontEnable: bind.PolyOffsetFrontEnable,
			PolyOffsetBackEnable:  bind.PolyOffsetBackEnable,
			PolyOffsetParaEnable:  bind.PolyOffsetParaEnable,
			ProvokingVertexLast:   bind.ProvokingVertexLast,

			BlendAttachment:   translateBlendControl(bind.RtBlendControl, colorWriteMask, bind.RtBlendBypass),
			DepthStencilState: translateDepthControl(bind.DbDepthControl, bind.DbStencilControl, bind.DbStencilRefMask, bind.DbStencilRefMaskBf),
			LogicOpEnable:     logicOpEnable,
			LogicOp:           logicOp,

			DbKillEnable:           bind.DbKillEnable,
			DbCoverageToMaskEnable: bind.DbCoverageToMaskEnable,
			DbAlphaToMaskDisable:   bind.DbAlphaToMaskDisable,

			VpScissorEnable:    bind.VpScissorEnable,
			WindowOffsetEnable: bind.WindowOffsetEnable,

			LineStippleEnable: bind.LineStippleEnable,

			MsaaEnable:          bind.MsaaEnable,
			MsaaSampleLocations: bind.MsaaSampleLocations,

			MultiPrimIbResetEnable: bind.MultiPrimIbResetEnable,
		},
	})
	if err != nil {
		return
	}

	// Select render pass and clear on first use in frame or if explicitly requested.
	var clearValues []vk.ClearValue
	renderPass := fb.RenderPassNoClear
	if surface.FrameUsed < frame {
		renderPass = fb.RenderPass
		clearColor := vk.ClearValue{}
		clearColorFloat := translateClearColor(bind.RtClearWord0, bind.RtClearWord1, bind.RtFormat, bind.RtNumberType, bind.RtCompSwap)
		clearColor.SetColor(clearColorFloat)
		clearValues = append(clearValues, clearColor)
		surface.FrameUsed = frame
		if depthSurface != nil {
			clearDepth := vk.ClearValue{}
			clearDepth.SetDepthStencil(math.Float32frombits(bind.DbDepthClearValue), bind.DbStencilClearValue)
			clearValues = append(clearValues, clearDepth)
			depthSurface.FrameUsed = frame
		}
	}

	// Start render pass.
	vk.CmdBeginRenderPass(commandBuffer, &vk.RenderPassBeginInfo{
		SType:           vk.StructureTypeRenderPassBeginInfo,
		RenderPass:      renderPass,
		Framebuffer:     fb.Framebuffer,
		RenderArea:      vk.Rect2D{Extent: vk.Extent2D{Width: surface.Value.width, Height: surface.Value.height}},
		ClearValueCount: uint32(len(clearValues)),
		PClearValues:    clearValues,
	}, vk.SubpassContentsInline)

	// Bind pipeline.
	vk.CmdBindPipeline(commandBuffer, vk.PipelineBindPointGraphics, pipeline)
	t.activePass = fb.RenderPass
	t.activeFramebuffer = fb.Framebuffer
	t.activePipeline = pipeline

	if logger.LogRenderer {
		logger.Printf("[%s] Bound pipeline (vertex=%s, fragment=%s).\n",
			color.Blue.Sprintf("Frame %d", frame),
			color.Yellow.Sprintf("0x%X", bind.VertexShader.Address),
			color.Yellow.Sprintf("0x%X", bind.PixelShader.Address),
		)
	}
}

func (t *GpuTranslator) EndRenderPass(commandBuffer vk.CommandBuffer) {
	vk.CmdEndRenderPass(commandBuffer)
	t.activePass = vk.NullRenderPass
}

func (t *GpuTranslator) GetPipeline(request GraphicsPipelineRequest) (vk.Pipeline, error) {
	// Get already created pipeline.
	t.pipelinesMutex.Lock()
	pipeline, ok := t.pipelines[request.GraphicsPipelineKey]
	t.pipelinesMutex.Unlock()
	if ok {
		return pipeline, nil
	}

	// Create the pipeline.
	pipeline, err := t.createGraphicsPipeline(request)
	if err != nil {
		return vk.NullPipeline, fmt.Errorf("createGraphicsPipeline: %w", err)
	}
	t.pipelinesMutex.Lock()
	t.pipelines[request.GraphicsPipelineKey] = pipeline
	t.pipelinesMutex.Unlock()

	return pipeline, nil
}

func (t *GpuTranslator) GetComputePipeline(request ComputePipelineRequest) (vk.Pipeline, error) {
	// Get already created pipeline.
	t.computePipelinesMutex.Lock()
	pipeline, ok := t.computePipelines[request.ComputePipelineKey]
	t.computePipelinesMutex.Unlock()
	if ok {
		return pipeline, nil
	}

	// Create the pipeline.
	pipeline, err := t.createComputePipeline(request)
	if err != nil {
		return vk.NullPipeline, fmt.Errorf("createComputePipeline: %w", err)
	}
	t.computePipelinesMutex.Lock()
	t.computePipelines[request.ComputePipelineKey] = pipeline
	t.computePipelinesMutex.Unlock()

	return pipeline, nil
}
