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
		return ctx.GetGcnSgprId(op)
	case op >= OpFlatScratchLo && op <= OpExecHi:
		return ctx.GetGcnSpecialId(op - OpFlatScratchLo)
	case op >= OpVccz && op <= OpScc:
		return ctx.GetGcnSpecialId((op - OpVccz) + GcnSpecIdxVccz)
	case op >= OpVgpr0 && op <= OpVgpr255:
		return ctx.GetGcnVgprId(op - OpVgpr0)
	}

	panic(fmt.Sprintf("unknown op %d", op))
}

// LoadRegisterPointer loads the value from the given register pointer.
func (ctx *SpirvBlockContext) LoadRegisterPointer(b *SpvBuilder, op uint32) uint32 {
	return b.EmitLoad(ctx.GetId(BlockContextIdTypeUint), ctx.GetRegisterPointer(op))
}

// StoreRegisterPointer stores the given value into the given register pointer.
func (ctx *SpirvBlockContext) StoreRegisterPointer(b *SpvBuilder, op uint32, value uint32) {
	b.EmitStore(ctx.GetRegisterPointer(op), value)
}

// GetOperandValue returns the result ID of the value of the given operand.
func (ctx *SpirvBlockContext) GetOperandValue(b *SpvBuilder, op uint32, literal uint32) uint32 {
	idUint := ctx.GetId(BlockContextIdTypeUint)
	switch {
	case op >= OpSgpr0 && op <= OpSgpr103:
		return b.EmitLoad(idUint, ctx.GetGcnSgprId(op))
	case op >= OpFlatScratchLo && op <= OpExecHi:
		return b.EmitLoad(idUint, ctx.GetGcnSpecialId(op-OpFlatScratchLo))
	case op >= OpInt0 && op <= OpFloatNeg40:
		return ctx.GetGcnConstId(op - OpInt0)
	case op >= OpVccz && op <= OpScc:
		return b.EmitLoad(idUint, ctx.GetGcnSpecialId((op-OpVccz)+GcnSpecIdxVccz))
	case op == OpLiteral:
		return b.EmitConstantUint(idUint, literal)
	case op >= OpVgpr0 && op <= OpVgpr255:
		return b.EmitLoad(idUint, ctx.GetGcnVgprId(op-OpVgpr0))
	}

	panic(fmt.Sprintf("unknown op %d", op))
}

// GetOperand64Value returns the result IDs of the low and high parts of the value of the given 64-bit operand.
func (ctx *SpirvBlockContext) GetOperand64Value(b *SpvBuilder, op uint32, literal uint32) (uint32, uint32) {
	idUint := ctx.GetId(BlockContextIdTypeUint)
	switch {
	case op >= OpSgpr0 && op <= OpSgpr103:
		return b.EmitLoad(idUint, ctx.GetGcnSgprId(op)), b.EmitLoad(idUint, ctx.GetGcnSgprId(op+1))
	case op >= OpFlatScratchLo && op <= OpExecHi:
		return b.EmitLoad(idUint, ctx.GetGcnSpecialId(op-OpFlatScratchLo)), b.EmitLoad(idUint, ctx.GetGcnSpecialId(op+1-OpFlatScratchLo))
	case op >= OpVgpr0 && op <= OpVgpr255:
		return b.EmitLoad(idUint, ctx.GetGcnVgprId(op-OpVgpr0)), b.EmitLoad(idUint, ctx.GetGcnVgprId(op-OpVgpr0+1))
	case op >= OpInt0 && op <= OpPosInt64:
		return ctx.GetGcnConstId(op - OpInt0), ctx.GetConstId(ConstIdxUint0)
	case op >= OpNegInt1 && op <= OpNegInt16:
		return ctx.GetGcnConstId(op - OpInt0), ctx.GetConstId(ConstIdxUintFFFFFFFF)
	case op >= OpFloat05 && op <= OpFloatNeg40:
		return ctx.GetGcnConstId(op - OpInt0), ctx.GetConstId(ConstIdxUint0)
	case op >= OpVccz && op <= OpScc:
		return b.EmitLoad(idUint, ctx.GetGcnSpecialId((op-OpVccz)+GcnSpecIdxVccz)), ctx.GetConstId(ConstIdxUint0)
	case op == OpLiteral:
		return b.EmitConstantUint(idUint, literal), ctx.GetConstId(ConstIdxUint0)
	}

	panic(fmt.Sprintf("unknown 64-bit op %d", op))
}

// GetOperandUintValue returns the result ID of the value of the given operand as a uint.
func (ctx *SpirvBlockContext) GetOperandUintValue(b *SpvBuilder, op uint32, literal uint32) uint32 {
	return ctx.GetOperandValue(b, op, literal)
}

// GetOperandIntValue returns the result ID of the value of the given operand as an int.
func (ctx *SpirvBlockContext) GetOperandIntValue(b *SpvBuilder, op uint32, literal uint32) uint32 {
	return b.EmitBitcast(ctx.GetId(BlockContextIdTypeInt), ctx.GetOperandValue(b, op, literal))
}

// GetOperandFloatValue returns the result ID of the value of the given operand as a float.
func (ctx *SpirvBlockContext) GetOperandFloatValue(b *SpvBuilder, op uint32, literal uint32) uint32 {
	return b.EmitBitcast(ctx.GetId(BlockContextIdTypeFloat), ctx.GetOperandValue(b, op, literal))
}

// TestMask returns a boolean result ID of (val & mask) != 0.
func (ctx *SpirvBlockContext) TestMask(b *SpvBuilder, val uint32, mask uint32) uint32 {
	maskId := b.EmitConstantUint(ctx.GetId(BlockContextIdTypeUint), mask)
	andId := b.EmitBitwiseAnd(ctx.GetId(BlockContextIdTypeUint), val, maskId)
	return b.EmitINotEqual(ctx.GetId(BlockContextIdTypeBool), andId, ctx.GetConstId(ConstIdxUint0))
}

// Pack64 combines two 32-bit values into one 64-bit value.
func (ctx *SpirvBlockContext) Pack64(b *SpvBuilder, lo, hi uint32) uint32 {
	idUint64 := ctx.GetId(BlockContextIdTypeUint64)
	lo64 := b.EmitUConvert(idUint64, lo)
	hi64 := b.EmitUConvert(idUint64, hi)
	shift64 := b.EmitConstantUint64(idUint64, 32)
	hiShifted := b.EmitShiftLeftLogical(idUint64, hi64, shift64)
	return b.EmitBitwiseOr(idUint64, lo64, hiShifted)
}

// LoadPushConstantPtr loads a pointer in push constant at offset and returns the ID.
func (ctx *SpirvBlockContext) LoadPushConstantPtr(b *SpvBuilder, i uint32) uint32 {
	idPtrPsbUint := ctx.GetId(BlockContextIdPtrPsbUint)
	ptrPcPsbUint := b.EmitAccessChain(ctx.GetId(BlockContextIdPtrPcPsbUint), ctx.GetId(BlockContextIdPcVar), b.EmitConstantUint(ctx.GetId(BlockContextIdTypeUint), i))
	return b.EmitLoad(idPtrPsbUint, ptrPcPsbUint)
}

// GetResourceBaseAddress extracts the base address from T# dword 0 and 1.
func (ctx *SpirvBlockContext) GetResourceBaseAddress(b *SpvBuilder, dw0, dw1 uint32) uint32 {
	idUint := ctx.GetId(BlockContextIdTypeUint)
	idUint64 := ctx.GetId(BlockContextIdTypeUint64)

	baseLo := dw0
	baseHi := b.EmitBitwiseAnd(idUint, dw1, ctx.GetConstId(ConstIdxUintFFFF))
	base := ctx.Pack64(b, baseLo, baseHi)

	// Add GPU memory offset from push constant.
	gpuBase := ctx.LoadPushConstantPtr(b, PushConstantGarlicAddress)

	return b.EmitIAdd(idUint64, base, gpuBase)
}

// GetResourceStride extracts stride from T# dword 1.
func (ctx *SpirvBlockContext) GetResourceStride(b *SpvBuilder, dw1 uint32) uint32 {
	idUint := ctx.GetId(BlockContextIdTypeUint)
	return b.EmitShiftRightLogical(idUint, dw1, b.EmitConstantUint(idUint, 16))
}

// emitInstruction emits the SPIR-V for a single instruction.
func emitInstruction(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	emitFunc, ok := InstructionEmitMap[instr.Encoding]
	if !ok {
		panic(fmt.Errorf("unknown encoding %s", instr.Encoding))
	}
	emitFunc(b, instr, ctx)
}
