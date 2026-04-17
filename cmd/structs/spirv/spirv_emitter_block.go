package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
	"github.com/LamkasDev/sharkie/cmd/structs/gpu"
)

type BlockContextId uint32

const (
	BlockContextIdFalse BlockContextId = iota
	BlockContextIdTrue
	BlockContextIdTypeBool
	BlockContextIdTypeFloat
	BlockContextIdTypeInt
	BlockContextIdTypeUint
	BlockContextIdTypeUint64
	BlockContextIdTypeInt64
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
	BlockContextIdPtrPsbUint
	BlockContextIdPtrPsbV2Uint
	BlockContextIdPtrPsbV3Uint
	BlockContextIdPtrPsbV4Uint
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
	BlockContextIdZeroVec4
	BlockContextIdBindlessTextures
	BlockContextIdPcVar
	BlockContextIdGlsl
	BlockContextIdSubgroupLocalInvocationId
	BlockContextIdVertexIndex
	BlockContextIdInstanceIndex
)

const (
	GcnSpecIdxFlatScrLo = 0
	GcnSpecIdxFlatScrHi = 1
	GcnSpecIdxVccLo     = 2
	GcnSpecIdxVccHi     = 3
	GcnSpecIdxTbaLo     = 4
	GcnSpecIdxTbaHi     = 5
	GcnSpecIdxTmaLo     = 6
	GcnSpecIdxTmaHi     = 7
	GcnSpecIdxTtmp0     = 8
	GcnSpecIdxTtmp11    = 19
	GcnSpecIdxM0        = 20
	GcnSpecIdxReserved  = 21
	GcnSpecIdxExecLo    = 22
	GcnSpecIdxExecHi    = 23
	GcnSpecIdxVccz      = 24
	GcnSpecIdxExecz     = 25
	GcnSpecIdxScc       = 26
)

const (
	GcnConstIdx0          = 0
	GcnConstIdxInt1       = 1
	GcnConstIdxInt64      = 64
	GcnConstIdxIntNeg1    = 65
	GcnConstIdxIntNeg16   = 80
	GcnConstIdxFloat05    = 112
	GcnConstIdxFloatNeg05 = 113
	GcnConstIdxFloat10    = 114
	GcnConstIdxFloatNeg10 = 115
	GcnConstIdxFloat20    = 116
	GcnConstIdxFloatNeg20 = 117
	GcnConstIdxFloat40    = 118
	GcnConstIdxFloatNeg40 = 119
)

const (
	ConstIdxUint0        = 0
	ConstIdxUint1        = 1
	ConstIdxUint2        = 2
	ConstIdxUint3        = 3
	ConstIdxUint4        = 4
	ConstIdxUint5        = 5
	ConstIdxUint6        = 6
	ConstIdxUint7        = 7
	ConstIdxUint8        = 8
	ConstIdxUint12       = 12
	ConstIdxUint15       = 15
	ConstIdxUint16       = 16
	ConstIdxUint19       = 19
	ConstIdxUint21       = 21
	ConstIdxUint23       = 23
	ConstIdxUint30       = 30
	ConstIdxUint31       = 31
	ConstIdxUint32       = 32
	ConstIdxUint33       = 33
	ConstIdxUint62       = 62
	ConstIdxUint63       = 63
	ConstIdxUint127      = 127
	ConstIdxUint7F       = ConstIdxUint127
	ConstIdxUint256      = 256
	ConstIdxUint3FFF     = 257
	ConstIdxUintFFFF     = 258
	ConstIdxUint11111111 = 259
	ConstIdxUintFFFFFFFF = 260
	ConstIdx64Uint0      = 261
	ConstIdx64Uint32     = 262
	ConstIdxFloat0       = 263
	ConstIdxFloat1       = 264
)

const (
	PushConstantTime            = 0
	PushConstant_               = 1
	PushConstantConstRamAddress = 2
	PushConstantUserDataAddress = 3
	PushConstantGarlicAddress   = 4
	PushConstantOnionAddress    = 5
)

type SpirvBlockContextUsedId struct {
	Id      uint32
	Used    bool
	Name    string
	Value   uint32
	Value64 uint64
}

type SpirvBlockContext struct {
	Stage    GcnShaderStage
	LabelIds []uint32
	Ids      map[BlockContextId]SpirvBlockContextUsedId
	ConstIds map[BlockContextId]SpirvBlockContextUsedId

	GcnSgprArrayId uint32
	GcnVgprArrayId uint32
	GcnSpecialIds  [27]SpirvBlockContextUsedId
	GcnConstIds    [120]SpirvBlockContextUsedId
	GcnConditionId uint32
}

func (ctx *SpirvBlockContext) GetLabelId(i int) uint32 {
	id := ctx.LabelIds[i]
	if id == 0 {
		panic(fmt.Sprintf("label id %d is zero", i))
	}

	return id
}

func (ctx *SpirvBlockContext) GetId(i BlockContextId) uint32 {
	id := ctx.Ids[i]
	if id.Id == 0 {
		panic(fmt.Sprintf("id %d is zero", i))
	}
	id.Used = true
	ctx.Ids[i] = id

	return id.Id
}

func (ctx *SpirvBlockContext) GetConstId(i BlockContextId) uint32 {
	id := ctx.ConstIds[i]
	if id.Id == 0 {
		panic(fmt.Sprintf("const id %d is zero", i))
	}
	id.Used = true
	ctx.ConstIds[i] = id

	return id.Id
}

func (ctx *SpirvBlockContext) GetGcnSgprPtr(b *SpvBuilder, reg uint32) uint32 {
	idPtrFnUint := ctx.GetId(BlockContextIdPtrFnUint)
	return b.EmitAccessChain(idPtrFnUint, ctx.GcnSgprArrayId, ctx.GetConstId(BlockContextId(ConstIdxUint0+reg)))
}

func (ctx *SpirvBlockContext) GetGcnSgprId(b *SpvBuilder, reg uint32) uint32 {
	return b.EmitLoad(ctx.GetId(BlockContextIdTypeUint), ctx.GetGcnSgprPtr(b, reg))
}

func (ctx *SpirvBlockContext) SetGcnSgprId(b *SpvBuilder, reg uint32, val uint32) {
	b.EmitStore(ctx.GetGcnSgprPtr(b, reg), val)
}

func (ctx *SpirvBlockContext) GetGcnVgprPtr(b *SpvBuilder, reg uint32) uint32 {
	idPtrFnUint := ctx.GetId(BlockContextIdPtrFnUint)
	return b.EmitAccessChain(idPtrFnUint, ctx.GcnVgprArrayId, ctx.GetConstId(BlockContextId(ConstIdxUint0+reg)))
}

func (ctx *SpirvBlockContext) GetGcnVgprId(b *SpvBuilder, reg uint32) uint32 {
	return b.EmitLoad(ctx.GetId(BlockContextIdTypeUint), ctx.GetGcnVgprPtr(b, reg))
}

func (ctx *SpirvBlockContext) SetGcnVgprId(b *SpvBuilder, reg uint32, val uint32) {
	b.EmitStore(ctx.GetGcnVgprPtr(b, reg), val)
}

func (ctx *SpirvBlockContext) GetGcnSpecialId(reg uint32) uint32 {
	id := ctx.GcnSpecialIds[reg]
	if id.Id == 0 {
		panic(fmt.Sprintf("gcn special id %d is zero", reg))
	}
	id.Used = true
	ctx.GcnSpecialIds[reg] = id

	return id.Id
}

func (ctx *SpirvBlockContext) GetGcnConstId(reg uint32) uint32 {
	id := ctx.GcnConstIds[reg]
	if id.Id == 0 {
		panic(fmt.Sprintf("gcn const id %d is zero", reg))
	}
	id.Used = true
	ctx.GcnConstIds[reg] = id

	return id.Id
}

// emitBlock emits the SPIR-V for a single block.
func emitBlock(b *SpvBuilder, block *GcnShaderCfgBlock, ctx *SpirvBlockContext) {
	// Start current block.
	b.EmitLabel(ctx.GetLabelId(block.Id))

	// Declare variables in entry block.
	if block.DwordOffset == 0 {
		// Load user data buffer address from the push constant.
		idPtrPsbUint := ctx.GetId(BlockContextIdPtrPsbUint)
		ptrBase := ctx.LoadPushConstantValue(b, PushConstantUserDataAddress)

		// Load 16 user data registers into s0-s15.
		stageOffset := gpu.GcnStageToUserDataOffset[ctx.Stage]
		for i := range uint32(16) {
			ptr := b.EmitPtrAccessChain(idPtrPsbUint, ptrBase, ctx.GetConstId(BlockContextId(stageOffset+i)))
			val := b.EmitLoad(ctx.GetId(BlockContextIdTypeUint), ptr, SpvMemoryAccessAligned, 4)
			ctx.SetGcnSgprId(b, i, val)
		}

		// Load vertex index and instance index into v0 and v1.
		if ctx.Stage == GcnShaderStageVertex {
			v0 := b.EmitLoad(ctx.GetId(BlockContextIdTypeUint), ctx.GetId(BlockContextIdVertexIndex))
			ctx.SetGcnVgprId(b, 0, v0)
			v1 := b.EmitLoad(ctx.GetId(BlockContextIdTypeUint), ctx.GetId(BlockContextIdInstanceIndex))
			ctx.SetGcnVgprId(b, 1, v1)
		}
	}

	// Reset condition ID.
	ctx.GcnConditionId = ctx.GetId(BlockContextIdFalse)

	// Emit instructions for current block.
	for i := range block.Instructions {
		emitInstruction(b, &block.Instructions[i], ctx)
	}

	// Terminate current block.
	switch block.Term {
	case TermCBranch:
		emitConditionalBranch(b, block, ctx)
	case TermBranch, TermFallthrough:
		if len(block.Successors) > 0 {
			b.EmitBranch(ctx.GetLabelId(block.Successors[0]))
		} else {
			b.EmitUnreachable()
		}
	case TermEndpgm, TermExpDone:
		b.EmitReturn()
	default:
		b.EmitReturn()
	}
}
