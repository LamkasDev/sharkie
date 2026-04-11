package gpu

import (
	"unsafe"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

const LiverpoolConstRamSize = 0x8000

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

type UserData [96]uint32
