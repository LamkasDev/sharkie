package gpu

import (
	"unsafe"

	. "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/gpu/pm4"
	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
)

// LiverpoolConstRam is the hardware memory, not useful to us uwu.
type LiverpoolConstRam [LiverpoolConstRamSize]uint32

const LiverpoolConstRamSize = 0x8000

// LiverpoolCommandRing is thge command ring holding pending buffers.
type LiverpoolCommandRing struct {
	Pending []PM4IndirectBuffer
}

const LiverpoolCommandRingSize = unsafe.Sizeof(LiverpoolCommandRing{})

// OrderedIndirectBuffer preserves DCB/CCB submission order from sceGnmSubmitCommandBuffers.
// Each workload pair is submitted as CCB then DCB and must be walked in that sequence.
type OrderedIndirectBuffer struct {
	RingName string
	Buffer   PM4IndirectBuffer
}

// LiverpoolRegisters mirrors register banks on the Liverpool GPU.
type LiverpoolRegisters struct {
	System        [GcnRegBankSize]uint32
	Config        [GcnRegBankSize]uint32
	Shader        [GcnRegBankSize]uint32
	Context       [GcnRegBankSize]uint32
	UserConfig    [GcnRegBankSize]uint32
	CbColorExtent [8]uint32
	DbDepthExtent uint32
}

var ContextRegisterDefaults = [85]uint32{
	0x00000000, 0x80000000, 0x40004000, 0xdeadbeef, 0x00000000, 0x40004000, 0x00000000,
	0x40004000, 0x00000000, 0x40004000, 0x00000000, 0x40004000, 0xaa99aaaa, 0x00000000,
	0xdeadbeef, 0xdeadbeef, 0x80000000, 0x40004000, 0x00000000, 0x00000000, 0x80000000,
	0x40004000, 0x80000000, 0x40004000, 0x80000000, 0x40004000, 0x80000000, 0x40004000,
	0x80000000, 0x40004000, 0x80000000, 0x40004000, 0x80000000, 0x40004000, 0x80000000,
	0x40004000, 0x80000000, 0x40004000, 0x80000000, 0x40004000, 0x80000000, 0x40004000,
	0x80000000, 0x40004000, 0x80000000, 0x40004000, 0x80000000, 0x40004000, 0x80000000,
	0x40004000, 0x80000000, 0x40004000, 0x00000000, 0x3f800000, 0x00000000, 0x3f800000,
	0x00000000, 0x3f800000, 0x00000000, 0x3f800000, 0x00000000, 0x3f800000, 0x00000000,
	0x3f800000, 0x00000000, 0x3f800000, 0x00000000, 0x3f800000, 0x00000000, 0x3f800000,
	0x00000000, 0x3f800000, 0x00000000, 0x3f800000, 0x00000000, 0x3f800000, 0x00000000,
	0x3f800000, 0x00000000, 0x3f800000, 0x00000000, 0x3f800000, 0x00000000, 0x3f800000,
	0x2a00161a,
}

func (r *LiverpoolRegisters) SetDefaults() {
	clear(r.System[:])
	clear(r.Config[:])
	clear(r.Shader[:])
	clear(r.Context[:])
	clear(r.UserConfig[:])
	clear(r.CbColorExtent[:])
	r.DbDepthExtent = 0

	copy(r.Context[0x80:], ContextRegisterDefaults[:])

	r.Context[0x000d] = 0x40004000 // PA_SC_SCREEN_SCISSOR_BR
	r.Context[0x01b6] = 0x00000002 // SPI_PS_IN_CONTROL
	r.Context[0x0202] = 0x00000010 // CB_COLOR_CONTROL MODE=NORMAL (CLEAR_STATE leaves this 0 = DISABLE)
	r.Context[0x0204] = 0x00090000 // PA_CL_CLIP_CNTL (CLIP_DISABLE=1, DX_CLIP_SPACE_DEF=1)
	r.Context[0x0205] = 0x00000004 // PA_SU_SC_MODE_CNTL
	r.Context[0x0295] = 0x00000100 // VGT_ES_PER_GS
	r.Context[0x0296] = 0x00000080 // VGT_GS_PER_ES
	r.Context[0x0297] = 0x00000002 // VGT_GS_PER_VS
	r.Context[0x02aa] = 0x00001000 // IA_MULTI_VGT_PARAM
	r.Context[0x02f7] = 0x00001000 // PA_SC_LINE_CNTL
	r.Context[0x02f9] = 0x00000005 // PA_SU_VTX_CNTL
	r.Context[0x02fa] = 0x3f800000 // PA_CL_GB_VERT_CLIP_ADJ (1.0f)
	r.Context[0x02fb] = 0x3f800000 // PA_CL_GB_VERT_DISC_ADJ (1.0f)
	r.Context[0x02fc] = 0x3f800000 // PA_CL_GB_HORZ_CLIP_ADJ (1.0f)
	r.Context[0x02fd] = 0x3f800000 // PA_CL_GB_HORZ_DISC_ADJ (1.0f)
	r.Context[0x0316] = 0x0000000e // VGT_VERTEX_REUSE_BLOCK_CNTL
	r.Context[0x0317] = 0x00000010 // VGT_OUT_DEALLOC_CNTL
}

var GlobalUserDataSnapshots = map[uint32]spirvStructs.UserData{}
var userDataDedup = map[spirvStructs.UserData]uint32{}
var nextUserDataID uint32 = 1

func (l *Liverpool) SnapshotUserData() uint32 {
	var userData spirvStructs.UserData
	copy(userData[spirvStructs.UserDataOffsetVertex:], l.Registers.Shader[GREG_MM_SPI_SHADER_USER_DATA_VS_0:GREG_MM_SPI_SHADER_USER_DATA_VS_15+1])
	copy(userData[spirvStructs.UserDataOffsetHull:], l.Registers.Shader[GREG_MM_SPI_SHADER_USER_DATA_HS_0:GREG_MM_SPI_SHADER_USER_DATA_HS_15+1])
	copy(userData[spirvStructs.UserDataOffsetEvaluation:], l.Registers.Shader[GREG_MM_SPI_SHADER_USER_DATA_ES_0:GREG_MM_SPI_SHADER_USER_DATA_ES_15+1])
	copy(userData[spirvStructs.UserDataOffsetGeometry:], l.Registers.Shader[GREG_MM_SPI_SHADER_USER_DATA_GS_0:GREG_MM_SPI_SHADER_USER_DATA_GS_15+1])
	copy(userData[spirvStructs.UserDataOffsetFragment:], l.Registers.Shader[GREG_MM_SPI_SHADER_USER_DATA_PS_0:GREG_MM_SPI_SHADER_USER_DATA_PS_15+1])
	copy(userData[spirvStructs.UserDataOffsetCompute:], l.Registers.Shader[GREG_MM_COMPUTE_USER_DATA_0:GREG_MM_COMPUTE_USER_DATA_15+1])
	if id, ok := userDataDedup[userData]; ok {
		return id
	}

	id := nextUserDataID
	nextUserDataID++
	userDataDedup[userData] = id
	GlobalUserDataSnapshots[id] = userData
	return id
}

// LiverpoolDrawState tracks per-draw state decoded from non-register packets.
type LiverpoolDrawState struct {
	InstanceCount uint32
	IndexType     uint32  // 0 = 16-bit, 1 = 32-bit
	IndexBase     uintptr // host address of current index buffer
	IndexCount    uint32
	IndexOffset   uint32
	ConstRam      LiverpoolConstRam
}

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

// CsGpuAddress returns the full compute shader GPU address.
func (l *Liverpool) CsGpuAddress() uintptr {
	return (uintptr(l.Registers.Shader[GREG_MM_COMPUTE_PGM_LO]) | uintptr(l.Registers.Shader[GREG_MM_COMPUTE_PGM_HI])<<32) << 8
}

// USER_SGPR in SPI_SHADER_PGM_RSRC2_* is encoded in bits [5:1].
func DecodeUserSgprCount(rsrc2 uint32) uint32 {
	return (rsrc2 >> 1) & 0x1F
}

// DescribeDepthCompare returns a human-readable description of
// the ZFUNC field from DB_DEPTH_CONTROL (bits 6:4).
func DescribeDepthCompare(depthControl uint32) string {
	zf := (depthControl >> 4) & 0x7
	switch zf {
	case 0:
		return "NEVER"
	case 1:
		return "LESS"
	case 2:
		return "EQUAL"
	case 3:
		return "LEQUAL"
	case 4:
		return "GREATER"
	case 5:
		return "NOTEQUAL"
	case 6:
		return "GEQUAL"
	case 7:
		return "ALWAYS"
	default:
		return "???"
	}
}
