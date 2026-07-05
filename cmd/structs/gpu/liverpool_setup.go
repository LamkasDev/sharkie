package gpu

import (
	"unsafe"

	spirvStructs "github.com/LamkasDev/sharkie/cmd/spirv/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
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
	System     [GcnRegBankSize]uint32
	Config     [GcnRegBankSize]uint32
	Shader     [GcnRegBankSize]uint32
	Context    [GcnRegBankSize]uint32
	UserConfig [GcnRegBankSize]uint32
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
	InstanceCount   uint32
	IndexType       uint32  // 0 = 16-bit, 1 = 32-bit
	IndexBase       uintptr // host address of current index buffer
	IndexBufferSize uint32
	ConstRam        LiverpoolConstRam
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
