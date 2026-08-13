package structs

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
)

type UserData [96]uint32

const UserDataSize = unsafe.Sizeof(UserData{})

const (
	UserDataOffsetVertex     = 0x0
	UserDataOffsetHull       = 0x10
	UserDataOffsetEvaluation = 0x20
	UserDataOffsetGeometry   = 0x30
	UserDataOffsetFragment   = 0x40
	UserDataOffsetCompute    = 0x50
)

var GcnStageToUserDataOffset = map[gcn.GcnShaderStage]uint32{
	gcn.GcnShaderStageVertex:     UserDataOffsetVertex,
	gcn.GcnShaderStageHull:       UserDataOffsetHull,
	gcn.GcnShaderStageEvaluation: UserDataOffsetEvaluation,
	gcn.GcnShaderStageGeometry:   UserDataOffsetGeometry,
	gcn.GcnShaderStageFragment:   UserDataOffsetFragment,
	gcn.GcnShaderStageCompute:    UserDataOffsetCompute,
}
