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
	SpvCapShader                         = uint32(1)
	SpvCapAddresses                      = uint32(4)
	SpvCapInt64                          = uint32(5)
	SpvCapSubgroupBallotKHR              = uint32(4423)
	SpvCapPhysicalStorageBufferAddresses = uint32(5341)
)

// SPIR-V built-in decorations.
const (
	SpvBuiltInSubgroupLocalInvocationId = uint32(41)
)

// SPIR-V addressing models.
const (
	SpvAddrModelLogical                 = uint32(0)
	SpvAddrModelPhysicalStorageBuffer64 = uint32(5348)
)

// SPIR-V memory models.
const (
	SpvMemModelGLSL450 = uint32(1)
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
	SpvStorageUniformConstant       = uint32(0)
	SpvStorageInput                 = uint32(1)
	SpvStorageUniform               = uint32(2)
	SpvStorageOutput                = uint32(3)
	SpvStorageFunction              = uint32(7)
	SpvStoragePushConstant          = uint32(9)
	SpvStoragePhysicalStorageBuffer = uint32(5349)
)

// SPIR-V scopes.
const (
	SpvScopeSubgroup = uint32(3)
)

// SPIR-V decorations.
const (
	SpvDecorationBlock          = uint32(2)
	SpvDecorationArrayStride    = uint32(6)
	SpvDecorationBuiltIn        = uint32(11)
	SpvDecorationAliased        = uint32(20)
	SpvDecorationLocation       = uint32(30)
	SpvDecorationDescriptorSet  = uint32(34)
	SpvDecorationOffset         = uint32(35)
	SpvDecorationAliasedPointer = uint32(5356)
)

// SPIR-V function, selection and loop control masks.
const (
	SpvFunctionControlNone  = uint32(0)
	SpvSelectionControlNone = uint32(0)
	SpvLoopControlNone      = uint32(0)
)

// SPIR-V memory access masks.
const (
	SpvMemoryAccessNone        = uint32(0)
	SpvMemoryAccessVolatile    = uint32(1)
	SpvMemoryAccessAligned     = uint32(2)
	SpvMemoryAccessNontemporal = uint32(4)
)

// SPIR-V Opcodes.
const (
	SpvOpNop                   = uint32(1)
	SpvOpExtension             = uint32(10)
	SpvOpExtInstImport         = uint32(11)
	SpvOpExtInst               = uint32(12)
	SpvOpMemoryModel           = uint32(14)
	SpvOpEntryPoint            = uint32(15)
	SpvOpExecutionMode         = uint32(16)
	SpvOpCapability            = uint32(17)
	SpvOpTypeVoid              = uint32(19)
	SpvOpTypeBool              = uint32(20)
	SpvOpTypeInt               = uint32(21)
	SpvOpTypeFloat             = uint32(22)
	SpvOpTypeVector            = uint32(23)
	SpvOpTypeArray             = uint32(28)
	SpvOpTypeStruct            = uint32(30)
	SpvOpTypePointer           = uint32(32)
	SpvOpTypeFunction          = uint32(33)
	SpvOpConstantTrue          = uint32(41)
	SpvOpConstantFalse         = uint32(42)
	SpvOpConstant              = uint32(43)
	SpvOpConstantComposite     = uint32(44)
	SpvOpVariable              = uint32(59)
	SpvOpFunction              = uint32(54)
	SpvOpFunctionEnd           = uint32(56)
	SpvOpLoad                  = uint32(61)
	SpvOpStore                 = uint32(62)
	SpvOpAccessChain           = uint32(65)
	SpvOpPtrAccessChain        = uint32(67)
	SpvOpDecorate              = uint32(71)
	SpvOpMemberDecorate        = uint32(72)
	SpvOpCompositeConstruct    = uint32(80)
	SpvOpCompositeExtract      = uint32(81)
	SpvOpUConvert              = uint32(113)
	SpvOpConvertUToPtr         = uint32(120)
	SpvOpBitcast               = uint32(124)
	SpvOpIAdd                  = uint32(128)
	SpvOpFAdd                  = uint32(129)
	SpvOpISub                  = uint32(130)
	SpvOpFSub                  = uint32(131)
	SpvOpIMul                  = uint32(132)
	SpvOpFMul                  = uint32(133)
	SpvOpLogicalOr             = uint32(166)
	SpvOpSelect                = uint32(169)
	SpvOpIEqual                = uint32(170)
	SpvOpINotEqual             = uint32(171)
	SpvOpUGreaterThan          = uint32(172)
	SpvOpFOrdEqual             = uint32(180)
	SpvOpFUnordNotEqual        = uint32(183)
	SpvOpShiftRightLogical     = uint32(194)
	SpvOpShiftLeftLogical      = uint32(196)
	SpvOpBitwiseOr             = uint32(197)
	SpvOpBitwiseXor            = uint32(198)
	SpvOpBitwiseAnd            = uint32(199)
	SpvOpLoopMerge             = uint32(246)
	SpvOpSelectionMerge        = uint32(247)
	SpvOpLabel                 = uint32(248)
	SpvOpBranch                = uint32(249)
	SpvOpBranchConditional     = uint32(250)
	SpvOpReturn                = uint32(253)
	SpvOpUnreachable           = uint32(255)
	SpvOpGroupNonUniformBallot = uint32(339)
)
