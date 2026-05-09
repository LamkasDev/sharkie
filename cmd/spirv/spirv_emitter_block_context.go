package spirv

import (
	"fmt"

	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

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
	BlockContextIdTypeV2Uint
	BlockContextIdTypeV3Uint
	BlockContextIdTypeV4Uint
	BlockContextIdTypeSampledImage
	BlockContextIdPtrUniformSampledImage
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

type SpirvBlockContext struct {
	Stage    GcnShaderStage
	Address  uintptr
	LabelIds []SpirvId
	Ids      map[SpirvId]SpirvUsedId
	ConstIds map[SpirvId]SpirvUsedId

	GcnSgprArrayId SpirvId
	GcnVgprArrayId SpirvId
	GcnSpecialIds  [27]SpirvUsedId
	GcnConstIds    [120]SpirvUsedId
	GcnConditionId SpirvId
	Resources      []SpirvShaderResource
}

func (ctx *SpirvBlockContext) GetLabelId(i int) SpirvId {
	id := ctx.LabelIds[i]
	if id == 0 {
		panic(fmt.Sprintf("label id %d is zero", i))
	}

	return id
}

func (ctx *SpirvBlockContext) GetId(i SpirvId) SpirvId {
	id := ctx.Ids[i]
	if id.Id == 0 {
		panic(fmt.Sprintf("id %d is zero", i))
	}
	id.Used = true
	ctx.Ids[i] = id

	return id.Id
}

func (ctx *SpirvBlockContext) GetConstId(i SpirvId) SpirvId {
	id := ctx.ConstIds[i]
	if id.Id == 0 {
		panic(fmt.Sprintf("const id %d is zero", i))
	}
	id.Used = true
	ctx.ConstIds[i] = id

	return id.Id
}

func (ctx *SpirvBlockContext) GetGcnSgprPtr(b *SpvBuilder, reg uint32) SpirvId {
	idPtrFnUint := ctx.GetId(BlockContextIdPtrFnUint)
	return b.EmitAccessChain(idPtrFnUint, ctx.GcnSgprArrayId, ctx.GetConstId(ConstIdUint0+SpirvId(reg)))
}

func (ctx *SpirvBlockContext) GetGcnSgprId(b *SpvBuilder, reg uint32) SpirvId {
	return b.EmitLoad(ctx.GetId(BlockContextIdTypeUint), ctx.GetGcnSgprPtr(b, reg))
}

func (ctx *SpirvBlockContext) SetGcnSgprId(b *SpvBuilder, reg uint32, value SpirvId) {
	b.EmitStore(ctx.GetGcnSgprPtr(b, reg), value)
}

func (ctx *SpirvBlockContext) GetGcnVgprPtr(b *SpvBuilder, reg uint32) SpirvId {
	idPtrFnUint := ctx.GetId(BlockContextIdPtrFnUint)
	return b.EmitAccessChain(idPtrFnUint, ctx.GcnVgprArrayId, ctx.GetConstId(ConstIdUint0+SpirvId(reg)))
}

func (ctx *SpirvBlockContext) GetGcnVgprId(b *SpvBuilder, reg uint32) SpirvId {
	return b.EmitLoad(ctx.GetId(BlockContextIdTypeUint), ctx.GetGcnVgprPtr(b, reg))
}

func (ctx *SpirvBlockContext) SetGcnVgprId(b *SpvBuilder, reg uint32, value SpirvId) {
	b.EmitStore(ctx.GetGcnVgprPtr(b, reg), value)
}

func (ctx *SpirvBlockContext) GetTexelBufferVariable(binding uint32) SpirvId {
	return ctx.GetId(BlockContextIdTexelBuffer0 + SpirvId(binding))
}

func (ctx *SpirvBlockContext) GetGcnSpecialId(specialId SpirvId) SpirvId {
	id := ctx.GcnSpecialIds[specialId]
	if id.Id == 0 {
		panic(fmt.Sprintf("gcn special id %d is zero", specialId))
	}
	id.Used = true
	ctx.GcnSpecialIds[specialId] = id

	return id.Id
}

func (ctx *SpirvBlockContext) GetGcnConstId(constId SpirvId) SpirvId {
	id := ctx.GcnConstIds[constId]
	if id.Id == 0 {
		panic(fmt.Sprintf("gcn const id %d is zero", constId))
	}
	id.Used = true
	ctx.GcnConstIds[constId] = id

	return id.Id
}

func (ctx *SpirvBlockContext) EmitDebugPrintRegisters(b *SpvBuilder) {
	formatId := b.EmitString("Vertex %d: SGPRs: %08x %08x %08x %08x %08x %08x %08x %08x %08x %08x %08x %08x %08x %08x %08x %08x\n" +
		"VGPRs: %08x %08x %08x %08x %08x %08x %08x %08x %08x %08x %08x %08x %08x %08x %08x %08x\n")
	vertexIndexId := b.EmitLoad(ctx.GetId(BlockContextIdTypeUint), ctx.GetId(BlockContextIdVertexIndex))

	args := make([]SpirvId, 0, 33)
	args = append(args, formatId, vertexIndexId)
	for i := uint32(0); i < 16; i++ {
		args = append(args, ctx.GetGcnSgprId(b, i))
	}
	for i := uint32(0); i < 16; i++ {
		args = append(args, ctx.GetGcnVgprId(b, i))
	}

	b.EmitExtInst(ctx.GetId(BlockContextIdTypeVoid), ctx.GetId(BlockContextIdDebugPrintf), 1, args...)
}

// EmitDebugPrintfLane emits a debug printf only for a specific subgroup lane.
func (ctx *SpirvBlockContext) EmitDebugPrintfLane(b *SpvBuilder, lane uint32, format string, args ...SpirvId) {
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	typeBool := ctx.GetId(BlockContextIdTypeBool)
	idLane := b.EmitLoad(typeUint, ctx.GetId(BlockContextIdSubgroupLocalInvocationId))
	isLane := b.EmitIEqual(typeBool, idLane, ctx.GetConstId(ConstIdUint0+SpirvId(lane)))

	thenLabel := b.AllocId()
	mergeLabel := b.AllocId()
	b.EmitSelectionMerge(mergeLabel, spec.SpvSelectionControlNone)
	b.EmitBranchConditional(isLane, thenLabel, mergeLabel)

	b.EmitLabel(thenLabel)
	ins := make([]SpirvId, 0, len(args)+1)
	ins = append(ins, b.EmitString(format))
	ins = append(ins, args...)
	b.EmitExtInst(ctx.GetId(BlockContextIdTypeVoid), ctx.GetId(BlockContextIdDebugPrintf), 1, ins...)
	b.EmitBranch(mergeLabel)

	b.EmitLabel(mergeLabel)
}
