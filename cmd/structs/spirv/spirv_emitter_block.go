package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
	"github.com/LamkasDev/sharkie/cmd/structs/gpu"
)

type SpirvBlockContextId uint8

const (
	SpirvBlockContextIdFalse SpirvBlockContextId = iota
	SpirvBlockContextIdTrue
	SpirvBlockContextIdBool
	SpirvBlockContextIdColorOut
	SpirvBlockContextIdZeroVec4
	SpirvBlockContextIdPcVar
	SpirvBlockContextIdPtrPcFloat
	SpirvBlockContextIdPtrPcPsbUint
	SpirvBlockContextIdPtrPsbUint
	SpirvBlockContextIdFloat
	SpirvBlockContextIdV4Float
	SpirvBlockContextIdPtrFnUint
	SpirvBlockContextIdUint
	SpirvBlockContextIdInt
	SpirvBlockContextIdUint64
	SpirvBlockContextIdV2Float
	SpirvBlockContextIdGlsl
	SpirvBlockContextIdC0
	SpirvBlockContextIdC1
	SpirvBlockContextIdC2
	SpirvBlockContextIdC3
	SpirvBlockContextIdC4
	SpirvBlockContextIdC5
	SpirvBlockContextIdC6
	SpirvBlockContextIdC7
	SpirvBlockContextIdC11111111
	SpirvBlockContextIdCFFFFFFFF
)

const (
	SpecIdxFlatScrLo = 0
	SpecIdxFlatScrHi = 1
	SpecIdxVccLo     = 2
	SpecIdxVccHi     = 3
	SpecIdxTbaLo     = 4
	SpecIdxTbaHi     = 5
	SpecIdxTmaLo     = 6
	SpecIdxTmaHi     = 7
	SpecIdxTtmp0     = 8
	SpecIdxTtmp11    = 19
	SpecIdxM0        = 20
	SpecIdxReserved  = 21
	SpecIdxExecLo    = 22
	SpecIdxExecHi    = 23
	SpecIdxVccz      = 24
	SpecIdxExecz     = 25
	SpecIdxScc       = 26
)

const (
	ConstIdx0          = 0
	ConstIdxInt1       = 1
	ConstIdxInt64      = 64
	ConstIdxIntNeg1    = 65
	ConstIdxIntNeg16   = 80
	ConstIdxFloat05    = 112
	ConstIdxFloatNeg05 = 113
	ConstIdxFloat10    = 114
	ConstIdxFloatNeg10 = 115
	ConstIdxFloat20    = 116
	ConstIdxFloatNeg20 = 117
	ConstIdxFloat40    = 118
	ConstIdxFloatNeg40 = 119
)

type SpirvBlockContext struct {
	Stage      GcnShaderStage
	LabelIds   []uint32
	Ids        map[SpirvBlockContextId]uint32
	SgprIds    [104]uint32
	VgprIds    [256]uint32
	SpecialIds [27]uint32
	ConstIds   [120]uint32
}

func (ctx *SpirvBlockContext) GetLabelId(i int) uint32 {
	id := ctx.LabelIds[i]
	if id == 0 {
		panic(fmt.Sprintf("label id %d is zero", i))
	}

	return id
}

func (ctx *SpirvBlockContext) GetId(i SpirvBlockContextId) uint32 {
	id := ctx.Ids[i]
	if id == 0 {
		panic(fmt.Sprintf("id %d is zero", i))
	}

	return id
}

func (ctx *SpirvBlockContext) TryGetId(i SpirvBlockContextId) (uint32, bool) {
	id, ok := ctx.Ids[i]
	if ok && id == 0 {
		panic(fmt.Sprintf("id %d is zero", i))
	}

	return id, ok
}

func (ctx *SpirvBlockContext) GetSgprId(reg uint32) uint32 {
	id := ctx.SgprIds[reg]
	if id == 0 {
		panic(fmt.Sprintf("sgpr id %d is zero", reg))
	}

	return id
}

func (ctx *SpirvBlockContext) GetVgprId(reg uint32) uint32 {
	id := ctx.VgprIds[reg]
	if id == 0 {
		panic(fmt.Sprintf("vgpr id %d is zero", reg))
	}

	return id
}

func (ctx *SpirvBlockContext) GetSpecialId(reg uint32) uint32 {
	id := ctx.SpecialIds[reg]
	if id == 0 {
		panic(fmt.Sprintf("special id %d is zero", reg))
	}

	return id
}

func (ctx *SpirvBlockContext) GetConstId(reg uint32) uint32 {
	id := ctx.ConstIds[reg]
	if id == 0 {
		panic(fmt.Sprintf("const id %d is zero", reg))
	}

	return id
}

// emitBlock emits the SPIR-V for a single block.
func emitBlock(b *SpvBuilder, block *GcnShaderCfgBlock, ctx SpirvBlockContext) {
	// Start current block.
	b.EmitLabel(ctx.GetLabelId(block.Id))

	// Declare variables in entry block.
	if block.DwordOffset == 0 {
		idPtrFnUint := ctx.GetId(SpirvBlockContextIdPtrFnUint)
		for i := range ctx.SgprIds {
			b.EmitLocalVariable(idPtrFnUint, ctx.GetSgprId(uint32(i)))
		}
		for i := range ctx.VgprIds {
			b.EmitLocalVariable(idPtrFnUint, ctx.GetVgprId(uint32(i)))
		}
		for i := range ctx.SpecialIds {
			if i == SpecIdxReserved {
				continue // reserved.
			}
			b.EmitLocalVariable(idPtrFnUint, ctx.GetSpecialId(uint32(i)))
		}

		// Load user data buffer address from the push constant.
		idPtrPsbUint := ctx.GetId(SpirvBlockContextIdPtrPsbUint)
		ptrPcPsbUint := b.EmitAccessChain(ctx.GetId(SpirvBlockContextIdPtrPcPsbUint), ctx.GetId(SpirvBlockContextIdPcVar), b.EmitConstantUint(ctx.GetId(SpirvBlockContextIdUint), 3))
		ptrBase := b.EmitLoad(idPtrPsbUint, ptrPcPsbUint)

		// Load 16 user data registers into s0-s15.
		stageOffset := gpu.GcnStageToUserDataOffset[ctx.Stage]
		for i := range uint32(16) {
			idx := b.EmitConstantUint(ctx.GetId(SpirvBlockContextIdUint), stageOffset+i)
			ptr := b.EmitPtrAccessChain(idPtrPsbUint, ptrBase, idx)
			val := b.EmitLoad(ctx.GetId(SpirvBlockContextIdUint), ptr, SpvMemoryAccessAligned, 4)
			b.EmitStore(ctx.GetSgprId(i), val)
		}
	}

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
