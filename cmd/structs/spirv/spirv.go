package spirv

import . "github.com/LamkasDev/sharkie/cmd/structs/gcn"

// Magic and version.
const (
	SpvMagic   = uint32(0x07230203)
	SpvVersion = uint32(0x00010300) // 1.3
	SpvGen     = uint32(0x534B4700) // "SKG\0"
)

// SPIR-V capabilities.
const (
	SpvCapShader = uint32(1)
)

// SPIR-V addressing and memory models.
const (
	SpvAddrModelLogical = uint32(iota)
	SpvMemModelGLSL450
)

// SPIR-V execution models.
const (
	SpvExecModelVertex = uint32(iota)
	SpvExecModelTesselationControl
	SpvExecModelTesselationEvaluation
	SpvExecModelGeometry
	SpvExecModelFragment
	SpvExecModelGLCompute
	SpvExecModelKernel
)

var GncStageToSpvExecModel = map[GcnShaderStage]uint32{
	GcnShaderStageVertex:     SpvExecModelVertex,
	GcnShaderStageHull:       SpvExecModelTesselationControl,
	GcnShaderStageEvaluation: SpvExecModelTesselationEvaluation,
	GcnShaderStageGeometry:   SpvExecModelGeometry,
	GcnShaderStageFragment:   SpvExecModelFragment,
	GcnShaderStageCompute:    SpvExecModelGLCompute,
}

// SPIR-V execution modes.
const (
	SpvExecModeOriginUpperLeft = uint32(7)
	SpvExecModeLocalSize       = uint32(17)
)

// SPIR-V storage classes.
const (
	SpvStorageUniformConstant = uint32(0)
	SpvStorageInput           = uint32(1)
	SpvStorageUniform         = uint32(2)
	SpvStorageOutput          = uint32(3)
	SpvStorageFunction        = uint32(7)
	SpvStoragePushConstant    = uint32(9)
)

// SPIR-V decorations.
const (
	SpvDecorationBlock       = uint32(2)
	SpvDecorationArrayStride = uint32(6)
	SpvDecorationBuiltIn     = uint32(11)
	SpvDecorationLocation    = uint32(30)
	SpvDecorationOffset      = uint32(35)
)

// SPIR-V function, selection and loop control masks.
const (
	SpvFunctionControlNone  = uint32(0)
	SpvSelectionControlNone = uint32(0)
	SpvLoopControlNone      = uint32(0)
)

// SPIR-V Opcodes.
const (
	SpvOpNop               = uint32(1)
	SpvOpExtInstImport     = uint32(11)
	SpvOpMemoryModel       = uint32(14)
	SpvOpEntryPoint        = uint32(15)
	SpvOpExecutionMode     = uint32(16)
	SpvOpCapability        = uint32(17)
	SpvOpTypeVoid          = uint32(19)
	SpvOpTypeBool          = uint32(20)
	SpvOpTypeInt           = uint32(21)
	SpvOpTypeFloat         = uint32(22)
	SpvOpTypeVector        = uint32(23)
	SpvOpTypeArray         = uint32(28)
	SpvOpTypeStruct        = uint32(30)
	SpvOpTypePointer       = uint32(32)
	SpvOpTypeFunction      = uint32(33)
	SpvOpConstantTrue      = uint32(41)
	SpvOpConstantFalse     = uint32(42)
	SpvOpConstant          = uint32(43)
	SpvOpConstantComposite = uint32(44)
	SpvOpVariable          = uint32(59)
	SpvOpDecorate          = uint32(71)
	SpvOpMemberDecorate    = uint32(72)
	SpvOpFunction          = uint32(54)
	SpvOpFunctionEnd       = uint32(56)
	SpvOpLoad              = uint32(61)
	SpvOpStore             = uint32(62)
	SpvOpLoopMerge         = uint32(246)
	SpvOpSelectionMerge    = uint32(247)
	SpvOpLabel             = uint32(248)
	SpvOpBranch            = uint32(249)
	SpvOpBranchConditional = uint32(250)
	SpvOpReturn            = uint32(253)
	SpvOpUnreachable       = uint32(255)
)
