package gpu

import (
	"math"

	"github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

// LiverpoolDrawState tracks per-draw state decoded from non-register packets.
type LiverpoolDrawState struct {
	InstanceCount    uint32
	IndexType        uint32  // 0 = 16-bit, 1 = 32-bit
	IndexBase        uintptr // host address of current index buffer
	IndexBufferSize  uint32
	BaseVertexOffset uint32
}

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

	// Shader programs.
	PixShPgmLo uint32
	PixShPgmHi uint32
	VerShPgmLo uint32
	VerShPgmHi uint32
}

// RtGpuAddress returns the 40-bit GPU address of the render target surface.
func (d *LiverpoolDrawCall) RtGpuAddress() uintptr { return uintptr(d.RtBase) << 8 }

// RtPitchPixels returns pitch in pixels from the raw CB_COLOR0_PITCH word.
func (d *LiverpoolDrawCall) RtPitchPixels() uint32 { return ((d.RtPitch & 0x7FF) + 1) * 8 }

// PSGPUAddress returns the full pixel shader GPU address.
func (d *LiverpoolDrawCall) PSGPUAddress() uintptr {
	return (uintptr(d.PixShPgmLo) | uintptr(d.PixShPgmHi)<<32) << 8
}

// VSGPUAddress returns the full vertex shader GPU address.
func (d *LiverpoolDrawCall) VSGPUAddress() uintptr {
	return (uintptr(d.VerShPgmLo) | uintptr(d.VerShPgmHi)<<32) << 8
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

// snapshotDrawCall captures the current register & draw state into a LiverpoolDrawCall.
func (l *Liverpool) snapshotDrawCall(vertexCount uint32, isIndexed bool) LiverpoolDrawCall {
	instanceCount := l.DrawState.InstanceCount
	if instanceCount == 0 {
		instanceCount = 1
	}

	return LiverpoolDrawCall{
		PrimType:      l.Registers.UserConfig[gcn.GREG_MM_VGT_PRIMITIVE_TYPE__CI__VI],
		VertexCount:   vertexCount,
		InstanceCount: instanceCount,
		IsIndexed:     isIndexed,

		IndexCount:       l.DrawState.IndexBufferSize,
		IndexType:        l.DrawState.IndexType,
		IndexBase:        l.DrawState.IndexBase,
		BaseVertexOffset: l.DrawState.BaseVertexOffset,

		RtBase:   l.Registers.Context[gcn.GREG_MM_CB_COLOR0_BASE],
		RtPitch:  l.Registers.Context[gcn.GREG_MM_CB_COLOR0_PITCH],
		RtSlice:  l.Registers.Context[gcn.GREG_MM_CB_COLOR0_SLICE],
		RtView:   l.Registers.Context[gcn.GREG_MM_CB_COLOR0_VIEW],
		RtInfo:   l.Registers.Context[gcn.GREG_MM_CB_COLOR0_INFO],
		RtAttrib: l.Registers.Context[gcn.GREG_MM_CB_COLOR0_ATTRIB],
		RtMask:   l.Registers.Context[gcn.GREG_MM_CB_TARGET_MASK],

		DbZInfo:      l.Registers.Context[gcn.GREG_MM_DB_Z_INFO],
		DbDepthSize:  l.Registers.Context[gcn.GREG_MM_DB_DEPTH_SIZE],
		DbZWriteBase: l.Registers.Context[gcn.GREG_MM_DB_Z_WRITE_BASE],

		VpXScale:  math.Float32frombits(l.Registers.Context[gcn.GREG_MM_PA_CL_VPORT_XSCALE]),
		VpXOffset: math.Float32frombits(l.Registers.Context[gcn.GREG_MM_PA_CL_VPORT_XOFFSET]),
		VpYScale:  math.Float32frombits(l.Registers.Context[gcn.GREG_MM_PA_CL_VPORT_YSCALE]),
		VpYOffset: math.Float32frombits(l.Registers.Context[gcn.GREG_MM_PA_CL_VPORT_YOFFSET]),
		VpZScale:  math.Float32frombits(l.Registers.Context[gcn.GREG_MM_PA_CL_VPORT_ZSCALE]),
		VpZOffset: math.Float32frombits(l.Registers.Context[gcn.GREG_MM_PA_CL_VPORT_ZOFFSET]),

		ScissorTl: l.Registers.Context[gcn.GREG_MM_PA_SC_SCREEN_SCISSOR_TL],
		ScissorBr: l.Registers.Context[gcn.GREG_MM_PA_SC_SCREEN_SCISSOR_BR],

		PixShPgmLo: l.Registers.Shader[gcn.GREG_MM_SPI_SHADER_PGM_LO_PS],
		PixShPgmHi: l.Registers.Shader[gcn.GREG_MM_SPI_SHADER_PGM_HI_PS],
		VerShPgmLo: l.Registers.Shader[gcn.GREG_MM_SPI_SHADER_PGM_LO_VS],
		VerShPgmHi: l.Registers.Shader[gcn.GREG_MM_SPI_SHADER_PGM_HI_VS],
	}
}
