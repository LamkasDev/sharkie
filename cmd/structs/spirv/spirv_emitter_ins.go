package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

type InstructionEmitFunc func(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext)

var InstructionEmitMap = map[Encoding]InstructionEmitFunc{
	EncSOP2:  emitSOP2,
	EncSOP1:  emitSOP1,
	EncSOPC:  emitSOPC,
	EncSOPP:  emitSOPP,
	EncVOP2:  emitVOP2,
	EncVOP1:  emitVOP1,
	EncVOPC:  emitVOPC,
	EncVOP3:  emitVOP3,
	EncSMRD:  emitSMRD,
	EncMUBUF: emitMUBUF,
	EncMIMG:  emitMIMG,
	EncEXP:   emitEXP,
}

// GetRegisterPointer returns the result ID of the pointer to the given register.
func (ctx *SpirvBlockContext) GetRegisterPointer(op uint32) uint32 {
	switch {
	case op >= OpSgpr0 && op <= OpSgpr103:
		return ctx.GetSgprId(op)
	case op >= OpVccLo && op <= OpExecHi:
		return ctx.GetSpecialId(op - OpVccLo)
	case op >= OpVccz && op <= OpScc:
		return ctx.GetSpecialId((op - OpVccz) + SpecIdxVccz)
	case op >= OpVgpr0 && op <= OpVgpr255:
		return ctx.GetVgprId(op - OpVgpr0)
	}

	panic(fmt.Sprintf("unknown op %d", op))
}

// GetOperandValue returns the result ID of the value of the given operand.
func (ctx *SpirvBlockContext) GetOperandValue(b *SpvBuilder, op uint32, literal uint32) uint32 {
	switch {
	case op >= OpSgpr0 && op <= OpSgpr103:
		return b.EmitLoad(ctx.GetId(SpirvBlockContextIdUint), ctx.GetSgprId(op))
	case op >= OpVccLo && op <= OpExecHi:
		return b.EmitLoad(ctx.GetId(SpirvBlockContextIdUint), ctx.GetSpecialId(op-OpVccLo))
	case op >= OpInt0 && op <= OpFloatNeg40:
		return ctx.GetConstId(op - OpInt0)
	case op >= OpVccz && op <= OpScc:
		return b.EmitLoad(ctx.GetId(SpirvBlockContextIdUint), ctx.GetSpecialId((op-OpVccz)+SpecIdxVccz))
	case op == OpLiteral:
		return b.EmitConstantUint(ctx.GetId(SpirvBlockContextIdUint), literal)
	case op >= OpVgpr0 && op <= OpVgpr255:
		return b.EmitLoad(ctx.GetId(SpirvBlockContextIdUint), ctx.GetVgprId(op-OpVgpr0))
	}

	panic(fmt.Sprintf("unknown op %d", op))
}

// emitInstruction emits the SPIR-V for a single instruction.
func emitInstruction(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	emitFunc, ok := InstructionEmitMap[instr.Encoding]
	if !ok {
		panic(fmt.Errorf("unknown encoding %s", instr.Encoding))
	}
	emitFunc(b, instr, ctx)
}
