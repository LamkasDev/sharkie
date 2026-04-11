package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

type SpirvBlockContextId uint8

const (
	SpirvBlockContextIdFalse SpirvBlockContextId = iota
	SpirvBlockContextIdTrue
	SpirvBlockContextIdConst0
	SpirvBlockContextIdConst1
	SpirvBlockContextIdConst2
	SpirvBlockContextIdConst3
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
	SpirvBlockContextIdUint64
	SpirvBlockContextIdV2Float
	SpirvBlockContextIdGlsl
)

type SpirvBlockContext struct {
	LabelIds   []uint32
	Ids        map[SpirvBlockContextId]uint32
	SgprIds    [104]uint32
	VgprIds    [256]uint32
	SpecialIds [25]uint32
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
			if i == 19 {
				continue // reserved.
			}
			b.EmitLocalVariable(idPtrFnUint, ctx.GetSpecialId(uint32(i)))
		}
	}

	for i := range block.Instructions {
		emitInstruction(b, &block.Instructions[i], ctx)
	}

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
