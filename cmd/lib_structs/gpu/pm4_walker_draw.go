package gpu

import (
	"math"

	. "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/reg"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/spirv/common"
	"github.com/gookit/color"
)

func (l *Liverpool) handleDrawIndexAuto(stream *LiverpoolCommandStream, payload []uint32) {
	if len(payload) < 2 {
		logger.Printf("[%s] failed draw index auto payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", stream.Name, len(payload)),
		)
		return
	}

	// Record draw.
	l.DrawState.IndexCount = payload[0]
	l.DrawState.IndexOffset = 0
	l.recordDraw(stream, false)
}

func (l *Liverpool) handleDrawIndex2(stream *LiverpoolCommandStream, payload []uint32) {
	if len(payload) < 5 {
		logger.Printf("[%s] failed draw index 2 payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", stream.Name, len(payload)),
		)
		return
	}

	// Record draw.
	l.DrawState.IndexBase = uintptr(uint64(payload[1]) | uint64(payload[2])<<32)
	l.DrawState.IndexCount = payload[3]
	l.DrawState.IndexOffset = 0
	l.recordDraw(stream, true)
}

func (l *Liverpool) handleDrawIndexOffset2(stream *LiverpoolCommandStream, payload []uint32) {
	if len(payload) < 4 {
		logger.Printf("[%s] failed draw index offset 2 payload too short.\n",
			color.Green.Sprintf("PM4-%s/%d", stream.Name, len(payload)),
		)
		return
	}

	// Record draw.
	l.DrawState.IndexCount = payload[2]
	l.DrawState.IndexOffset = payload[1]
	l.recordDraw(stream, true)
}

func (l *Liverpool) recordDraw(stream *LiverpoolCommandStream, isIndexed bool) {
	// Lock registers.
	l.StateMutex.Lock()
	defer l.StateMutex.Unlock()

	// Construct resource binds.
	bindResources := LiverpoolBindResources{
		LiverpoolBindResourcesInternal: LiverpoolBindResourcesInternal{
			UserDataHash: l.SnapshotUserData(),
		},
	}
	if address := l.VsGpuAddress(); address != 0 {
		bindResources.VertexShader = l.GetShader(GcnShaderStageVertex, address)
		bindResources.VertexShaderAddress = address
		paClVsOutCntl := reg.PaClVsOutCntl(l.Registers.Context[GREG_MM_PA_CL_VS_OUT_CNTL])
		bindResources.VertexContext = common.SpirvVertexShaderContext{
			ClipDistEnable: paClVsOutCntl.ClipDistEna(),
			CullDistEnable: paClVsOutCntl.CullDistEna(),
		}
	} else {
		panic("no vertex shader")
	}
	if address := l.PsGpuAddress(); address != 0 {
		bindResources.FragmentShader = l.GetShader(GcnShaderStageFragment, address)
		bindResources.FragmentShaderAddress = address
		psInputAddress := reg.SpiPsInputAddr(l.Registers.Context[GREG_MM_SPI_PS_INPUT_ADDR])
		dbShaderControl := reg.DbShaderControl(l.Registers.Context[GREG_MM_DB_SHADER_CONTROL])
		bindResources.FragmentContext = common.SpirvFragmentShaderContext{
			PsInControl:       l.Registers.Context[GREG_MM_SPI_PS_IN_CONTROL],
			PsInputAddress:    l.Registers.Context[GREG_MM_SPI_PS_INPUT_ADDR],
			DepthBeforeShader: dbShaderControl.DepthBeforeShader(),
			ZOrder:            dbShaderControl.ZOrder(),
			FrontFaceEnable:   psInputAddress.FrontFaceEna(),
		}
		copy(bindResources.FragmentContext.PsInputControls[:], l.Registers.Context[GREG_MM_SPI_PS_INPUT_CNTL_0:GREG_MM_SPI_PS_INPUT_CNTL_31+1])
	}
	if address := l.HsGpuAddress(); address != 0 {
		bindResources.HullShader = l.GetShader(GcnShaderStageHull, address)
		bindResources.HullShaderAddress = address
	}
	if address := l.EsGpuAddress(); address != 0 {
		bindResources.EvalShader = l.GetShader(GcnShaderStageEvaluation, address)
		bindResources.EvalShaderAddress = address
	}
	if address := l.GsGpuAddress(); address != 0 {
		bindResources.GeometryShader = l.GetShader(GcnShaderStageGeometry, address)
		bindResources.GeometryShaderAddress = address
	}

	// Add to command stream.
	resHash := bindResources.Hash()
	resIndex, ok := stream.BindResourcesMap[resHash]
	if !ok {
		resIndex = uint32(len(stream.BindResources))
		stream.BindResources = append(stream.BindResources, bindResources)
		stream.BindResourcesMap[resHash] = resIndex
	}
	stream.Commands = append(stream.Commands, LiverpoolCommand{Type: LiverpoolCommandTypeBindResources, Index: resIndex})

	// Construct pipeline state.
	bindPipeline := LiverpoolBindPipeline{
		LiverpoolBindPipelineInternal: LiverpoolBindPipelineInternal{
			PrimType: l.Registers.UserConfig[GREG_MM_VGT_PRIMITIVE_TYPE__CI__VI],

			RtBase:             reg.GpuMemoryBase(l.Registers.Context[GREG_MM_CB_COLOR0_BASE]),
			RtPitch:            reg.CbColorPitch(l.Registers.Context[GREG_MM_CB_COLOR0_PITCH]),
			RtSlice:            l.Registers.Context[GREG_MM_CB_COLOR0_SLICE],
			RtView:             reg.CbColorView(l.Registers.Context[GREG_MM_CB_COLOR0_VIEW]),
			RtAttrib:           reg.CbColorAttrib(l.Registers.Context[GREG_MM_CB_COLOR0_ATTRIB]),
			RtTargetMask:       reg.CbTargetMask(l.Registers.Context[GREG_MM_CB_TARGET_MASK]),
			RtColorControl:     reg.CbColorControl(l.Registers.Context[GREG_MM_CB_COLOR_CONTROL]),
			RtBlendControl:     reg.CbBlendControl(l.Registers.Context[GREG_MM_CB_BLEND0_CONTROL]),
			SpiShaderColFormat: reg.SpiShaderColFormat(l.Registers.Context[GREG_MM_SPI_SHADER_COL_FORMAT]),
			SpiShaderZFormat:   reg.SpiShaderZFormat(l.Registers.Context[GREG_MM_SPI_SHADER_Z_FORMAT]),
			PaClVsOutCntl:      reg.PaClVsOutCntl(l.Registers.Context[GREG_MM_PA_CL_VS_OUT_CNTL]),
			RtClearWord0:       l.Registers.Context[GREG_MM_CB_COLOR0_CLEAR_WORD0],
			RtClearWord1:       l.Registers.Context[GREG_MM_CB_COLOR0_CLEAR_WORD1],

			PaSuScModeCntl: reg.PaSuScModeCntl(l.Registers.Context[GREG_MM_PA_SU_SC_MODE_CNTL]),

			CbColorInfo0: reg.CbColorInfo(l.Registers.Context[GREG_MM_CB_COLOR0_INFO]),
			CbShaderMask: reg.CbShaderMask(l.Registers.Context[GREG_MM_CB_SHADER_MASK]),

			PaSuPolyOffsetClamp:       reg.PaSuPolyOffsetClamp(l.Registers.Context[GREG_MM_PA_SU_POLY_OFFSET_CLAMP]),
			PaSuPolyOffsetFrontScale:  reg.PaSuPolyOffsetFrontScale(l.Registers.Context[GREG_MM_PA_SU_POLY_OFFSET_FRONT_SCALE]),
			PaSuPolyOffsetFrontOffset: reg.PaSuPolyOffsetFrontOffset(l.Registers.Context[GREG_MM_PA_SU_POLY_OFFSET_FRONT_OFFSET]),
			PaSuPolyOffsetBackScale:   reg.PaSuPolyOffsetBackScale(l.Registers.Context[GREG_MM_PA_SU_POLY_OFFSET_BACK_SCALE]),
			PaSuPolyOffsetBackOffset:  reg.PaSuPolyOffsetBackOffset(l.Registers.Context[GREG_MM_PA_SU_POLY_OFFSET_BACK_OFFSET]),
			DbShaderControl:           reg.DbShaderControl(l.Registers.Context[GREG_MM_DB_SHADER_CONTROL]),

			DbDepthControl:    reg.DbDepthControl(l.Registers.Context[GREG_MM_DB_DEPTH_CONTROL]),
			DbDepthClearValue: l.Registers.Context[GREG_MM_DB_DEPTH_CLEAR],
			DbDepthSize:       reg.DbDepthSize(l.Registers.Context[GREG_MM_DB_DEPTH_SIZE]),
			DbZWriteBase: func() reg.GpuMemoryBase {
				if l.Registers.Context[GREG_MM_DB_Z_WRITE_BASE] != 0 {
					return reg.GpuMemoryBase(l.Registers.Context[GREG_MM_DB_Z_WRITE_BASE])
				}
				return reg.GpuMemoryBase(l.Registers.Context[GREG_MM_DB_Z_READ_BASE])
			}(),
			DbZInfo: reg.DbZInfo(l.Registers.Context[GREG_MM_DB_Z_INFO]),

			DbStencilControl:    reg.DbStencilControl(l.Registers.Context[GREG_MM_DB_STENCIL_CONTROL]),
			DbStencilRefMask:    reg.DbStencilrefmask(l.Registers.Context[GREG_MM_DB_STENCILREFMASK]),
			DbStencilRefMaskBf:  reg.DbStencilrefmaskBf(l.Registers.Context[GREG_MM_DB_STENCILREFMASK_BF]),
			DbStencilClearValue: l.Registers.Context[GREG_MM_DB_STENCIL_CLEAR],

			DbRenderControl: reg.DbRenderControl(l.Registers.Context[GREG_MM_DB_RENDER_CONTROL]),

			PaScModeCntl0:         reg.PaScModeCntl0(l.Registers.Context[GREG_MM_PA_SC_MODE_CNTL_0]),
			PaScAaConfig:          reg.PaScAaConfig(l.Registers.Context[GREG_MM_PA_SC_AA_CONFIG]),
			VgtMultiPrimIbResetEn: reg.VgtMultiPrimIbResetEn(l.Registers.Context[GREG_MM_VGT_MULTI_PRIM_IB_RESET_EN]),
			PaSuLineCntl:          reg.PaSuLineCntl(l.Registers.Context[GREG_MM_PA_SU_LINE_CNTL]),
			PaScAaMaskX0y0X1y0:    reg.PaScAaMaskX0y0X1y0(l.Registers.Context[GREG_MM_PA_SC_AA_MASK_X0Y0_X1Y0]),
			PaScAaMaskX0y1X1y1:    reg.PaScAaMaskX0y1X1y1(l.Registers.Context[GREG_MM_PA_SC_AA_MASK_X0Y1_X1Y1]),

			PsInControl:    reg.SpiPsInControl(l.Registers.Context[GREG_MM_SPI_PS_IN_CONTROL]),
			PsInputAddress: reg.SpiPsInputAddr(l.Registers.Context[GREG_MM_SPI_PS_INPUT_ADDR]),

			MultiPrimIbResetIndex: l.Registers.Context[GREG_MM_VGT_MULTI_PRIM_IB_RESET_INDX],

			UserDataHash: bindResources.UserDataHash,
		},
	}

	// Add to command stream.
	bindHash := bindPipeline.Hash()
	bindIndex, ok := stream.PipelinesMap[bindHash]
	if !ok {
		bindIndex = uint32(len(stream.Pipelines))
		stream.Pipelines = append(stream.Pipelines, bindPipeline)
		stream.PipelinesMap[bindHash] = bindIndex
	}
	stream.Commands = append(stream.Commands, LiverpoolCommand{Type: LiverpoolCommandTypeBindPipeline, Index: bindIndex})

	// Construct dynamic state.
	setDynamicState := LiverpoolSetDynamicState{
		LiverpoolSetDynamicStateInternal: LiverpoolSetDynamicStateInternal{
			VpXScale:  math.Float32frombits(l.Registers.Context[GREG_MM_PA_CL_VPORT_XSCALE]),
			VpXOffset: math.Float32frombits(l.Registers.Context[GREG_MM_PA_CL_VPORT_XOFFSET]),
			VpYScale:  math.Float32frombits(l.Registers.Context[GREG_MM_PA_CL_VPORT_YSCALE]),
			VpYOffset: math.Float32frombits(l.Registers.Context[GREG_MM_PA_CL_VPORT_YOFFSET]),
			VpZScale:  math.Float32frombits(l.Registers.Context[GREG_MM_PA_CL_VPORT_ZSCALE]),
			VpZOffset: math.Float32frombits(l.Registers.Context[GREG_MM_PA_CL_VPORT_ZOFFSET]),
			VpZMin:    math.Float32frombits(l.Registers.Context[GREG_MM_PA_SC_VPORT_ZMIN_0]),
			VpZMax:    math.Float32frombits(l.Registers.Context[GREG_MM_PA_SC_VPORT_ZMAX_0]),

			PaClVteCntl: reg.PaClVteCntl(l.Registers.Context[GREG_MM_PA_CL_VTE_CNTL]),

			ClipControl:    reg.PaClClipCntl(l.Registers.Context[GREG_MM_PA_CL_CLIP_CNTL]),
			PaScModeCntl0:  reg.PaScModeCntl0(l.Registers.Context[GREG_MM_PA_SC_MODE_CNTL_0]),
			PaSuScModeCntl: reg.PaSuScModeCntl(l.Registers.Context[GREG_MM_PA_SU_SC_MODE_CNTL]),
			GbVertClipAdj:  math.Float32frombits(l.Registers.Context[GREG_MM_PA_CL_GB_VERT_CLIP_ADJ]),
			GbVertDiscAdj:  math.Float32frombits(l.Registers.Context[GREG_MM_PA_CL_GB_VERT_DISC_ADJ]),
			GbHorzClipAdj:  math.Float32frombits(l.Registers.Context[GREG_MM_PA_CL_GB_HORZ_CLIP_ADJ]),
			GbHorzDiscAdj:  math.Float32frombits(l.Registers.Context[GREG_MM_PA_CL_GB_HORZ_DISC_ADJ]),

			BlendRed:   l.Registers.Context[GREG_MM_CB_BLEND_RED],
			BlendGreen: l.Registers.Context[GREG_MM_CB_BLEND_GREEN],
			BlendBlue:  l.Registers.Context[GREG_MM_CB_BLEND_BLUE],
			BlendAlpha: l.Registers.Context[GREG_MM_CB_BLEND_ALPHA],

			ScissorTl: reg.PaScScreenScissorTl(l.Registers.Context[GREG_MM_PA_SC_SCREEN_SCISSOR_TL]),
			ScissorBr: l.Registers.Context[GREG_MM_PA_SC_SCREEN_SCISSOR_BR],

			VpScissorTl: reg.PaScVportScissorTl(l.Registers.Context[GREG_MM_PA_SC_VPORT_SCISSOR_0_TL]),
			VpScissorBr: l.Registers.Context[GREG_MM_PA_SC_VPORT_SCISSOR_0_BR],

			GenericScissorTl: reg.PaScGenericScissorTl(l.Registers.Context[GREG_MM_PA_SC_GENERIC_SCISSOR_TL]),
			GenericScissorBr: l.Registers.Context[GREG_MM_PA_SC_GENERIC_SCISSOR_BR],

			WindowScissorTl: reg.PaScWindowScissorTl(l.Registers.Context[GREG_MM_PA_SC_WINDOW_SCISSOR_TL]),
			WindowScissorBr: l.Registers.Context[GREG_MM_PA_SC_WINDOW_SCISSOR_BR],
			WindowOffset:    reg.PaScWindowOffset(l.Registers.Context[GREG_MM_PA_SC_WINDOW_OFFSET]),

			LineStippleRepeatCount: (l.Registers.Context[GREG_MM_PA_SU_LINE_STIPPLE_CNTL] >> 16) & 0xFF,
			LineStipplePattern:     l.Registers.Context[GREG_MM_PA_SU_LINE_STIPPLE_CNTL] & 0xFFFF,

			HardwareScreenOffset: reg.PaSuHardwareScreenOffset(l.Registers.Context[GREG_MM_PA_SU_HARDWARE_SCREEN_OFFSET]),
		},
	}

	// Add to command stream.
	dynHash := setDynamicState.Hash()
	dynIndex, ok := stream.DynamicStatesMap[dynHash]
	if !ok {
		dynIndex = uint32(len(stream.DynamicStates))
		stream.DynamicStates = append(stream.DynamicStates, setDynamicState)
		stream.DynamicStatesMap[dynHash] = dynIndex
	}
	stream.Commands = append(stream.Commands, LiverpoolCommand{Type: LiverpoolCommandTypeSetDynamicState, Index: dynIndex})

	// Construct draw.
	draw := LiverpoolDraw{
		LiverpoolDrawInternal: LiverpoolDrawInternal{
			InstanceCount: max(l.DrawState.InstanceCount, 1),
			PrimType:      bindPipeline.PrimType,
			IsIndexed:     isIndexed,

			IndexType:   l.DrawState.IndexType,
			IndexBase:   l.DrawState.IndexBase,
			IndexCount:  l.DrawState.IndexCount,
			IndexOffset: l.DrawState.IndexOffset,

			VertexShRsrc1:   l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_RSRC1_VS],
			VertexShRsrc2:   l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_RSRC2_VS],
			PixelShRsrc1:    l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_RSRC1_PS],
			PixelShRsrc2:    l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_RSRC2_PS],
			HullShRsrc1:     l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_RSRC1_HS],
			HullShRsrc2:     l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_RSRC2_HS],
			EvalShRsrc1:     l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_RSRC1_ES],
			EvalShRsrc2:     l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_RSRC2_ES],
			GeometryShRsrc1: l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_RSRC1_GS],
			GeometryShRsrc2: l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_RSRC2_GS],

			DbRenderControl:     reg.DbRenderControl(l.Registers.Context[GREG_MM_DB_RENDER_CONTROL]),
			DbDepthClearValue:   l.Registers.Context[GREG_MM_DB_DEPTH_CLEAR],
			DbStencilClearValue: l.Registers.Context[GREG_MM_DB_STENCIL_CLEAR],

			UserDataHash: bindPipeline.UserDataHash,
		},
	}

	// Add to command stream.
	drawHash := draw.Hash()
	drawIndex, ok := stream.DrawsMap[bindHash]
	if !ok || true {
		drawIndex = uint32(len(stream.Draws))
		stream.Draws = append(stream.Draws, draw)
		stream.DrawsMap[drawHash] = drawIndex
	}
	stream.Commands = append(stream.Commands, LiverpoolCommand{Type: LiverpoolCommandTypeDraw, Index: drawIndex})

	if LogPM4Packets {
		if isIndexed {
			logger.Printf("[%s] draw index 2 (index_base=%s, index_count=%s, prim=%s, rt=%s, vs=%s, ps=%s).\n",
				color.Green.Sprintf("PM4-%s", stream.Name),
				color.Yellow.Sprintf("0x%X", draw.IndexBase),
				color.Green.Sprintf("%d", draw.IndexCount),
				color.Green.Sprintf("%d", draw.PrimType),
				color.Yellow.Sprintf("0x%X", bindPipeline.RtBase),
				color.Yellow.Sprintf("0x%X", bindResources.VertexShader.Address),
				color.Yellow.Sprintf("0x%X", bindResources.FragmentShader.Address),
			)
		} else {
			logger.Printf("[%s] draw index auto (index_count=%s, prim=%s, rt=%s, vs=%s, ps=%s).\n",
				color.Green.Sprintf("PM4-%s", stream.Name),
				color.Green.Sprintf("%d", draw.IndexCount),
				color.Green.Sprintf("%d", draw.PrimType),
				color.Yellow.Sprintf("0x%X", bindPipeline.RtBase),
				color.Yellow.Sprintf("0x%X", bindResources.VertexShader.Address),
				color.Yellow.Sprintf("0x%X", bindResources.FragmentShader.Address),
			)
		}
	}
}
