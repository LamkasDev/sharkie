package gpu

import (
	"math"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

// LiverpoolDrawState tracks per-draw state decoded from non-register packets.
type LiverpoolDrawState struct {
	InstanceCount    uint32
	IndexType        uint32  // 0 = 16-bit, 1 = 32-bit
	IndexBase        uintptr // host address of current index buffer
	IndexBufferSize  uint32
	BaseVertexOffset uint32
	ConstRam         LiverpoolConstRam
}

type LiverpoolConstRam [LiverpoolConstRamSize]uint32

// LiverpoolDrawCall is a snapshot of GPU state needed to issue a single draw.
type LiverpoolDrawCall struct {
	// Primitive types.
	PrimType      uint32
	VertexCount   uint32
	InstanceCount uint32
	IsIndexed     bool

	// Indexed draw parameters.
	IndexCount       uint32
	IndexType        uint32
	IndexBaseAddress uintptr
	BaseVertexOffset uint32

	// Render target.
	RtBase         uint32
	RtPitch        uint32
	RtSlice        uint32
	RtView         uint32
	RtInfo         uint32
	RtAttrib       uint32
	RtTargetMask   uint32
	RtColorControl uint32
	RtBlendControl uint32

	// Render target info fields (decoded from RtInfo/CB_COLOR0_INFO).
	RtFormat               uint32
	RtNumberType           uint32
	RtCompSwap             uint32
	RtLinearGeneral        bool
	RtFastClear            bool
	RtCompression          bool
	RtBlendClamp           bool
	RtBlendBypass          bool
	RtSimpleFloat          bool
	RtRoundMode            uint32
	RtCmaskIsLinear        bool
	RtBlendOptDontRdDst    uint32
	RtBlendOptDiscardPixel uint32
	RtFmaskCompressionDis  bool

	// Shader control fields (decoded from DbShaderControl/DB_SHADER_CONTROL).
	DbZExportEnable              bool
	DbStencilTestValExportEnable bool
	DbStencilOpValExportEnable   bool
	DbZOrder                     uint32
	DbKillEnable                 bool
	DbCoverageToMaskEnable       bool
	DbMaskExportEnable           bool
	DbExecOnHierFail             bool
	DbExecOnNoop                 bool
	DbAlphaToMaskDisable         bool
	DbDepthBeforeShader          bool
	DbConservativeZExport        uint32

	// Blend constants.
	BlendRed   uint32
	BlendGreen uint32
	BlendBlue  uint32
	BlendAlpha uint32

	// Depth buffer.
	DbDepthControl   uint32
	DbStencilControl uint32
	DbZInfo          uint32
	DbDepthSize      uint32
	DbZWriteBase     uint32

	// Viewport.
	VpXScale  float32
	VpXOffset float32
	VpYScale  float32
	VpYOffset float32
	VpZScale  float32
	VpZOffset float32
	VpZMin    float32
	VpZMax    float32

	// Screen scissor.
	ScissorTl uint32
	ScissorBr uint32

	// Shader programs parameters.
	VertexShRsrc1, VertexShRsrc2     uint32
	HullShRsrc1, HullShRsrc2         uint32
	EvalShRsrc1, EvalShRsrc2         uint32
	GeometryShRsrc1, GeometryShRsrc2 uint32
	PixelShRsrc1, PixelShRsrc2       uint32

	// Pointers to parsed shader programs.
	VertexShader   *GcnShader
	HullShader     *GcnShader
	EvalShader     *GcnShader
	GeometryShader *GcnShader
	PixelShader    *GcnShader

	// Snapshots of register states.
	UserDataHash uint32
}

// RtGpuAddress returns the 40-bit GPU address of the render target surface.
func (d *LiverpoolDrawCall) RtGpuAddress() uintptr { return uintptr(d.RtBase) << 8 }

// RtPitchPixels returns pitch in pixels from the raw CB_COLOR0_PITCH word.
func (d *LiverpoolDrawCall) RtPitchPixels() uint32 { return ((d.RtPitch & 0x7FF) + 1) * 8 }

// VsGpuAddress returns the full vertex shader GPU address.
func (l *Liverpool) VsGpuAddress() uintptr {
	return (uintptr(l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_LO_VS]) | uintptr(l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_HI_VS])<<32) << 8
}

// PsGpuAddress returns the full pixel shader GPU address.
func (l *Liverpool) PsGpuAddress() uintptr {
	return (uintptr(l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_LO_PS]) | uintptr(l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_HI_PS])<<32) << 8
}

// HsGpuAddress returns the full hull shader GPU address.
func (l *Liverpool) HsGpuAddress() uintptr {
	return (uintptr(l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_LO_HS]) | uintptr(l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_HI_HS])<<32) << 8
}

// EsGpuAddress returns the full evaluation shader GPU address.
func (l *Liverpool) EsGpuAddress() uintptr {
	return (uintptr(l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_LO_ES]) | uintptr(l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_HI_ES])<<32) << 8
}

// GsGpuAddress returns the full geometry shader GPU address.
func (l *Liverpool) GsGpuAddress() uintptr {
	return (uintptr(l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_LO_GS]) | uintptr(l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_HI_GS])<<32) << 8
}

// VpWidth / VpHeight returns pixel dimensions from the viewport XY scale.
func (d *LiverpoolDrawCall) VpWidth() uint32  { return uint32(math.Abs(float64(d.VpXScale)) * 2) }
func (d *LiverpoolDrawCall) VpHeight() uint32 { return uint32(math.Abs(float64(d.VpYScale)) * 2) }

// ScissorRect returns dimensions from the packed scissor registers.
func (d *LiverpoolDrawCall) ScissorRect() (x, y, width, height int) {
	x = int(d.ScissorTl & 0x7FFF)
	y = int((d.ScissorTl >> 16) & 0x7FFF)
	x1 := int(d.ScissorBr & 0x7FFF)
	y1 := int((d.ScissorBr >> 16) & 0x7FFF)

	return x, y, x1 - x, y1 - y
}

// USER_SGPR in SPI_SHADER_PGM_RSRC2_* is encoded in bits [5:1].
func DecodeUserSgprCount(rsrc2 uint32) uint32 {
	return (rsrc2 >> 1) & 0x1F
}

// NewDrawCall captures the current register & draw state into a LiverpoolDrawCall.
func (l *Liverpool) NewDrawCall(vertexCount uint32, isIndexed bool) LiverpoolDrawCall {
	l.StateMutex.Lock()
	drawCall := LiverpoolDrawCall{
		PrimType:      l.Registers.UserConfig[GREG_MM_VGT_PRIMITIVE_TYPE__CI__VI],
		VertexCount:   vertexCount,
		InstanceCount: max(l.DrawState.InstanceCount, 1),
		IsIndexed:     isIndexed,

		IndexCount:       l.DrawState.IndexBufferSize,
		IndexType:        l.DrawState.IndexType,
		IndexBaseAddress: l.DrawState.IndexBase,
		BaseVertexOffset: l.DrawState.BaseVertexOffset,

		RtBase:         l.Registers.Context[GREG_MM_CB_COLOR0_BASE],
		RtPitch:        l.Registers.Context[GREG_MM_CB_COLOR0_PITCH],
		RtSlice:        l.Registers.Context[GREG_MM_CB_COLOR0_SLICE],
		RtView:         l.Registers.Context[GREG_MM_CB_COLOR0_VIEW],
		RtInfo:         l.Registers.Context[GREG_MM_CB_COLOR0_INFO],
		RtAttrib:       l.Registers.Context[GREG_MM_CB_COLOR0_ATTRIB],
		RtTargetMask:   l.Registers.Context[GREG_MM_CB_TARGET_MASK],
		RtColorControl: l.Registers.Context[GREG_MM_CB_COLOR_CONTROL],
		RtBlendControl: l.Registers.Context[GREG_MM_CB_BLEND0_CONTROL],

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

		BlendRed:   l.Registers.Context[GREG_MM_CB_BLEND_RED],
		BlendGreen: l.Registers.Context[GREG_MM_CB_BLEND_GREEN],
		BlendBlue:  l.Registers.Context[GREG_MM_CB_BLEND_BLUE],
		BlendAlpha: l.Registers.Context[GREG_MM_CB_BLEND_ALPHA],

		DbDepthControl:   l.Registers.Context[GREG_MM_DB_DEPTH_CONTROL],
		DbStencilControl: l.Registers.Context[GREG_MM_DB_STENCIL_CONTROL],
		DbZInfo:          l.Registers.Context[GREG_MM_DB_Z_INFO],
		DbDepthSize:      l.Registers.Context[GREG_MM_DB_DEPTH_SIZE],
		DbZWriteBase:     l.Registers.Context[GREG_MM_DB_Z_WRITE_BASE],

		VpXScale:  math.Float32frombits(l.Registers.Context[GREG_MM_PA_CL_VPORT_XSCALE]),
		VpXOffset: math.Float32frombits(l.Registers.Context[GREG_MM_PA_CL_VPORT_XOFFSET]),
		VpYScale:  math.Float32frombits(l.Registers.Context[GREG_MM_PA_CL_VPORT_YSCALE]),
		VpYOffset: math.Float32frombits(l.Registers.Context[GREG_MM_PA_CL_VPORT_YOFFSET]),
		VpZScale:  math.Float32frombits(l.Registers.Context[GREG_MM_PA_CL_VPORT_ZSCALE]),
		VpZOffset: math.Float32frombits(l.Registers.Context[GREG_MM_PA_CL_VPORT_ZOFFSET]),
		VpZMin:    math.Float32frombits(l.Registers.Context[GREG_MM_PA_SC_VPORT_ZMIN_0]),
		VpZMax:    math.Float32frombits(l.Registers.Context[GREG_MM_PA_SC_VPORT_ZMAX_0]),

		ScissorTl: l.Registers.Context[GREG_MM_PA_SC_SCREEN_SCISSOR_TL],
		ScissorBr: l.Registers.Context[GREG_MM_PA_SC_SCREEN_SCISSOR_BR],

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

		UserDataHash: l.SnapshotUserData(),
	}
	l.StateMutex.Unlock()

	drawCall.VertexShader = l.GetShader(GcnShaderStageVertex, l.VsGpuAddress())
	drawCall.PixelShader = l.GetShader(GcnShaderStageFragment, l.PsGpuAddress())
	if address := l.HsGpuAddress(); address != 0 {
		drawCall.HullShader = l.GetShader(GcnShaderStageHull, address)
	}
	if address := l.EsGpuAddress(); address != 0 {
		drawCall.EvalShader = l.GetShader(GcnShaderStageEvaluation, address)
	}
	if address := l.GsGpuAddress(); address != 0 {
		drawCall.GeometryShader = l.GetShader(GcnShaderStageGeometry, address)
	}

	return drawCall
}
