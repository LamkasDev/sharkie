package translation

import (
	"fmt"
	"math"

	"github.com/LamkasDev/sharkie/cmd/lib_structs/gpu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/spirv"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan"
	vk "github.com/goki/vulkan"
	"github.com/gookit/color"
)

func (t *GpuTranslator) BindPipeline(frame uint64, bind *gpu.LiverpoolBindPipeline) {
	t.EndRenderPass()
	rtAddress := spirvStructs.GetPhysicalGpuAddress(uintptr(bind.RtBase) << 8)
	if rtAddress == 0 {
		return
	}

	rtWidth := ((bind.RtPitch & 0x7FF) + 1) * 8
	rtHeight := colorBufferHeight(bind.RtPitch, bind.RtSlice)

	// Get or create surface.
	surface, err := t.GetSurface(spirvStructs.ImageDescriptor{
		BaseAddress: rtAddress,
		Width:       uint16(rtWidth), Height: uint16(rtHeight),
		DataFormat: 10, NumFormat: 0,
		DstSelX: 4, DstSelY: 5, DstSelZ: 6, DstSelW: 7,
		Depth: 1, Pitch: uint16(rtWidth),
	}, translateColorFormat(bind.RtFormat, bind.RtNumberType, bind.RtCompSwap))
	if err != nil {
		return
	}
	t.activeSurface = surface

	// The game keeps one depth buffer (DB_Z_WRITE_BASE) while ping-ponging CB_COLOR0.
	// Stale depth from a previous color target corrupts compositing on the next one.
	if t.lastColorRtAddress != 0 && t.lastColorRtAddress != rtAddress {
		t.surfacesMutex.Lock()
		depthSurface := t.surfaces[spirvStructs.GetPhysicalGpuAddress(uintptr(bind.DbZWriteBase)<<8)]
		t.surfacesMutex.Unlock()
		if depthSurface != nil {
			depthSurface.FrameUsed = 0
		}
	}
	t.lastColorRtAddress = rtAddress

	// Handle depth surface.
	var depthSurface *vulkan.VulkanSurface
	dbAddress := spirvStructs.GetPhysicalGpuAddress(uintptr(bind.DbZWriteBase) << 8)
	if dbAddress != 0 {
		depthSurface, err = t.GetSurface(spirvStructs.ImageDescriptor{
			BaseAddress: dbAddress,
			Width:       uint16(surface.ImageView.Image.FirstDescriptor.Width),
			Height:      uint16(surface.ImageView.Image.FirstDescriptor.Height),
			DataFormat:  10, NumFormat: 0,
			DstSelX: 4, DstSelY: 5, DstSelZ: 6, DstSelW: 7,
			Depth: 1, Pitch: surface.ImageView.Image.FirstDescriptor.Width,
		}, t.TranslateGcnDepthFormat(bind.DbZFormat))
		if err != nil {
			return
		}
	}

	// Get or create framebuffer.
	fbRequest := vulkan.FramebufferRequest{
		ImageView: surface.ImageView,
		FramebufferKey: vulkan.FramebufferKey{
			GpuAddress:      rtAddress,
			DepthGpuAddress: dbAddress,
			Format:          surface.ImageView.Image.ImageFormat,
			Width:           uint32(surface.ImageView.Image.FirstDescriptor.Width),
			Height:          uint32(surface.ImageView.Image.FirstDescriptor.Height),
		},
	}
	if depthSurface != nil {
		fbRequest.DepthFormat = depthSurface.ImageView.Image.ImageFormat
		fbRequest.DepthImageView = depthSurface.ImageView
	}
	fb, err := t.GetFramebuffer(fbRequest)
	if err != nil {
		return
	}

	// Get shader modules.
	vsSpirv := t.GetShader(bind.VertexShader)
	t.activeVertexShader = vsSpirv
	vsModule, err := t.GetShaderModule(vsSpirv)
	if err != nil {
		return
	}
	psSpirv := t.GetShaderWithContext(bind.PixelShader, spirv.SpirvShaderContext{
		PsInControl:     bind.PsInControl,
		PsInputAddress:  bind.PsInputAddress,
		PsInputControls: bind.PsInputControls,
	})
	t.activeFragmentShader = psSpirv
	psModule, err := t.GetShaderModule(psSpirv)
	if err != nil {
		return
	}

	var tcsModule, tesModule, gsModule vk.ShaderModule
	if bind.PrimType == 17 { // RECTLIST
		tcsModule, err = t.GetRectlistTcsShader()
		if err != nil {
			return
		}
		tesModule, err = t.GetRectlistTesShader()
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
	pipeline, err := t.GetPipeline(vulkan.GraphicsPipelineRequest{
		VertexModule:      vsModule,
		TessControlModule: tcsModule,
		TessEvalModule:    tesModule,
		GeometryModule:    gsModule,
		FragmentModule:    psModule,
		RenderPass:        fb.RenderPass,
		GraphicsPipelineKey: vulkan.GraphicsPipelineKey{
			VertexModuleAddress:   bind.VertexShader.Address,
			FragmentModuleAddress: bind.PixelShader.Address,
			RenderTargetAddress:   rtAddress,
			DepthTargetAddress:    dbAddress,

			Width:    uint32(surface.ImageView.Image.FirstDescriptor.Width),
			Height:   uint32(surface.ImageView.Image.FirstDescriptor.Height),
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
	clearColor := vk.ClearValue{}
	clearColorFloat := translateClearColor(bind.RtClearWord0, bind.RtClearWord1, bind.RtFormat, bind.RtNumberType, bind.RtCompSwap)
	clearColor.SetColor(clearColorFloat)
	clearValues = append(clearValues, clearColor)
	if depthSurface != nil {
		clearDepth := vk.ClearValue{}
		clearDepth.SetDepthStencil(math.Float32frombits(bind.DbDepthClearValue), bind.DbStencilClearValue)
		clearValues = append(clearValues, clearDepth)
	}

	// Clear on first use within this guest frame.
	renderPass := fb.RenderPassNoClear
	clearColorNeeded := surface.FrameUsed < frame
	clearDepthNeeded := depthSurface != nil && depthSurface.FrameUsed < frame
	if clearColorNeeded && clearDepthNeeded {
		renderPass = fb.RenderPass
		surface.FrameUsed = frame
		depthSurface.FrameUsed = frame
	} else if clearColorNeeded && !clearDepthNeeded {
		if depthSurface != nil {
			renderPass = fb.RenderPassClearColorLoadDepth
		} else {
			renderPass = fb.RenderPass
		}
		surface.FrameUsed = frame
	} else if !clearColorNeeded && clearDepthNeeded {
		renderPass = fb.RenderPassLoadColorClearDepth
		depthSurface.FrameUsed = frame
	}
	t.EndRenderPass()
	surface.ImageView.Image.BarrierColorAttachment(t.commandBuffer)

	if renderPass == fb.RenderPassNoClear {
		clearValues = nil
	} else if renderPass == fb.RenderPassClearColorLoadDepth {
		clearValues = clearValues[:1]
	} else if renderPass == fb.RenderPassLoadColorClearDepth {
		// Needs both elements because depth is index 1.
		// Color clear value is ignored since loadOp is LOAD, but must be present.
	}

	t.StartRenderPass(renderPass, fb.RenderPassNoClear, fb.Framebuffer, pipeline, clearValues,
		uint32(surface.ImageView.Image.FirstDescriptor.Width),
		uint32(surface.ImageView.Image.FirstDescriptor.Height), rtAddress)

	if logger.LogRenderer {
		logger.Printf("[%s] Bound pipeline (vertex=%s, fragment=%s, rtAddress=0x%X, rtPitch=%d, rtSize=%dx%d).\n",
			color.Blue.Sprintf("Frame %d", frame),
			color.Yellow.Sprintf("0x%X", bind.VertexShader.Address),
			color.Yellow.Sprintf("0x%X", bind.PixelShader.Address),
			rtAddress, bind.RtPitch, rtWidth, rtHeight,
		)
	}
}

func (t *GpuTranslator) GetPipeline(request vulkan.GraphicsPipelineRequest) (vk.Pipeline, error) {
	// Get already created pipeline.
	t.pipelinesMutex.Lock()
	pipeline, ok := t.pipelines[request.GraphicsPipelineKey]
	t.pipelinesMutex.Unlock()
	if ok {
		return pipeline, nil
	}

	// Create the pipeline.
	pipeline, err := vulkan.CreateGraphicsPipeline(&t.handles, request, request.RenderPass, t.pipelineLayout, vk.NullPipelineCache)
	if err != nil {
		return vk.NullPipeline, fmt.Errorf("createGraphicsPipeline: %w", err)
	}
	t.pipelinesMutex.Lock()
	t.pipelines[request.GraphicsPipelineKey] = pipeline
	t.pipelinesMutex.Unlock()

	return pipeline, nil
}

func (t *GpuTranslator) GetComputePipeline(request vulkan.ComputePipelineRequest) (vk.Pipeline, error) {
	// Get already created pipeline.
	t.computePipelinesMutex.Lock()
	pipeline, ok := t.computePipelines[request.ComputePipelineKey]
	t.computePipelinesMutex.Unlock()
	if ok {
		return pipeline, nil
	}

	// Create the pipeline.
	pipeline, err := vulkan.CreateComputePipeline(&t.handles, request, t.pipelineLayout, vk.NullPipelineCache)
	if err != nil {
		return vk.NullPipeline, fmt.Errorf("createComputePipeline: %w", err)
	}
	t.computePipelinesMutex.Lock()
	t.computePipelines[request.ComputePipelineKey] = pipeline
	t.computePipelinesMutex.Unlock()

	return pipeline, nil
}
