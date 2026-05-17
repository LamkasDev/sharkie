package common

type SpirvId uint32

type SpirvUsedId struct {
	Id      SpirvId
	Used    bool
	Name    string
	Value   uint32
	Value64 uint64
}

const (
	BlockContextIdFalse SpirvId = iota
	BlockContextIdTrue
	BlockContextIdTypeBool
	BlockContextIdTypeFloat
	BlockContextIdTypeInt
	BlockContextIdTypeUint
	BlockContextIdTypeUint64
	BlockContextIdTypeInt64
	BlockContextIdTypeVoid
	BlockContextIdDebugPrintf
	BlockContextIdTypeV2Float
	BlockContextIdTypeV4Float
	BlockContextIdTypeV2Int
	BlockContextIdTypeV4Int
	BlockContextIdTypeV2Uint
	BlockContextIdTypeV3Uint
	BlockContextIdTypeV4Uint
	BlockContextIdTypeStructUintUint
	BlockContextIdTypeSampledImage
	BlockContextIdTypeImage
	BlockContextIdPtrUniformSampledImage
	BlockContextIdTypeStorageImage
	BlockContextIdPtrUniformStorageImage
	BlockContextIdPtrPcFloat
	BlockContextIdPtrPcPsbUint
	BlockContextIdPtrPcPsbUint64
	BlockContextIdPtrPcUint
	BlockContextIdPtrPcUint64
	BlockContextIdPtrPsbUint
	BlockContextIdPtrPsbV2Uint
	BlockContextIdPtrPsbV3Uint
	BlockContextIdPtrPsbV4Uint
	BlockContextIdPtrStorageUint
	BlockContextIdPtrFnUint
	BlockContextIdPosOut
	BlockContextIdFragDepthOut
	BlockContextIdColorOut0
	BlockContextIdColorOut1
	BlockContextIdColorOut2
	BlockContextIdColorOut3
	BlockContextIdColorOut4
	BlockContextIdColorOut5
	BlockContextIdColorOut6
	BlockContextIdColorOut7
	BlockContextIdParamOut0
	BlockContextIdParamOut1
	BlockContextIdParamOut2
	BlockContextIdParamOut3
	BlockContextIdParamOut4
	BlockContextIdParamOut5
	BlockContextIdParamOut6
	BlockContextIdParamOut7
	BlockContextIdParamOut8
	BlockContextIdParamOut9
	BlockContextIdParamOut10
	BlockContextIdParamOut11
	BlockContextIdParamOut12
	BlockContextIdParamOut13
	BlockContextIdParamOut14
	BlockContextIdParamOut15
	BlockContextIdParamIn0
	BlockContextIdParamIn1
	BlockContextIdParamIn2
	BlockContextIdParamIn3
	BlockContextIdParamIn4
	BlockContextIdParamIn5
	BlockContextIdParamIn6
	BlockContextIdParamIn7
	BlockContextIdParamIn8
	BlockContextIdParamIn9
	BlockContextIdParamIn10
	BlockContextIdParamIn11
	BlockContextIdParamIn12
	BlockContextIdParamIn13
	BlockContextIdParamIn14
	BlockContextIdParamIn15
	BlockContextIdZeroVec4
	BlockContextIdBindlessTextures
	BlockContextIdBindlessStorageTextures
	BlockContextIdPcVar
	BlockContextIdGlsl
	BlockContextIdSubgroupLocalInvocationId
	BlockContextIdVertexIndex
	BlockContextIdInstanceIndex
	BlockContextIdFragCoord
	BlockContextIdTypeImageBuffer
	BlockContextIdTexelBuffer0
	BlockContextIdTexelBuffer1
	BlockContextIdTexelBuffer2
	BlockContextIdTexelBuffer3
	BlockContextIdGlobalDescriptorMap
	BlockContextIdMissingResourceBuffer
	BlockContextIdWorkgroupId
	BlockContextIdLocalInvocationId
)

const (
	ConstIdUint0                   = SpirvId(0)
	ConstIdUint1                   = SpirvId(1)
	ConstIdUint2                   = SpirvId(2)
	ConstIdUint3                   = SpirvId(3)
	ConstIdUint4                   = SpirvId(4)
	ConstIdUint5                   = SpirvId(5)
	ConstIdUint6                   = SpirvId(6)
	ConstIdUint7                   = SpirvId(7)
	ConstIdUint8                   = SpirvId(8)
	ConstIdUint9                   = SpirvId(9)
	ConstIdUint10                  = SpirvId(10)
	ConstIdUint12                  = SpirvId(12)
	ConstIdUint15                  = SpirvId(15)
	ConstIdUint16                  = SpirvId(16)
	ConstIdUint19                  = SpirvId(19)
	ConstIdUint21                  = SpirvId(21)
	ConstIdUint23                  = SpirvId(23)
	ConstIdUint30                  = SpirvId(30)
	ConstIdUint31                  = SpirvId(31)
	ConstIdUint32                  = SpirvId(32)
	ConstIdUint33                  = SpirvId(33)
	ConstIdUint62                  = SpirvId(62)
	ConstIdUint63                  = SpirvId(63)
	ConstIdUint127                 = SpirvId(127)
	ConstIdUint7F                  = ConstIdUint127
	ConstIdUint256                 = SpirvId(256)
	ConstIdUint3FFF                = SpirvId(257)
	ConstIdUintFFFF                = SpirvId(258)
	ConstIdUint11111111            = SpirvId(259)
	ConstIdUintFFFFFFFF            = SpirvId(260)
	ConstId64Uint0                 = SpirvId(261)
	ConstId64Uint1                 = SpirvId(262)
	ConstId64Uint32                = SpirvId(263)
	ConstId64UintNot3              = SpirvId(264)
	ConstId64UintOnionBaseAddress  = SpirvId(265)
	ConstId64UintGarlicBaseAddress = SpirvId(266)
	ConstIdFloat0                  = SpirvId(267)
	ConstIdFloat1                  = SpirvId(268)
	ConstIdFloat05                 = SpirvId(269)
)

const (
	GcnSpecIdFlatScrLo SpirvId = iota
	GcnSpecIdFlatScrHi
	GcnSpecIdVccLo
	GcnSpecIdVccHi
	GcnSpecIdTbaLo
	GcnSpecIdTbaHi
	GcnSpecIdTmaLo
	GcnSpecIdTmaHi
	GcnSpecIdTtmp0
	GcnSpecIdTtmp1
	GcnSpecIdTtmp2
	GcnSpecIdTtmp3
	GcnSpecIdTtmp4
	GcnSpecIdTtmp5
	GcnSpecIdTtmp6
	GcnSpecIdTtmp7
	GcnSpecIdTtmp8
	GcnSpecIdTtmp9
	GcnSpecIdTtmp10
	GcnSpecIdTtmp11
	GcnSpecIdM0
	GcnSpecIdReserved
	GcnSpecIdExecLo
	GcnSpecIdExecHi
	GcnSpecIdVccz
	GcnSpecIdExecz
	GcnSpecIdScc
)

const (
	GcnConstId0          = SpirvId(0)
	GcnConstIdInt1       = SpirvId(1)
	GcnConstIdInt64      = SpirvId(64)
	GcnConstIdIntNeg1    = SpirvId(65)
	GcnConstIdIntNeg16   = SpirvId(80)
	GcnConstIdFloat05    = SpirvId(112)
	GcnConstIdFloatNeg05 = SpirvId(113)
	GcnConstIdFloat10    = SpirvId(114)
	GcnConstIdFloatNeg10 = SpirvId(115)
	GcnConstIdFloat20    = SpirvId(116)
	GcnConstIdFloatNeg20 = SpirvId(117)
	GcnConstIdFloat40    = SpirvId(118)
	GcnConstIdFloatNeg40 = SpirvId(119)
)

const (
	PushConstantUserDataAddress          = 0
	PushConstantOnionMemoryBaseAddress   = 1
	PushConstantGarlicMemoryBaseAddress  = 2
	PushConstantTexelBuffer0FormatSize   = 3
	PushConstantTexelBuffer1FormatSize   = 4
	PushConstantTexelBuffer2FormatSize   = 5
	PushConstantTexelBuffer3FormatSize   = 6
	PushConstantTexelBuffer0FormatStride = 7
	PushConstantTexelBuffer1FormatStride = 8
	PushConstantTexelBuffer2FormatStride = 9
	PushConstantTexelBuffer3FormatStride = 10
	PushConstantUserSgprCount            = 11
	PushConstantShaderRsrc2              = 12
)
