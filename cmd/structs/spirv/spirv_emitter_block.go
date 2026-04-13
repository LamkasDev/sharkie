package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
	"github.com/LamkasDev/sharkie/cmd/structs/gpu"
)

type BlockContextId uint8

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
	BlockContextIdTypeV4Uint
	BlockContextIdTypeSampledImage
	BlockContextIdPtrUniformSampledImage
	BlockContextIdPtrPcFloat
	BlockContextIdPtrPcPsbUint
	BlockContextIdPtrPcPsbUint64
	BlockContextIdPtrPsbUint
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
	ConstIdxUint0 BlockContextId = iota
	ConstIdxUint1
	ConstIdxUint2
	ConstIdxUint3
	ConstIdxUint4
	ConstIdxUint5
	ConstIdxUint6
	ConstIdxUint7
	ConstIdxUint32
	ConstIdxUint63
	ConstIdxUintFFFF
	ConstIdxUint11111111
	ConstIdxUintFFFFFFFF
	ConstIdxFloat1
	ConstIdxFloat0
)

const (
	PushConstantTime            = 0
	PushConstant_               = 1
	PushConstantConstRamAddress = 2
	PushConstantUserDataAddress = 3
	PushConstantGarlicAddress   = 4
	PushConstantOnionAddress    = 5
)

type SpirvBlockContext struct {
	Stage    GcnShaderStage
	LabelIds []uint32
	Ids      map[BlockContextId]uint32
	ConstIds map[BlockContextId]uint32

	GcnSgprIds     [104]uint32
	GcnVgprIds     [256]uint32
	GcnSpecialIds  [27]uint32
	GcnConstIds    [120]uint32
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
	if id == 0 {
		panic(fmt.Sprintf("id %d is zero", i))
	}

	return id
}

func (ctx *SpirvBlockContext) GetConstId(i BlockContextId) uint32 {
	id := ctx.ConstIds[i]
	if id == 0 {
		panic(fmt.Sprintf("const id %d is zero", i))
	}

	return id
}

func (ctx *SpirvBlockContext) GetGcnSgprId(reg uint32) uint32 {
	id := ctx.GcnSgprIds[reg]
	if id == 0 {
		panic(fmt.Sprintf("gcn sgpr id %d is zero", reg))
	}

	return id
}

func (ctx *SpirvBlockContext) GetGcnVgprId(reg uint32) uint32 {
	id := ctx.GcnVgprIds[reg]
	if id == 0 {
		panic(fmt.Sprintf("gcn vgpr id %d is zero", reg))
	}

	return id
}

func (ctx *SpirvBlockContext) GetGcnSpecialId(reg uint32) uint32 {
	id := ctx.GcnSpecialIds[reg]
	if id == 0 {
		panic(fmt.Sprintf("gcn special id %d is zero", reg))
	}

	return id
}

func (ctx *SpirvBlockContext) GetGcnConstId(reg uint32) uint32 {
	id := ctx.GcnConstIds[reg]
	if id == 0 {
		panic(fmt.Sprintf("gcn const id %d is zero", reg))
	}

	return id
}

// emitBlock emits the SPIR-V for a single block.
func emitBlock(b *SpvBuilder, block *GcnShaderCfgBlock, ctx SpirvBlockContext) {
	// Start current block.
	b.EmitLabel(ctx.GetLabelId(block.Id))

	// Declare variables in entry block.
	if block.DwordOffset == 0 {
		idPtrFnUint := ctx.GetId(BlockContextIdPtrFnUint)
		for i := range ctx.GcnSgprIds {
			b.EmitLocalVariable(idPtrFnUint, ctx.GetGcnSgprId(uint32(i)))
		}
		for i := range ctx.GcnVgprIds {
			b.EmitLocalVariable(idPtrFnUint, ctx.GetGcnVgprId(uint32(i)))
		}
		for i := range ctx.GcnSpecialIds {
			if i == GcnSpecIdxReserved {
				continue // reserved.
			}
			b.EmitLocalVariable(idPtrFnUint, ctx.GetGcnSpecialId(uint32(i)))
		}

		// Load user data buffer address from the push constant.
		idPtrPsbUint := ctx.GetId(BlockContextIdPtrPsbUint)
		ptrBase := ctx.LoadPushConstantValue(b, PushConstantUserDataAddress)

		// Load 16 user data registers into s0-s15.
		stageOffset := gpu.GcnStageToUserDataOffset[ctx.Stage]
		for i := range uint32(16) {
			idx := b.EmitConstantUint(ctx.GetId(BlockContextIdTypeUint), stageOffset+i)
			ptr := b.EmitPtrAccessChain(idPtrPsbUint, ptrBase, idx)
			val := b.EmitLoad(ctx.GetId(BlockContextIdTypeUint), ptr, SpvMemoryAccessAligned, 4)
			b.EmitStore(ctx.GetGcnSgprId(i), val)
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
