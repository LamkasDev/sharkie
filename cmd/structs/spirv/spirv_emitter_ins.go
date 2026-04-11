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
	case op >= OpFlatScratchLo && op <= OpExecHi:
		return ctx.GetSpecialId(op - OpFlatScratchLo)
	case op >= OpVccz && op <= OpScc:
		return ctx.GetSpecialId((op - OpVccz) + SpecIdxVccz)
	case op >= OpVgpr0 && op <= OpVgpr255:
		return ctx.GetVgprId(op - OpVgpr0)
	}

	panic(fmt.Sprintf("unknown op %d", op))
}

// LoadRegisterPointer loads the value from the given register pointer.
func (ctx *SpirvBlockContext) LoadRegisterPointer(b *SpvBuilder, op uint32) uint32 {
	return b.EmitLoad(ctx.GetId(SpirvBlockContextIdUint), ctx.GetRegisterPointer(op))
}

// StoreRegisterPointer stores the given value into the given register pointer.
func (ctx *SpirvBlockContext) StoreRegisterPointer(b *SpvBuilder, op uint32, value uint32) {
	b.EmitStore(ctx.GetRegisterPointer(op), value)
}

// GetOperandValue returns the result ID of the value of the given operand.
func (ctx *SpirvBlockContext) GetOperandValue(b *SpvBuilder, op uint32, literal uint32) uint32 {
	idUint := ctx.GetId(SpirvBlockContextIdUint)
	switch {
	case op >= OpSgpr0 && op <= OpSgpr103:
		return b.EmitLoad(idUint, ctx.GetSgprId(op))
	case op >= OpFlatScratchLo && op <= OpExecHi:
		return b.EmitLoad(idUint, ctx.GetSpecialId(op-OpFlatScratchLo))
	case op >= OpInt0 && op <= OpFloatNeg40:
		return ctx.GetConstId(op - OpInt0)
	case op >= OpVccz && op <= OpScc:
		return b.EmitLoad(idUint, ctx.GetSpecialId((op-OpVccz)+SpecIdxVccz))
	case op == OpLiteral:
		return b.EmitConstantUint(idUint, literal)
	case op >= OpVgpr0 && op <= OpVgpr255:
		return b.EmitLoad(idUint, ctx.GetVgprId(op-OpVgpr0))
	}

	panic(fmt.Sprintf("unknown op %d", op))
}

// GetOperand64Value returns the result IDs of the low and high parts of the value of the given 64-bit operand.
func (ctx *SpirvBlockContext) GetOperand64Value(b *SpvBuilder, op uint32, literal uint32) (uint32, uint32) {
	idUint := ctx.GetId(SpirvBlockContextIdUint)
	switch {
	case op >= OpSgpr0 && op <= OpSgpr103:
		return b.EmitLoad(idUint, ctx.GetSgprId(op)), b.EmitLoad(idUint, ctx.GetSgprId(op+1))
	case op >= OpFlatScratchLo && op <= OpExecHi:
		return b.EmitLoad(idUint, ctx.GetSpecialId(op-OpFlatScratchLo)), b.EmitLoad(idUint, ctx.GetSpecialId(op+1-OpFlatScratchLo))
	case op >= OpVgpr0 && op <= OpVgpr255:
		return b.EmitLoad(idUint, ctx.GetVgprId(op-OpVgpr0)), b.EmitLoad(idUint, ctx.GetVgprId(op-OpVgpr0+1))
	case op >= OpInt0 && op <= OpPosInt64:
		return ctx.GetConstId(op - OpInt0), ctx.GetId(SpirvBlockContextIdC0)
	case op >= OpNegInt1 && op <= OpNegInt16:
		return ctx.GetConstId(op - OpInt0), ctx.GetId(SpirvBlockContextIdCFFFFFFFF)
	case op >= OpFloat05 && op <= OpFloatNeg40:
		return ctx.GetConstId(op - OpInt0), ctx.GetId(SpirvBlockContextIdC0)
	case op >= OpVccz && op <= OpScc:
		return b.EmitLoad(idUint, ctx.GetSpecialId((op-OpVccz)+SpecIdxVccz)), ctx.GetId(SpirvBlockContextIdC0)
	case op == OpLiteral:
		return b.EmitConstantUint(idUint, literal), ctx.GetId(SpirvBlockContextIdC0)
	}

	panic(fmt.Sprintf("unknown 64-bit op %d", op))
}

// GetOperandUintValue returns the result ID of the value of the given operand as a uint.
func (ctx *SpirvBlockContext) GetOperandUintValue(b *SpvBuilder, op uint32, literal uint32) uint32 {
	return ctx.GetOperandValue(b, op, literal)
}

// GetOperandIntValue returns the result ID of the value of the given operand as an int.
func (ctx *SpirvBlockContext) GetOperandIntValue(b *SpvBuilder, op uint32, literal uint32) uint32 {
	return b.EmitBitcast(ctx.GetId(SpirvBlockContextIdInt), ctx.GetOperandValue(b, op, literal))
}

// GetOperandFloatValue returns the result ID of the value of the given operand as a float.
func (ctx *SpirvBlockContext) GetOperandFloatValue(b *SpvBuilder, op uint32, literal uint32) uint32 {
	return b.EmitBitcast(ctx.GetId(SpirvBlockContextIdFloat), ctx.GetOperandValue(b, op, literal))
}

// emitInstruction emits the SPIR-V for a single instruction.
func emitInstruction(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	emitFunc, ok := InstructionEmitMap[instr.Encoding]
	if !ok {
		panic(fmt.Errorf("unknown encoding %s", instr.Encoding))
	}
	emitFunc(b, instr, ctx)
}
