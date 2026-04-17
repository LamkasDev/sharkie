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
	IndexBase        uintptr
	BaseVertexOffset uint32

	// Render target.
	RtBase   uint32
	RtPitch  uint32
	RtSlice  uint32
	RtView   uint32
	RtInfo   uint32
	RtAttrib uint32
	RtMask   uint32

	// Depth buffer.
	DbZInfo      uint32
	DbDepthSize  uint32
	DbZWriteBase uint32

	// Viewport.
	VpXScale  float32
	VpXOffset float32
	VpYScale  float32
	VpYOffset float32
	VpZScale  float32
	VpZOffset float32

	// Screen scissor.
	ScissorTl uint32
	ScissorBr uint32

	// Shader programs (vertex, tessalation hull & evaluation, geometry, pixel).
	VertexShPgmLo, VertexShPgmHi     uint32
	VertexShRsrc1, VertexShRsrc2     uint32
	HullShPgmLo, HullShPgmHi         uint32
	HullShRsrc1, HullShRsrc2         uint32
	EvalShPgmLo, EvalShPgmHi         uint32
	EvalShRsrc1, EvalShRsrc2         uint32
	GeometryShPgmLo, GeometryShPgmHi uint32
	GeometryShRsrc1, GeometryShRsrc2 uint32
	PixelShPgmLo, PixelShPgmHi       uint32
	PixelShRsrc1, PixelShRsrc2       uint32

	// Pointers to parsed shader programs.
	VertexShader   *GcnShader
	HullShader     *GcnShader
	EvalShader     *GcnShader
	GeometryShader *GcnShader
	PixelShader    *GcnShader

	// Snapshots of register states.
	ConstRamHash uint32
	UserDataHash uint32
}

// RtGpuAddress returns the 40-bit GPU address of the render target surface.
func (d *LiverpoolDrawCall) RtGpuAddress() uintptr { return uintptr(d.RtBase) << 8 }

// RtPitchPixels returns pitch in pixels from the raw CB_COLOR0_PITCH word.
func (d *LiverpoolDrawCall) RtPitchPixels() uint32 { return ((d.RtPitch & 0x7FF) + 1) * 8 }

// VsGpuAddress returns the full vertex shader GPU address.
func (d *LiverpoolDrawCall) VsGpuAddress() uintptr {
	return (uintptr(d.VertexShPgmLo) | uintptr(d.VertexShPgmHi)<<32) << 8
}

// PsGpuAddress returns the full hull shader GPU address.
func (d *LiverpoolDrawCall) HsGpuAddress() uintptr {
	return (uintptr(d.HullShPgmLo) | uintptr(d.HullShPgmHi)<<32) << 8
}

// PsGpuAddress returns the full evaluation shader GPU address.
func (d *LiverpoolDrawCall) EsGpuAddress() uintptr {
	return (uintptr(d.EvalShPgmLo) | uintptr(d.EvalShPgmHi)<<32) << 8
}

// PsGpuAddress returns the full geometry shader GPU address.
func (d *LiverpoolDrawCall) GsGpuAddress() uintptr {
	return (uintptr(d.GeometryShPgmLo) | uintptr(d.GeometryShPgmHi)<<32) << 8
}

// PsGpuAddress returns the full pixel shader GPU address.
func (d *LiverpoolDrawCall) PsGpuAddress() uintptr {
	return (uintptr(d.PixelShPgmLo) | uintptr(d.PixelShPgmHi)<<32) << 8
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
		IndexBase:        l.DrawState.IndexBase,
		BaseVertexOffset: l.DrawState.BaseVertexOffset,

		RtBase:   l.Registers.Context[GREG_MM_CB_COLOR0_BASE],
		RtPitch:  l.Registers.Context[GREG_MM_CB_COLOR0_PITCH],
		RtSlice:  l.Registers.Context[GREG_MM_CB_COLOR0_SLICE],
		RtView:   l.Registers.Context[GREG_MM_CB_COLOR0_VIEW],
		RtInfo:   l.Registers.Context[GREG_MM_CB_COLOR0_INFO],
		RtAttrib: l.Registers.Context[GREG_MM_CB_COLOR0_ATTRIB],
		RtMask:   l.Registers.Context[GREG_MM_CB_TARGET_MASK],

		DbZInfo:      l.Registers.Context[GREG_MM_DB_Z_INFO],
		DbDepthSize:  l.Registers.Context[GREG_MM_DB_DEPTH_SIZE],
		DbZWriteBase: l.Registers.Context[GREG_MM_DB_Z_WRITE_BASE],

		VpXScale:  math.Float32frombits(l.Registers.Context[GREG_MM_PA_CL_VPORT_XSCALE]),
		VpXOffset: math.Float32frombits(l.Registers.Context[GREG_MM_PA_CL_VPORT_XOFFSET]),
		VpYScale:  math.Float32frombits(l.Registers.Context[GREG_MM_PA_CL_VPORT_YSCALE]),
		VpYOffset: math.Float32frombits(l.Registers.Context[GREG_MM_PA_CL_VPORT_YOFFSET]),
		VpZScale:  math.Float32frombits(l.Registers.Context[GREG_MM_PA_CL_VPORT_ZSCALE]),
		VpZOffset: math.Float32frombits(l.Registers.Context[GREG_MM_PA_CL_VPORT_ZOFFSET]),

		ScissorTl: l.Registers.Context[GREG_MM_PA_SC_SCREEN_SCISSOR_TL],
		ScissorBr: l.Registers.Context[GREG_MM_PA_SC_SCREEN_SCISSOR_BR],

		VertexShPgmLo:   l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_LO_VS],
		VertexShPgmHi:   l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_HI_VS],
		VertexShRsrc1:   l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_RSRC1_VS],
		VertexShRsrc2:   l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_RSRC2_VS],
		HullShPgmLo:     l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_LO_HS],
		HullShPgmHi:     l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_HI_HS],
		HullShRsrc1:     l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_RSRC1_HS],
		HullShRsrc2:     l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_RSRC2_HS],
		EvalShPgmLo:     l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_LO_ES],
		EvalShPgmHi:     l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_HI_ES],
		EvalShRsrc1:     l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_RSRC1_ES],
		EvalShRsrc2:     l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_RSRC2_ES],
		GeometryShPgmLo: l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_LO_GS],
		GeometryShPgmHi: l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_HI_GS],
		GeometryShRsrc1: l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_RSRC1_GS],
		GeometryShRsrc2: l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_RSRC2_GS],
		PixelShPgmLo:    l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_LO_PS],
		PixelShPgmHi:    l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_HI_PS],
		PixelShRsrc1:    l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_RSRC1_PS],
		PixelShRsrc2:    l.Registers.Shader[GREG_MM_SPI_SHADER_PGM_RSRC2_PS],

		ConstRamHash: l.SnapshotConstRam(),
		UserDataHash: l.SnapshotUserData(),
	}
	l.StateMutex.Unlock()

	drawCall.VertexShader = l.GetShader(GcnShaderStageVertex, drawCall.VsGpuAddress())
	if address := drawCall.HsGpuAddress(); address != 0 {
		drawCall.HullShader = l.GetShader(GcnShaderStageHull, address)
	}
	if address := drawCall.EsGpuAddress(); address != 0 {
		drawCall.EvalShader = l.GetShader(GcnShaderStageEvaluation, address)
	}
	if address := drawCall.GsGpuAddress(); address != 0 {
		drawCall.GeometryShader = l.GetShader(GcnShaderStageGeometry, address)
	}
	drawCall.PixelShader = l.GetShader(GcnShaderStageFragment, drawCall.PsGpuAddress())

	return drawCall
}
