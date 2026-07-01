package gpu

import (
	"hash/adler32"
	"unsafe"

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

// LiverpoolRegisters mirrors register banks on the Liverpool GPU.
type LiverpoolRegisters struct {
	System     [GcnRegBankSize]uint32
	Config     [GcnRegBankSize]uint32
	Shader     [GcnRegBankSize]uint32
	Context    [GcnRegBankSize]uint32
	UserConfig [GcnRegBankSize]uint32
}

type UserData [96]uint32

var GlobalUserDataSnapshots = map[uint32]UserData{}

const (
	UserDataOffsetVertex     = 0x0
	UserDataOffsetHull       = 0x10
	UserDataOffsetEvaluation = 0x20
	UserDataOffsetGeometry   = 0x30
	UserDataOffsetFragment   = 0x40
	UserDataOffsetCompute    = 0x50
)

var GcnStageToUserDataOffset = map[GcnShaderStage]uint32{
	GcnShaderStageVertex:     UserDataOffsetVertex,
	GcnShaderStageHull:       UserDataOffsetHull,
	GcnShaderStageEvaluation: UserDataOffsetEvaluation,
	GcnShaderStageGeometry:   UserDataOffsetGeometry,
	GcnShaderStageFragment:   UserDataOffsetFragment,
	GcnShaderStageCompute:    UserDataOffsetCompute,
}

func (l *Liverpool) SnapshotUserData() uint32 {
	var userData UserData
	copy(userData[UserDataOffsetVertex:], l.Registers.Shader[GREG_MM_SPI_SHADER_USER_DATA_VS_0:GREG_MM_SPI_SHADER_USER_DATA_VS_15+1])
	copy(userData[UserDataOffsetHull:], l.Registers.Shader[GREG_MM_SPI_SHADER_USER_DATA_HS_0:GREG_MM_SPI_SHADER_USER_DATA_HS_15+1])
	copy(userData[UserDataOffsetEvaluation:], l.Registers.Shader[GREG_MM_SPI_SHADER_USER_DATA_ES_0:GREG_MM_SPI_SHADER_USER_DATA_ES_15+1])
	copy(userData[UserDataOffsetGeometry:], l.Registers.Shader[GREG_MM_SPI_SHADER_USER_DATA_GS_0:GREG_MM_SPI_SHADER_USER_DATA_GS_15+1])
	copy(userData[UserDataOffsetFragment:], l.Registers.Shader[GREG_MM_SPI_SHADER_USER_DATA_PS_0:GREG_MM_SPI_SHADER_USER_DATA_PS_15+1])
	copy(userData[UserDataOffsetCompute:], l.Registers.Shader[GREG_MM_COMPUTE_USER_DATA_0:GREG_MM_COMPUTE_USER_DATA_15+1])
	userDataBytes := unsafe.Slice((*byte)(unsafe.Pointer(&userData[0])), len(userData)*4)
	hash := adler32.Checksum(userDataBytes)
	if _, ok := GlobalUserDataSnapshots[hash]; !ok {
		GlobalUserDataSnapshots[hash] = userData
	}

	return hash
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
