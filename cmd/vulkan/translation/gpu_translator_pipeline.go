package translation

import (
	"fmt"
	"math"

	gcn2 "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/reg"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gpu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	spirvCommon "github.com/LamkasDev/sharkie/cmd/spirv/common"

	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan"
	vkGcn "github.com/LamkasDev/sharkie/cmd/vulkan/gcn"
	vk "github.com/goki/vulkan"
	"github.com/gookit/color"
)

func (t *GpuTranslator) BindPipeline(frame uint64, bind *gpu.LiverpoolBindPipeline) {
	rtAddress := bind.RtBase.Address()
	dbAddress := bind.DbZWriteBase.Address()
	if rtAddress == 0 && dbAddress == 0 {
		return
	}

	// Handle depth surface.
	var dbWidth, dbHeight uint32
	var depthSurface *vulkan.VulkanSurface
	depthTestEnable := bind.DbDepthControl.ZEnable()
	depthWriteEnable := bind.DbDepthControl.ZWriteEnable()
	depthFormat, _ := vkGcn.TranslateGcnFormat(uint8(bind.DbZInfo.Format()), gcn2.GcnNumFormatConvertToDepthPls, 0)
	if dbAddress != 0 && depthFormat != vk.FormatUndefined && (depthTestEnable || depthWriteEnable) {
		dbWidth = uint32(bind.DbWidth)
		if dbWidth == 0 {
			dbWidth = (bind.DbDepthSize.PitchTileMax() + 1) * 8
		}
		dbHeight = uint32(bind.DbHeight)
		if dbHeight == 0 {
			dbHeight = (bind.DbDepthSize.HeightTileMax() + 1) * 8
		}

		var err error
		depthSurface, err = t.GetSurface(spirvStructs.ImageDescriptor{
			BaseAddress: dbAddress,
			Width:       uint16(dbWidth),
			Height:      uint16(dbHeight),
			DataFormat:  uint8(bind.DbZInfo.Format()), NumFormat: gcn2.GcnNumFormatConvertToDepthPls,
			DstSelX: 4, DstSelY: 5, DstSelZ: 6, DstSelW: 7,
			Depth: 1, Pitch: uint16(dbWidth),
		}, 0)
		if err != nil {
			return
		}
		dbWidth = uint32(depthSurface.ImageView.Image.FirstDescriptor.Width)
		dbHeight = uint32(depthSurface.ImageView.Image.FirstDescriptor.Height)
	} else {
		dbAddress = 0
		dbWidth = 0
		dbHeight = 0
	}
	t.activeDepthSurface = depthSurface

	// Handle color surface.
	var rtWidth, rtHeight uint32
	var colorSurface *vulkan.VulkanSurface
	rtFormat, _ := vkGcn.TranslateGcnFormat(uint8(bind.CbColorInfo0.Format()), uint8(bind.CbColorInfo0.NumberType()), bind.CbColorInfo0.CompSwap())
	if rtAddress != 0 && rtFormat != vk.FormatUndefined {
		rtWidth = uint32(bind.RtWidth)
		if rtWidth == 0 {
			rtWidth = vkGcn.ColorBufferPitch(reg.CbColorPitch(bind.RtPitch))
		}
		rtHeight = uint32(bind.RtHeight)
		if rtHeight == 0 {
			rtHeight = vkGcn.ColorBufferHeight(reg.CbColorPitch(bind.RtPitch), bind.RtSlice)
		}
		// TODO: can't figure this out.
		if rtWidth == 1920 && rtHeight == 1088 {
			rtHeight = 1080
		}
		if rtWidth == 640 && rtHeight == 512 {
			rtHeight = 480
		}

		var err error
		colorSurface, err = t.GetSurface(spirvStructs.ImageDescriptor{
			BaseAddress: rtAddress,
			Width:       uint16(rtWidth), Height: uint16(rtHeight),
			DataFormat: uint8(bind.CbColorInfo0.Format()), NumFormat: uint8(bind.CbColorInfo0.NumberType()),
			DstSelX: 4, DstSelY: 5, DstSelZ: 6, DstSelW: 7,
			Depth: 1, Pitch: uint16(rtWidth),
			TilingIndex: uint8(bind.RtAttrib.TileModeIndex()),
		}, bind.CbColorInfo0.CompSwap())
		if err != nil {
			return
		}
		rtWidth = uint32(colorSurface.ImageView.Image.FirstDescriptor.Width)
		rtHeight = uint32(colorSurface.ImageView.Image.FirstDescriptor.Height)
	} else {
		// Fallback for depth-only rendering.
		rtWidth = dbWidth
		rtHeight = dbHeight
	}
	t.activeSurface = colorSurface

	// The game keeps one depth buffer (DB_Z_WRITE_BASE) while ping-ponging CB_COLOR0.
	// Stale depth from a previous color target corrupts compositing on the next one.
	if t.lastColorRtAddress != 0 && t.lastColorRtAddress != rtAddress {
		t.surfacesMutex.Lock()
		clearDepthSurface := t.surfaces[bind.DbZWriteBase.Address()]
		t.surfacesMutex.Unlock()
		if clearDepthSurface != nil {
			clearDepthSurface.FrameUsed = 0
		}
	}
	t.lastColorRtAddress = rtAddress

	// Handle depth writes.
	if bind.SpiShaderZFormat.ZExportFormat() == 0 {
		depthWriteEnable = false // SPI_SHADER_ZERO (No depth export)
	}
	zfunc := bind.DbDepthControl.Zfunc()
	if zfunc == 7 { // ALWAYS
		depthWriteEnable = false
	}

	// Get or create framebuffer.
	format := vk.FormatUndefined
	if colorSurface != nil {
		format = colorSurface.ImageView.Image.ImageFormat
	}
	fbRequest := vulkan.FramebufferRequest{
		ImageView: nil,
		FramebufferKey: vulkan.FramebufferKey{
			GpuAddress:      rtAddress,
			DepthGpuAddress: dbAddress,
			Format:          format,
			Width:           rtWidth,
			Height:          rtHeight,
		},
	}
	if colorSurface != nil {
		fbRequest.ImageView = colorSurface.ImageView
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
	var vertexShaderKey, fragmentShaderKey spirvCommon.SpirvShaderKey
	if t.activeVertexShader != nil {
		vsModule, err = t.GetShaderModule(t.activeVertexShaderKey, t.activeVertexShader)
		if err != nil {
			return
		}
		vertexShaderKey = t.activeVertexShaderKey
	}
	if t.activeFragmentShader != nil {
		psModule, err = t.GetShaderModule(t.activeFragmentShaderKey, t.activeFragmentShader)
		if err != nil {
			return
		}
		fragmentShaderKey = t.activeFragmentShaderKey
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
	key := vulkan.GraphicsPipelineKey{
		VertexShaderKey:     vertexShaderKey,
		FragmentShaderKey:   fragmentShaderKey,
		RenderTargetAddress: rtAddress,
		DepthTargetAddress:  dbAddress,

		Width:    rtWidth,
		Height:   rtHeight,
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
		SpiVsOutConfig:            bind.SpiVsOutConfig,
		PaClVsOutCntl:             bind.PaClVsOutCntl,
		CbColorInfo0:              bind.CbColorInfo0,
		CbTargetMask:              bind.RtTargetMask,
		CbShaderMask:              bind.CbShaderMask,
		CbColorControl:            bind.RtColorControl,
		CbColorAttrib:             bind.RtAttrib,
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
	}
	pipeline, err := t.GetPipeline(vulkan.GraphicsPipelineRequest{
		VertexModule:        vsModule,
		TessControlModule:   tcsModule,
		TessEvalModule:      tesModule,
		GeometryModule:      gsModule,
		FragmentModule:      psModule,
		ColorFormat:         fb.ColorFormat,
		DepthFormat:         fb.DepthFormat,
		GraphicsPipelineKey: key,
	})
	if err != nil {
		return
	}

	// Prepare rendering info.
	renderingInfo := &vk.RenderingInfo{
		SType:      vk.StructureTypeRenderingInfo,
		RenderArea: vk.Rect2D{Extent: vk.Extent2D{Width: rtWidth, Height: rtHeight}},
		LayerCount: 1,
	}

	colorAttachments := make([]vk.RenderingAttachmentInfo, 8)
	for i := 0; i < 8; i++ {
		colorAttachments[i] = vk.RenderingAttachmentInfo{
			SType: vk.StructureTypeRenderingAttachmentInfo,
		}
	}
	if colorSurface != nil {
		clearColorNeeded := colorSurface.FrameUsed < frame
		loadOp := vk.AttachmentLoadOpLoad
		if clearColorNeeded {
			loadOp = vk.AttachmentLoadOpClear
			colorSurface.FrameUsed = frame
		}

		clearColor := vk.ClearValue{}
		clearColorFloat := vkGcn.TranslateClearColor(bind.RtClearWord0, bind.RtClearWord1, bind.CbColorInfo0.Format(), bind.CbColorInfo0.NumberType(), bind.CbColorInfo0.CompSwap())
		clearColor.SetColor(clearColorFloat)

		colorAttachments[0] = vk.RenderingAttachmentInfo{
			SType:       vk.StructureTypeRenderingAttachmentInfo,
			ImageView:   fb.ColorView,
			ImageLayout: vk.ImageLayoutGeneral,
			LoadOp:      loadOp,
			StoreOp:     vk.AttachmentStoreOpStore,
			ClearValue:  clearColor,
		}
		colorSurface.ImageView.Image.BarrierColorAttachment(t.commandBuffer)
	}

	renderingInfo.ColorAttachmentCount = 8
	renderingInfo.PColorAttachments = colorAttachments

	var depthAttachments []vk.RenderingAttachmentInfo
	if depthSurface != nil {
		clearDepthNeeded := depthSurface.FrameUsed < frame
		loadOp := vk.AttachmentLoadOpLoad
		if clearDepthNeeded {
			loadOp = vk.AttachmentLoadOpClear
			depthSurface.FrameUsed = frame
		}

		clearDepth := vk.ClearValue{}
		clearDepth.SetDepthStencil(math.Float32frombits(bind.DbDepthClearValue), bind.DbStencilClearValue)

		depthAttachments = append(depthAttachments, vk.RenderingAttachmentInfo{
			SType:       vk.StructureTypeRenderingAttachmentInfo,
			ImageView:   fb.DepthView,
			ImageLayout: vk.ImageLayoutDepthStencilAttachmentOptimal,
			LoadOp:      loadOp,
			StoreOp:     vk.AttachmentStoreOpStore,
			ClearValue:  clearDepth,
		})

		renderingInfo.PDepthAttachment = depthAttachments
		renderingInfo.PStencilAttachment = depthAttachments
	}

	// Start rendering and bind pipeline.
	vulkan.CmdBeginRendering(t.handles.Instance, t.commandBuffer.CommandBuffer, renderingInfo)
	vk.CmdBindPipeline(t.commandBuffer.CommandBuffer, vk.PipelineBindPointGraphics, pipeline)

	t.insideRenderPass = true
	t.activePipeline = pipeline

	if logger.LogRenderer {
		logger.Printf("[%s] Bound pipeline (vertex=%s, fragment=%s, rtAddress=0x%X, rtPitch=%d, rtSize=%dx%d).\n",
			color.Blue.Sprintf("Frame %d", frame),
			color.Yellow.Sprintf("0x%X", vertexShaderKey.Address),
			color.Yellow.Sprintf("0x%X", fragmentShaderKey.Address),
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
	pipeline, err := vulkan.CreateGraphicsPipeline(t.handles, request, t.pipelineLayout, vk.NullPipelineCache)
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
