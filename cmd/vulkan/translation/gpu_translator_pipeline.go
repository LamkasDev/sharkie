package translation

import (
	"fmt"
	"math"

	"github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/reg"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gpu"
	"github.com/LamkasDev/sharkie/cmd/logger"

	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan"
	vkGcn "github.com/LamkasDev/sharkie/cmd/vulkan/gcn"
	vk "github.com/goki/vulkan"
	"github.com/gookit/color"
)

func (t *GpuTranslator) BindPipeline(frame uint64, bind *gpu.LiverpoolBindPipeline) {
	rtAddress := bind.RtBase.Address()
	if rtAddress == 0 {
		return
	}

	rtWidth := vkGcn.ColorBufferPitch(reg.CbColorPitch(bind.RtPitch))
	rtHeight := vkGcn.ColorBufferHeight(reg.CbColorPitch(bind.RtPitch), bind.RtSlice)

	// Get or create surface.
	surface, err := t.GetSurface(spirvStructs.ImageDescriptor{
		BaseAddress: rtAddress,
		Width:       uint16(rtWidth), Height: uint16(rtHeight),
		DataFormat: 10, NumFormat: 0,
		DstSelX: 4, DstSelY: 5, DstSelZ: 6, DstSelW: 7,
		Depth: 1, Pitch: uint16(rtWidth),
		TilingIndex: uint8(bind.RtAttrib.TileModeIndex()),
	}, vkGcn.TranslateColorFormat(bind.CbColorInfo0.Format(), bind.CbColorInfo0.NumberType(), bind.CbColorInfo0.CompSwap()))
	if err != nil {
		return
	}
	t.activeSurface = surface

	// The game keeps one depth buffer (DB_Z_WRITE_BASE) while ping-ponging CB_COLOR0.
	// Stale depth from a previous color target corrupts compositing on the next one.
	if t.lastColorRtAddress != 0 && t.lastColorRtAddress != rtAddress {
		t.surfacesMutex.Lock()
		depthSurface := t.surfaces[bind.DbZWriteBase.Address()]
		t.surfacesMutex.Unlock()
		if depthSurface != nil {
			depthSurface.FrameUsed = 0
		}
	}
	t.lastColorRtAddress = rtAddress

	// Handle depth surface.
	var depthSurface *vulkan.VulkanSurface
	dbAddress := bind.DbZWriteBase.Address()
	depthFormat := vkGcn.TranslateGcnDepthFormat(bind.DbZInfo.Format(), t.handles.FormatProperties)
	depthTestEnable := bind.DbDepthControl.ZEnable()
	depthWriteEnable := bind.DbDepthControl.ZWriteEnable()
	if bind.SpiShaderZFormat.ZExportFormat() == 0 {
		depthWriteEnable = false // SPI_SHADER_ZERO (No depth export)
	}
	zfunc := bind.DbDepthControl.Zfunc()
	if zfunc == 7 { // ALWAYS
		depthWriteEnable = false
	}
	if dbAddress != 0 && depthFormat != vk.FormatUndefined && (depthTestEnable || depthWriteEnable) {
		depthSurface, err = t.GetSurface(spirvStructs.ImageDescriptor{
			BaseAddress: dbAddress,
			Width:       uint16(surface.ImageView.Image.FirstDescriptor.Width),
			Height:      uint16(surface.ImageView.Image.FirstDescriptor.Height),
			DataFormat:  10, NumFormat: 0,
			DstSelX: 4, DstSelY: 5, DstSelZ: 6, DstSelW: 7,
			Depth: 1, Pitch: surface.ImageView.Image.FirstDescriptor.Width,
		}, depthFormat)
		if err != nil {
			return
		}
	} else {
		dbAddress = 0
	}

	// Get or create framebuffer.
	fbRequest := vulkan.FramebufferRequest{
		ImageView: surface.ImageView,
		FramebufferKey: vulkan.FramebufferKey{
			GpuAddress:      rtAddress,
			DepthGpuAddress: dbAddress,
			Format:          surface.ImageView.Image.ImageFormat,
			Width:           rtWidth,
			Height:          rtHeight,
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

	// Gather shaders.
	var vsModule, psModule vk.ShaderModule
	var fetchShaderAddress uintptr
	var vertexShaderAddress, pixelShaderAddress uintptr
	if t.activeVertexShader != nil {
		vsModule, err = t.GetShaderModule(t.activeVertexShaderKey, t.activeVertexShader)
		if err != nil {
			return
		}
		fetchShaderAddress = t.activeVertexShaderKey.FetchShaderAddress
		vertexShaderAddress = t.activeVertexShader.GcnShader.Address
	}
	if t.activeFragmentShader != nil {
		psModule, err = t.GetShaderModule(t.activeFragmentShaderKey, t.activeFragmentShader)
		if err != nil {
			return
		}
		pixelShaderAddress = t.activeFragmentShader.GcnShader.Address
	}
	var tcsModule, tesModule, gsModule vk.ShaderModule
	if bind.PrimType == 17 { // RECTLIST
		tcsModule, err = t.GetRectlistTescShader()
		if err != nil {
			return
		}
		tesModule, err = t.GetRectlistTeseShader()
		if err != nil {
			return
		}
	} else if t.activeGeometryShader != nil {
		gsModule, err = t.GetShaderModule(t.activeGeometryShaderKey, t.activeGeometryShader)
		if err != nil {
			return
		}
	}

	// Translate pipeline parameters.
	colorWriteMask := bind.RtTargetMask
	if bind.RtColorControl.Mode() == 0 { // CB_DISABLE
		colorWriteMask = 0
	}

	logicOpEnable := vk.Bool32(vk.False)
	logicOp := vk.LogicOpCopy
	if !bind.RtBlendControl.Enable() { // 0 means blending disabled.
		if !bind.RtBlendControl.DisableRop3() { // 1 means ROP3 disabled.
			rop3 := bind.RtColorControl.Rop3()
			logicOpEnable = vk.Bool32(vk.True)
			logicOp = vkGcn.TranslateLogicOp(rop3)
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
			VertexModuleAddress:   vertexShaderAddress,
			FetchShaderAddress:    fetchShaderAddress,
			FragmentModuleAddress: pixelShaderAddress,
			RenderTargetAddress:   rtAddress,
			DepthTargetAddress:    dbAddress,

			Width:    uint32(surface.ImageView.Image.FirstDescriptor.Width),
			Height:   uint32(surface.ImageView.Image.FirstDescriptor.Height),
			PrimType: bind.PrimType,

			PaSuScModeCntl:            bind.PaSuScModeCntl,
			PaScModeCntl0:             bind.PaScModeCntl0,
			DbShaderControl:           bind.DbShaderControl,
			PaScAaConfig:              bind.PaScAaConfig,
			VgtMultiPrimIbResetEn:     bind.VgtMultiPrimIbResetEn,
			PaSuLineCntl:              bind.PaSuLineCntl,
			PaScAaMaskX0y0X1y0:        bind.PaScAaMaskX0y0X1y0,
			PaScAaMaskX0y1X1y1:        bind.PaScAaMaskX0y1X1y1,
			SpiShaderColFormat:        bind.SpiShaderColFormat,
			SpiShaderZFormat:          bind.SpiShaderZFormat,
			PaClVsOutCntl:             bind.PaClVsOutCntl,
			CbColorInfo0:              bind.CbColorInfo0,
			CbTargetMask:              bind.RtTargetMask,
			CbShaderMask:              bind.CbShaderMask,
			CbColorControl:            bind.RtColorControl,
			PaSuPolyOffsetClamp:       bind.PaSuPolyOffsetClamp,
			PaSuPolyOffsetFrontScale:  bind.PaSuPolyOffsetFrontScale,
			PaSuPolyOffsetFrontOffset: bind.PaSuPolyOffsetFrontOffset,
			PaSuPolyOffsetBackScale:   bind.PaSuPolyOffsetBackScale,
			PaSuPolyOffsetBackOffset:  bind.PaSuPolyOffsetBackOffset,
			DbRenderControl:           bind.DbRenderControl,

			BlendAttachment:   vkGcn.TranslateBlendControl(bind.RtBlendControl, reg.CbTargetMask(colorWriteMask), bind.CbColorInfo0.BlendBypass()),
			DepthStencilState: vkGcn.TranslateDepthControl(bind.DbDepthControl, bind.DbStencilControl, bind.DbStencilRefMask, bind.DbStencilRefMaskBf),
			LogicOpEnable:     logicOpEnable,
			LogicOp:           logicOp,
		},
	})
	if err != nil {
		return
	}

	// Select render pass and clear on first use in frame or if explicitly requested.
	var clearValues []vk.ClearValue
	clearColor := vk.ClearValue{}
	clearColorFloat := vkGcn.TranslateClearColor(bind.RtClearWord0, bind.RtClearWord1, bind.CbColorInfo0.Format(), bind.CbColorInfo0.NumberType(), bind.CbColorInfo0.CompSwap())
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
	surface.ImageView.Image.BarrierColorAttachment(t.commandBuffer)

	if renderPass == fb.RenderPassNoClear {
		clearValues = nil
	} else if renderPass == fb.RenderPassClearColorLoadDepth {
		clearValues = clearValues[:1]
	} else if renderPass == fb.RenderPassLoadColorClearDepth {
		// Needs both elements because depth is index 1.
		// Color clear value is ignored since loadOp is LOAD, but must be present.
	}

	// Begin render pass.
	vk.CmdBeginRenderPass(t.commandBuffer.CommandBuffer, &vk.RenderPassBeginInfo{
		SType:       vk.StructureTypeRenderPassBeginInfo,
		RenderPass:  renderPass,
		Framebuffer: fb.Framebuffer,
		RenderArea: vk.Rect2D{Extent: vk.Extent2D{
			Width:  uint32(surface.ImageView.Image.FirstDescriptor.Width),
			Height: uint32(surface.ImageView.Image.FirstDescriptor.Height),
		}},
		ClearValueCount: uint32(len(clearValues)),
		PClearValues:    clearValues,
	}, vk.SubpassContentsInline)
	vk.CmdBindPipeline(t.commandBuffer.CommandBuffer, vk.PipelineBindPointGraphics, pipeline)

	t.activeFramebuffer = fb.Framebuffer
	t.activePipeline = pipeline

	if logger.LogRenderer {
		logger.Printf("[%s] Bound pipeline (vertex=%s, fragment=%s, rtAddress=0x%X, rtPitch=%d, rtSize=%dx%d).\n",
			color.Blue.Sprintf("Frame %d", frame),
			color.Yellow.Sprintf("0x%X", vertexShaderAddress),
			color.Yellow.Sprintf("0x%X", pixelShaderAddress),
			rtAddress, bind.RtPitch, rtWidth, rtHeight,
		)
	}
}

func (t *GpuTranslator) BindComputePipeline(frame uint64, bind *gpu.LiverpoolBindComputePipeline) {
	// Gather shaders.
	var csModule vk.ShaderModule
	var err error
	if t.activeComputeShader != nil {
		csModule, err = t.GetShaderModule(t.activeComputeShaderKey, t.activeComputeShader)
		if err != nil {
			panic(err)
		}
	}

	// Get pipeline.
	pipeline, err := t.GetComputePipeline(vulkan.ComputePipelineRequest{
		ComputeModule: csModule,
		ComputePipelineKey: vulkan.ComputePipelineKey{
			ComputeModuleAddress: t.activeComputeShader.GcnShader.Address,
		},
	})
	if err != nil {
		panic(err)
	}

	// Bind pipeline.
	vk.CmdBindPipeline(t.commandBuffer.CommandBuffer, vk.PipelineBindPointCompute, pipeline)
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
	pipeline, err := vulkan.CreateGraphicsPipeline(t.handles, request, request.RenderPass, t.pipelineLayout, vk.NullPipelineCache)
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
	pipeline, err := vulkan.CreateComputePipeline(t.handles, request, t.pipelineLayout, vk.NullPipelineCache)
	if err != nil {
		return vk.NullPipeline, fmt.Errorf("createComputePipeline: %w", err)
	}
	t.computePipelinesMutex.Lock()
	t.computePipelines[request.ComputePipelineKey] = pipeline
	t.computePipelinesMutex.Unlock()

	return pipeline, nil
}
