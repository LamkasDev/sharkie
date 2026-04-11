package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitSOP1(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	details := instr.Details.(*ScalarDetails)
	switch details.Op {
	case Sop1OpMovB32:
		val := ctx.GetOperandUintValue(b, details.Src0, instr.Literal)
		ctx.StoreRegisterPointer(b, details.Dst, val)
	case Sop1OpMovB64:
		valLo, valHi := ctx.GetOperand64Value(b, details.Src0, instr.Literal)
		ctx.StoreRegisterPointer(b, details.Dst, valLo)
		ctx.StoreRegisterPointer(b, details.Dst+1, valHi)
	case Sop1OpWqmB32:
		idUint := ctx.GetId(SpirvBlockContextIdUint)
		idBool := ctx.GetId(SpirvBlockContextIdBool)
		idC0 := ctx.GetId(SpirvBlockContextIdC0)
		idC1 := ctx.GetId(SpirvBlockContextIdC1)

		val := ctx.GetOperandUintValue(b, details.Src0, instr.Literal)
		res := emitWqmDword(b, ctx, val)
		ctx.StoreRegisterPointer(b, details.Dst, res)

		isNonZero := b.EmitINotEqual(idBool, res, idC0)
		sccVal := b.EmitSelect(idUint, isNonZero, idC1, idC0)
		ctx.StoreRegisterPointer(b, OpScc, sccVal)
	case Sop1OpWqmB64:
		idUint := ctx.GetId(SpirvBlockContextIdUint)
		idBool := ctx.GetId(SpirvBlockContextIdBool)
		idC0 := ctx.GetId(SpirvBlockContextIdC0)
		idC1 := ctx.GetId(SpirvBlockContextIdC1)

		valLo, valHi := ctx.GetOperand64Value(b, details.Src0, instr.Literal)
		resLo := emitWqmDword(b, ctx, valLo)
		resHi := emitWqmDword(b, ctx, valHi)
		ctx.StoreRegisterPointer(b, details.Dst, resLo)
		ctx.StoreRegisterPointer(b, details.Dst+1, resHi)

		isNonZeroLo := b.EmitINotEqual(idBool, resLo, idC0)
		isNonZeroHi := b.EmitINotEqual(idBool, resHi, idC0)
		isNonZero := b.EmitLogicalOr(idBool, isNonZeroLo, isNonZeroHi)
		sccVal := b.EmitSelect(idUint, isNonZero, idC1, idC0)
		ctx.StoreRegisterPointer(b, OpScc, sccVal)
	default:
		panic(fmt.Sprintf("unknown sop1 op %d", details.Op))
	}
}

func emitWqmDword(b *SpvBuilder, ctx SpirvBlockContext, val uint32) uint32 {
	idUint := ctx.GetId(SpirvBlockContextIdUint)
	idC1 := ctx.GetId(SpirvBlockContextIdC1)
	idC2 := ctx.GetId(SpirvBlockContextIdC2)
	idC3 := ctx.GetId(SpirvBlockContextIdC3)
	idMask := ctx.GetId(SpirvBlockContextIdC11111111)

	// Whole quad mode checks each group of four bits in the bitmask;
	// if any bit is set to 1, all four bits are set to 1 in the result.
	// This operation is repeated for the entire bitmask.
	s1 := b.EmitShiftRightLogical(idUint, val, idC1)
	s2 := b.EmitShiftRightLogical(idUint, val, idC2)
	s3 := b.EmitShiftRightLogical(idUint, val, idC3)
	t := b.EmitBitwiseOr(idUint, val, s1)
	t = b.EmitBitwiseOr(idUint, t, s2)
	t = b.EmitBitwiseOr(idUint, t, s3)
	s0 := b.EmitBitwiseAnd(idUint, t, idMask)
	l1 := b.EmitShiftLeftLogical(idUint, s0, idC1)
	l2 := b.EmitShiftLeftLogical(idUint, s0, idC2)
	l3 := b.EmitShiftLeftLogical(idUint, s0, idC3)
	res := b.EmitBitwiseOr(idUint, s0, l1)
	res = b.EmitBitwiseOr(idUint, res, l2)
	res = b.EmitBitwiseOr(idUint, res, l3)

	return res
}
