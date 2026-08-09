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
	// Constants.
	BlockContextIdFalse SpirvId = iota
	BlockContextIdTrue
	BlockContextIdTypeZeroVec4
	BlockContextIdTypeGlsl
	BlockContextIdTypeDebugPrintf

	// Scalar types.
	BlockContextIdTypeVoid
	BlockContextIdTypeBool
	BlockContextIdTypeV4Bool
	BlockContextIdTypeFloat
	BlockContextIdTypeInt
	BlockContextIdTypeUint
	BlockContextIdTypeUint8
	BlockContextIdTypeUint16
	BlockContextIdTypeUint64
	BlockContextIdTypeInt64

	// Vector & composite types.
	BlockContextIdTypeV2Float
	BlockContextIdTypeV3Float
	BlockContextIdTypeV4Float
	BlockContextIdTypeV2Int
	BlockContextIdTypeV3Int
	BlockContextIdTypeV4Int
	BlockContextIdTypeV2Uint
	BlockContextIdTypeV3Uint
	BlockContextIdTypeV4Uint
	BlockContextIdTypeStructUintUint

	// 1D images.
	BlockContextIdTypeImage1d
	BlockContextIdTypeSampledImage1d
	BlockContextIdPtrUniformSampledImage1d
	BlockContextIdStaticTextures1d
	BlockContextIdTypeStorageImage1d
	BlockContextIdPtrUniformStorageImage1d
	BlockContextIdStaticStorageTextures1d

	// 2D images.
	BlockContextIdTypeImage2d
	BlockContextIdTypeSampledImage2d
	BlockContextIdPtrUniformSampledImage2d
	BlockContextIdStaticTextures2d
	BlockContextIdTypeStorageImage2d
	BlockContextIdPtrUniformStorageImage2d
	BlockContextIdStaticStorageTextures2d

	// 3D images.
	BlockContextIdTypeImage3d
	BlockContextIdTypeSampledImage3d
	BlockContextIdPtrUniformSampledImage3d
	BlockContextIdStaticTextures3d
	BlockContextIdTypeStorageImage3d
	BlockContextIdPtrUniformStorageImage3d
	BlockContextIdStaticStorageTextures3d

	// 2D arrays.
	BlockContextIdTypeImage2dArray
	BlockContextIdTypeSampledImage2dArray
	BlockContextIdPtrUniformSampledImage2dArray
	BlockContextIdStaticTextures2dArray
	BlockContextIdTypeStorageImage2dArray
	BlockContextIdPtrUniformStorageImage2dArray
	BlockContextIdStaticStorageTextures2dArray

	// Pointers.
	BlockContextIdPtrPcPsbUint
	BlockContextIdPtrPcUint64
	BlockContextIdPtrPcUint
	BlockContextIdPtrPcFloat
	BlockContextIdPtrPsbUint
	BlockContextIdPtrPsbUint8
	BlockContextIdPtrPsbUint16
	BlockContextIdPtrPsbV2Uint
	BlockContextIdPtrPsbV3Uint
	BlockContextIdPtrPsbV4Uint
	BlockContextIdPtrFnUint

	// System variables & built-ins.
	BlockContextIdPcVar
	BlockContextIdAddressTranslationBuffer
	BlockContextIdSubgroupLocalInvocationId
	BlockContextIdVertexIndex
	BlockContextIdInstanceIndex
	BlockContextIdFragCoord
	BlockContextIdIsValidPixel
	BlockContextIdLdsArray
	BlockContextIdWorkgroupId
	BlockContextIdLocalInvocationId

	// Pipeline outputs.
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

	// Parameter inputs.
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
	BlockContextIdParamIn16
	BlockContextIdParamIn17
	BlockContextIdParamIn18
	BlockContextIdParamIn19
	BlockContextIdParamIn20
	BlockContextIdParamIn21
	BlockContextIdParamIn22
	BlockContextIdParamIn23
	BlockContextIdParamIn24
	BlockContextIdParamIn25
	BlockContextIdParamIn26
	BlockContextIdParamIn27
	BlockContextIdParamIn28
	BlockContextIdParamIn29
	BlockContextIdParamIn30
	BlockContextIdParamIn31

	// Parameter outputs.
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
	BlockContextIdParamOut16
	BlockContextIdParamOut17
	BlockContextIdParamOut18
	BlockContextIdParamOut19
	BlockContextIdParamOut20
	BlockContextIdParamOut21
	BlockContextIdParamOut22
	BlockContextIdParamOut23
	BlockContextIdParamOut24
	BlockContextIdParamOut25
	BlockContextIdParamOut26
	BlockContextIdParamOut27
	BlockContextIdParamOut28
	BlockContextIdParamOut29
	BlockContextIdParamOut30
	BlockContextIdParamOut31

	BlockContextIdContinueBlocks = SpirvId(10000)
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
	ConstIdUint11                  = SpirvId(11)
	ConstIdUint12                  = SpirvId(12)
	ConstIdUint13                  = SpirvId(13)
	ConstIdUint14                  = SpirvId(14)
	ConstIdUint15                  = SpirvId(15)
	ConstIdUint16                  = SpirvId(16)
	ConstIdUint19                  = SpirvId(19)
	ConstIdUint21                  = SpirvId(21)
	ConstIdUint23                  = SpirvId(23)
	ConstIdUint24                  = SpirvId(24)
	ConstIdUint30                  = SpirvId(30)
	ConstIdUint31                  = SpirvId(31)
	ConstIdUint32                  = SpirvId(32)
	ConstIdUint33                  = SpirvId(33)
	ConstIdUint62                  = SpirvId(62)
	ConstIdUint63                  = SpirvId(63)
	ConstIdUint127                 = SpirvId(127)
	ConstIdUint7F                  = ConstIdUint127
	ConstIdUint255                 = SpirvId(255)
	ConstIdUint256                 = SpirvId(256)
	ConstIdUint3FFF                = SpirvId(257)
	ConstIdUintFFFF                = SpirvId(258)
	ConstIdUint11111111            = SpirvId(259)
	ConstIdUintFFFFFFFF            = SpirvId(260)
	ConstIdUintFFFFFFFFC           = SpirvId(261)
	ConstId64Uint0                 = SpirvId(262)
	ConstId64Uint1                 = SpirvId(263)
	ConstId64Uint4                 = SpirvId(264)
	ConstId64Uint8                 = SpirvId(265)
	ConstId64Uint12                = SpirvId(266)
	ConstId64Uint32                = SpirvId(267)
	ConstId64UintNot3              = SpirvId(268)
	ConstId64UintOnionBaseAddress  = SpirvId(269)
	ConstId64UintGarlicBaseAddress = SpirvId(270)
	ConstIdFloat0                  = SpirvId(271)
	ConstIdFloat05                 = SpirvId(272)
	ConstIdFloat1                  = SpirvId(273)
	ConstIdFloat2                  = SpirvId(274)
	ConstIdFloat4                  = SpirvId(275)
	ConstIdFloat255                = SpirvId(276)
	ConstIdFloat65535              = SpirvId(277)
	ConstIdFloatMin                = SpirvId(278)
	ConstIdFloatMax                = SpirvId(279)
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
	PushConstantUserDataAddress         = 0
	PushConstantOnionMemoryBaseAddress  = 1
	PushConstantGarlicMemoryBaseAddress = 2
	PushConstantGdsMemoryBaseAddress    = 3
	PushConstantUserSgprCount           = 4
	PushConstantShaderRsrc2             = 5
	PushConstantVteControl              = 6
	PushConstantClipControl             = 7
	PushConstantGbHorzClipAdj           = 8
	PushConstantGbVertClipAdj           = 9
)
