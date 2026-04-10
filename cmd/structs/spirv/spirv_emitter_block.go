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
	SpirvBlockContextIdFloat
	SpirvBlockContextIdV4Float
	SpirvBlockContextIdPtrFnUint
	SpirvBlockContextIdUint
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

// GetRegisterPointer returns the result ID of the pointer to the given register.
func (ctx *SpirvBlockContext) GetRegisterPointer(op uint32) uint32 {
	switch {
	case op <= 103:
		return ctx.GetSgprId(op)
	case op >= 106 && op <= 127:
		return ctx.GetSpecialId(op - 106)
	case op >= 251 && op <= 253:
		return ctx.GetSpecialId((op - 251) + 22)
	case op >= 256 && op <= 511:
		return ctx.GetVgprId(op - 256)
	}

	panic(fmt.Sprintf("unknown op %d", op))
}

// GetOperandValue returns the result ID of the value of the given operand.
func (ctx *SpirvBlockContext) GetOperandValue(b *SpvBuilder, op uint32, literal uint32) uint32 {
	switch {
	case op <= 103:
		return b.EmitLoad(ctx.GetId(SpirvBlockContextIdUint), ctx.GetSgprId(op))
	case op >= 106 && op <= 127:
		return b.EmitLoad(ctx.GetId(SpirvBlockContextIdUint), ctx.GetSpecialId(op-106))
	case op >= 128 && op <= 247:
		return ctx.GetConstId(op - 128)
	case op >= 251 && op <= 253:
		return b.EmitLoad(ctx.GetId(SpirvBlockContextIdUint), ctx.GetSpecialId((op-251)+22))
	case op == 255:
		return b.EmitConstantUint(ctx.GetId(SpirvBlockContextIdUint), literal)
	case op >= 256 && op <= 511:
		return b.EmitLoad(ctx.GetId(SpirvBlockContextIdUint), ctx.GetVgprId(op-256))
	}

	panic(fmt.Sprintf("unknown op %d", op))
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

// emitInstruction emits the SPIR-V for a single instruction.
func emitInstruction(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	switch instr.Encoding {
	case EncSOP1:
		switch instr.SOp {
		case 0x03: // s_mov_b32
			val := ctx.GetOperandValue(b, instr.SSrc0, instr.Literal)
			ptr := ctx.GetRegisterPointer(instr.SDst)
			b.EmitStore(ptr, val)
		}
	case EncSMRD:
		switch instr.SmOp {
		case 0x0A: // s_buffer_load_dwordx4
			// Mock load from push constant for now to demonstrate.
			if idPcVar, ok := ctx.TryGetId(SpirvBlockContextIdPcVar); ok {
				for i := range uint32(4) {
					ptrPc := b.EmitAccessChain(ctx.GetId(SpirvBlockContextIdPtrPcFloat), idPcVar, ctx.GetId(SpirvBlockContextIdConst0), b.EmitConstantUint(ctx.GetId(SpirvBlockContextIdUint), i))
					valF := b.EmitLoad(ctx.GetId(SpirvBlockContextIdFloat), ptrPc)
					valU := b.EmitBitcast(ctx.GetId(SpirvBlockContextIdUint), valF)
					ptrReg := ctx.GetRegisterPointer(instr.SmDst + i)
					b.EmitStore(ptrReg, valU)
				}
			}
		}
	case EncSOPP:
		switch instr.SOp {
		case SoppOpWaitCnt:
			// No-op in SPIR-V for now.
		}
	case EncVOP1:
		switch instr.VOp {
		case 0x01: // v_mov_b32
			val := ctx.GetOperandValue(b, instr.VSrc0, instr.Literal)
			ptr := ctx.GetRegisterPointer(instr.VDst + 256)
			b.EmitStore(ptr, val)
		}
	case EncVOP2:
		switch instr.VOp {
		case 0x2F: // v_cvt_pkrtz_f16_f32
			val0 := ctx.GetOperandValue(b, instr.VSrc0, instr.Literal)
			val1 := ctx.GetOperandValue(b, instr.VSrc1+256, 0)
			fval0 := b.EmitBitcast(ctx.GetId(SpirvBlockContextIdFloat), val0)
			fval1 := b.EmitBitcast(ctx.GetId(SpirvBlockContextIdFloat), val1)
			vec := b.EmitCompositeConstruct(ctx.GetId(SpirvBlockContextIdV2Float), fval0, fval1)
			resU := b.EmitExtInst(ctx.GetId(SpirvBlockContextIdUint), ctx.GetId(SpirvBlockContextIdGlsl), 58, vec) // PackHalf2x16
			ptr := ctx.GetRegisterPointer(instr.VDst + 256)
			b.EmitStore(ptr, resU)
		}
	case EncEXP:
		if instr.ExpTarget == 0 { // target=0 (color)
			val0 := ctx.GetOperandValue(b, instr.ExpVSrcs[0]+256, 0)
			val1 := ctx.GetOperandValue(b, instr.ExpVSrcs[1]+256, 0)
			v01 := b.EmitExtInst(ctx.GetId(SpirvBlockContextIdV2Float), ctx.GetId(SpirvBlockContextIdGlsl), 62, val0) // UnpackHalf2x16
			v23 := b.EmitExtInst(ctx.GetId(SpirvBlockContextIdV2Float), ctx.GetId(SpirvBlockContextIdGlsl), 62, val1) // UnpackHalf2x16
			f0 := b.EmitCompositeExtract(ctx.GetId(SpirvBlockContextIdFloat), v01, 0)
			f1 := b.EmitCompositeExtract(ctx.GetId(SpirvBlockContextIdFloat), v01, 1)
			f2 := b.EmitCompositeExtract(ctx.GetId(SpirvBlockContextIdFloat), v23, 0)
			f3 := b.EmitCompositeExtract(ctx.GetId(SpirvBlockContextIdFloat), v23, 1)
			vec := b.EmitCompositeConstruct(ctx.GetId(SpirvBlockContextIdV4Float), f0, f1, f2, f3)
			if idColorOut, ok := ctx.TryGetId(SpirvBlockContextIdColorOut); ok {
				b.EmitStore(idColorOut, vec)
			}
		}
	}
}
