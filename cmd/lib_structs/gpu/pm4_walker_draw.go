package gpu

import (
	"math"

	. "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
	"github.com/LamkasDev/sharkie/cmd/logger"
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

	// Construct pipeline state.
	bindPipeline := LiverpoolBindPipeline{
		VertexShader: l.GetShader(GcnShaderStageVertex, l.VsGpuAddress()),
		LiverpoolBindPipelineInternal: LiverpoolBindPipelineInternal{
			PrimType: l.Registers.UserConfig[GREG_MM_VGT_PRIMITIVE_TYPE__CI__VI],

			RtBase:         l.Registers.Context[GREG_MM_CB_COLOR0_BASE],
			RtPitch:        l.Registers.Context[GREG_MM_CB_COLOR0_PITCH],
			RtSlice:        l.Registers.Context[GREG_MM_CB_COLOR0_SLICE],
			RtView:         l.Registers.Context[GREG_MM_CB_COLOR0_VIEW],
			RtAttrib:       l.Registers.Context[GREG_MM_CB_COLOR0_ATTRIB],
			RtTargetMask:   l.Registers.Context[GREG_MM_CB_TARGET_MASK],
			RtColorControl: l.Registers.Context[GREG_MM_CB_COLOR_CONTROL],
			RtBlendControl: l.Registers.Context[GREG_MM_CB_BLEND0_CONTROL],
			RtClearWord0:   l.Registers.Context[GREG_MM_CB_COLOR0_CLEAR_WORD0],
			RtClearWord1:   l.Registers.Context[GREG_MM_CB_COLOR0_CLEAR_WORD1],

			CullFront:             (l.Registers.Context[GREG_MM_PA_SU_SC_MODE_CNTL]>>0)&1 == 1,
			CullBack:              (l.Registers.Context[GREG_MM_PA_SU_SC_MODE_CNTL]>>1)&1 == 1,
			Face:                  (l.Registers.Context[GREG_MM_PA_SU_SC_MODE_CNTL]>>2)&1 == 1,
			PolyMode:              (l.Registers.Context[GREG_MM_PA_SU_SC_MODE_CNTL] >> 3) & 0x3,
			PolyModeFrontPtype:    (l.Registers.Context[GREG_MM_PA_SU_SC_MODE_CNTL] >> 5) & 0x7,
			PolyModeBackPtype:     (l.Registers.Context[GREG_MM_PA_SU_SC_MODE_CNTL] >> 8) & 0x7,
			PolyOffsetFrontEnable: (l.Registers.Context[GREG_MM_PA_SU_SC_MODE_CNTL]>>11)&1 == 1,
			PolyOffsetBackEnable:  (l.Registers.Context[GREG_MM_PA_SU_SC_MODE_CNTL]>>12)&1 == 1,
			PolyOffsetParaEnable:  (l.Registers.Context[GREG_MM_PA_SU_SC_MODE_CNTL]>>13)&1 == 1,
			ProvokingVertexLast:   (l.Registers.Context[GREG_MM_PA_SU_SC_MODE_CNTL]>>19)&1 == 1,

			RtFormat:               (l.Registers.Context[GREG_MM_CB_COLOR0_INFO] >> 2) & 0x1F,
			RtNumberType:           (l.Registers.Context[GREG_MM_CB_COLOR0_INFO] >> 8) & 0x7,
			RtCompSwap:             (l.Registers.Context[GREG_MM_CB_COLOR0_INFO] >> 11) & 0x3,
			RtLinearGeneral:        (l.Registers.Context[GREG_MM_CB_COLOR0_INFO]>>7)&1 == 1,
			RtFastClear:            (l.Registers.Context[GREG_MM_CB_COLOR0_INFO]>>13)&1 == 1,
			RtCompression:          (l.Registers.Context[GREG_MM_CB_COLOR0_INFO]>>14)&1 == 1,
			RtBlendClamp:           (l.Registers.Context[GREG_MM_CB_COLOR0_INFO]>>15)&1 == 1,
			RtBlendBypass:          (l.Registers.Context[GREG_MM_CB_COLOR0_INFO]>>16)&1 == 1,
			RtSimpleFloat:          (l.Registers.Context[GREG_MM_CB_COLOR0_INFO]>>17)&1 == 1,
			RtRoundMode:            (l.Registers.Context[GREG_MM_CB_COLOR0_INFO] >> 18) & 1,
			RtCmaskIsLinear:        (l.Registers.Context[GREG_MM_CB_COLOR0_INFO]>>19)&1 == 1,
			RtBlendOptDontRdDst:    (l.Registers.Context[GREG_MM_CB_COLOR0_INFO] >> 20) & 0x7,
			RtBlendOptDiscardPixel: (l.Registers.Context[GREG_MM_CB_COLOR0_INFO] >> 23) & 0x7,
			RtFmaskCompressionDis:  (l.Registers.Context[GREG_MM_CB_COLOR0_INFO]>>26)&1 == 1,

			DbZExportEnable:              (l.Registers.Context[GREG_MM_DB_SHADER_CONTROL]>>0)&1 == 1,
			DbStencilTestValExportEnable: (l.Registers.Context[GREG_MM_DB_SHADER_CONTROL]>>1)&1 == 1,
			DbStencilOpValExportEnable:   (l.Registers.Context[GREG_MM_DB_SHADER_CONTROL]>>2)&1 == 1,
			DbZOrder:                     (l.Registers.Context[GREG_MM_DB_SHADER_CONTROL] >> 4) & 0x3,
			DbKillEnable:                 (l.Registers.Context[GREG_MM_DB_SHADER_CONTROL]>>6)&1 == 1,
			DbCoverageToMaskEnable:       (l.Registers.Context[GREG_MM_DB_SHADER_CONTROL]>>7)&1 == 1,
			DbMaskExportEnable:           (l.Registers.Context[GREG_MM_DB_SHADER_CONTROL]>>8)&1 == 1,
			DbExecOnHierFail:             (l.Registers.Context[GREG_MM_DB_SHADER_CONTROL]>>9)&1 == 1,
			DbExecOnNoop:                 (l.Registers.Context[GREG_MM_DB_SHADER_CONTROL]>>10)&1 == 1,
			DbAlphaToMaskDisable:         (l.Registers.Context[GREG_MM_DB_SHADER_CONTROL]>>11)&1 == 1,
			DbDepthBeforeShader:          (l.Registers.Context[GREG_MM_DB_SHADER_CONTROL]>>12)&1 == 1,
			DbConservativeZExport:        (l.Registers.Context[GREG_MM_DB_SHADER_CONTROL] >> 13) & 0x3,

			DbDepthControl:    l.Registers.Context[GREG_MM_DB_DEPTH_CONTROL],
			DbDepthClearValue: l.Registers.Context[GREG_MM_DB_DEPTH_CLEAR],
			DbDepthSize:       l.Registers.Context[GREG_MM_DB_DEPTH_SIZE],
			DbZWriteBase: func() uint32 {
				if l.Registers.Context[GREG_MM_DB_Z_WRITE_BASE] != 0 {
					return l.Registers.Context[GREG_MM_DB_Z_WRITE_BASE]
				}
				return l.Registers.Context[GREG_MM_DB_Z_READ_BASE]
			}(),
			DbZFormat: l.Registers.Context[GREG_MM_DB_Z_INFO] & 0x3,

			DbStencilControl:    l.Registers.Context[GREG_MM_DB_STENCIL_CONTROL],
			DbStencilRefMask:    l.Registers.Context[GREG_MM_DB_STENCILREFMASK],
			DbStencilRefMaskBf:  l.Registers.Context[GREG_MM_DB_STENCILREFMASK_BF],
			DbStencilClearValue: l.Registers.Context[GREG_MM_DB_STENCIL_CLEAR],

			DbDepthClearEnable:   (l.Registers.Context[GREG_MM_DB_RENDER_CONTROL]>>0)&1 == 1,
			DbStencilClearEnable: (l.Registers.Context[GREG_MM_DB_RENDER_CONTROL]>>1)&1 == 1,
			DbDepthCopy:          (l.Registers.Context[GREG_MM_DB_RENDER_CONTROL]>>2)&1 == 1,
			DbStencilCopy:        (l.Registers.Context[GREG_MM_DB_RENDER_CONTROL]>>3)&1 == 1,

			VpScissorEnable:    (l.Registers.Context[GREG_MM_PA_SC_MODE_CNTL_0]>>1)&1 == 1,
			WindowOffsetEnable: (l.Registers.Context[GREG_MM_PA_SU_SC_MODE_CNTL]>>16)&1 == 1,

			LineStippleEnable: (l.Registers.Context[GREG_MM_PA_SC_MODE_CNTL_0]>>2)&1 == 1,

			MsaaEnable:          (l.Registers.Context[GREG_MM_PA_SC_MODE_CNTL_0]>>0)&1 == 1,
			MsaaSampleLocations: l.Registers.Context[GREG_MM_PA_SC_AA_CONFIG] & 0x7,

			PsInControl:    l.Registers.Context[GREG_MM_SPI_PS_IN_CONTROL],
			PsInputAddress: l.Registers.Context[GREG_MM_SPI_PS_INPUT_ADDR],

			MultiPrimIbResetEnable: l.Registers.Context[GREG_MM_VGT_MULTI_PRIM_IB_RESET_EN]&1 == 1,
			MultiPrimIbResetIndex:  l.Registers.Context[GREG_MM_VGT_MULTI_PRIM_IB_RESET_INDX],

			UserDataHash: l.SnapshotUserData(),
		},
	}
	copy(bindPipeline.PsInputControls[:], l.Registers.Context[GREG_MM_SPI_PS_INPUT_CNTL_0:GREG_MM_SPI_PS_INPUT_CNTL_31+1])
	bindPipeline.VertexShaderAddress = bindPipeline.VertexShader.Address
	if address := l.PsGpuAddress(); address != 0 {
		bindPipeline.PixelShader = l.GetShader(GcnShaderStageFragment, address)
		bindPipeline.PixelShaderAddress = address
	}
	if address := l.HsGpuAddress(); address != 0 {
		bindPipeline.HullShader = l.GetShader(GcnShaderStageHull, address)
		bindPipeline.HullShaderAddress = address
	}
	if address := l.EsGpuAddress(); address != 0 {
		bindPipeline.EvalShader = l.GetShader(GcnShaderStageEvaluation, address)
		bindPipeline.EvalShaderAddress = address
	}
	if address := l.GsGpuAddress(); address != 0 {
		bindPipeline.GeometryShader = l.GetShader(GcnShaderStageGeometry, address)
		bindPipeline.GeometryShaderAddress = address
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

			VteControl:      l.Registers.Context[GREG_MM_PA_CL_VTE_CNTL],
			VpXScaleEnable:  (l.Registers.Context[GREG_MM_PA_CL_VTE_CNTL]>>0)&1 == 1,
			VpXOffsetEnable: (l.Registers.Context[GREG_MM_PA_CL_VTE_CNTL]>>1)&1 == 1,
			VpYScaleEnable:  (l.Registers.Context[GREG_MM_PA_CL_VTE_CNTL]>>2)&1 == 1,
			VpYOffsetEnable: (l.Registers.Context[GREG_MM_PA_CL_VTE_CNTL]>>3)&1 == 1,
			VpZScaleEnable:  (l.Registers.Context[GREG_MM_PA_CL_VTE_CNTL]>>4)&1 == 1,
			VpZOffsetEnable: (l.Registers.Context[GREG_MM_PA_CL_VTE_CNTL]>>5)&1 == 1,
			VtxXyFmt:        (l.Registers.Context[GREG_MM_PA_CL_VTE_CNTL]>>8)&1 == 1,
			VtxZFmt:         (l.Registers.Context[GREG_MM_PA_CL_VTE_CNTL]>>9)&1 == 1,
			VtxW0Fmt:        (l.Registers.Context[GREG_MM_PA_CL_VTE_CNTL]>>10)&1 == 1,

			ClipControl:   l.Registers.Context[GREG_MM_PA_CL_CLIP_CNTL],
			GbVertClipAdj: math.Float32frombits(l.Registers.Context[GREG_MM_PA_CL_GB_VERT_CLIP_ADJ]),
			GbVertDiscAdj: math.Float32frombits(l.Registers.Context[GREG_MM_PA_CL_GB_VERT_DISC_ADJ]),
			GbHorzClipAdj: math.Float32frombits(l.Registers.Context[GREG_MM_PA_CL_GB_HORZ_CLIP_ADJ]),
			GbHorzDiscAdj: math.Float32frombits(l.Registers.Context[GREG_MM_PA_CL_GB_HORZ_DISC_ADJ]),

			BlendRed:   l.Registers.Context[GREG_MM_CB_BLEND_RED],
			BlendGreen: l.Registers.Context[GREG_MM_CB_BLEND_GREEN],
			BlendBlue:  l.Registers.Context[GREG_MM_CB_BLEND_BLUE],
			BlendAlpha: l.Registers.Context[GREG_MM_CB_BLEND_ALPHA],

			ScissorTl: l.Registers.Context[GREG_MM_PA_SC_SCREEN_SCISSOR_TL],
			ScissorBr: l.Registers.Context[GREG_MM_PA_SC_SCREEN_SCISSOR_BR],

			VpScissorEnable: (l.Registers.Context[GREG_MM_PA_SC_MODE_CNTL_0]>>1)&1 == 1,
			VpScissorTl:     l.Registers.Context[GREG_MM_PA_SC_VPORT_SCISSOR_0_TL],
			VpScissorBr:     l.Registers.Context[GREG_MM_PA_SC_VPORT_SCISSOR_0_BR],

			GenericScissorTl: l.Registers.Context[GREG_MM_PA_SC_GENERIC_SCISSOR_TL],
			GenericScissorBr: l.Registers.Context[GREG_MM_PA_SC_GENERIC_SCISSOR_BR],

			WindowScissorTl:    l.Registers.Context[GREG_MM_PA_SC_WINDOW_SCISSOR_TL],
			WindowScissorBr:    l.Registers.Context[GREG_MM_PA_SC_WINDOW_SCISSOR_BR],
			WindowOffset:       l.Registers.Context[GREG_MM_PA_SC_WINDOW_OFFSET],
			WindowOffsetEnable: (l.Registers.Context[GREG_MM_PA_SU_SC_MODE_CNTL]>>16)&1 == 1,

			LineStippleRepeatCount: (l.Registers.Context[GREG_MM_PA_SU_LINE_STIPPLE_CNTL] >> 16) & 0xFF,
			LineStipplePattern:     l.Registers.Context[GREG_MM_PA_SU_LINE_STIPPLE_CNTL] & 0xFFFF,

			HardwareScreenOffset: l.Registers.Context[GREG_MM_PA_SU_HARDWARE_SCREEN_OFFSET],
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

			DbDepthClearEnable:   (l.Registers.Context[GREG_MM_DB_RENDER_CONTROL]>>0)&1 == 1,
			DbStencilClearEnable: (l.Registers.Context[GREG_MM_DB_RENDER_CONTROL]>>1)&1 == 1,
			DbDepthClearValue:    l.Registers.Context[GREG_MM_DB_DEPTH_CLEAR],
			DbStencilClearValue:  l.Registers.Context[GREG_MM_DB_STENCIL_CLEAR],

			UserDataHash: bindPipeline.UserDataHash,
		},
	}

	// Add to command stream.
	drawHash := draw.Hash()
	drawIndex, ok := stream.DrawsMap[bindHash]
	if !ok {
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
				color.Yellow.Sprintf("0x%X", bindPipeline.VertexShader.Address),
				color.Yellow.Sprintf("0x%X", bindPipeline.PixelShader.Address),
			)
		} else {
			logger.Printf("[%s] draw index auto (index_count=%s, prim=%s, rt=%s, vs=%s, ps=%s).\n",
				color.Green.Sprintf("PM4-%s", stream.Name),
				color.Green.Sprintf("%d", draw.IndexCount),
				color.Green.Sprintf("%d", draw.PrimType),
				color.Yellow.Sprintf("0x%X", bindPipeline.RtBase),
				color.Yellow.Sprintf("0x%X", bindPipeline.VertexShader.Address),
				color.Yellow.Sprintf("0x%X", bindPipeline.PixelShader.Address),
			)
		}
	}
}
