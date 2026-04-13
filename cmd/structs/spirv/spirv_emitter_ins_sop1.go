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
		idUint := ctx.GetId(BlockContextIdTypeUint)
		idBool := ctx.GetId(BlockContextIdTypeBool)
		idC0 := ctx.GetConstId(ConstIdxUint0)
		idC1 := ctx.GetConstId(ConstIdxUint1)

		val := ctx.GetOperandUintValue(b, details.Src0, instr.Literal)
		res := emitWqmDword(b, ctx, val)
		ctx.StoreRegisterPointer(b, details.Dst, res)

		isNonZero := b.EmitINotEqual(idBool, res, idC0)
		sccVal := b.EmitSelect(idUint, isNonZero, idC1, idC0)
		ctx.StoreRegisterPointer(b, OpScc, sccVal)
	case Sop1OpWqmB64:
		idUint := ctx.GetId(BlockContextIdTypeUint)
		idBool := ctx.GetId(BlockContextIdTypeBool)
		idC0 := ctx.GetConstId(ConstIdxUint0)
		idC1 := ctx.GetConstId(ConstIdxUint1)

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
	case Sop1OpFlbitI32I64:
		idInt := ctx.GetId(BlockContextIdTypeInt)
		idUint := ctx.GetId(BlockContextIdTypeUint)
		idBool := ctx.GetId(BlockContextIdTypeBool)
		idGlsl := ctx.GetId(BlockContextIdGlsl)
		idC30 := b.EmitConstantUint(idUint, 30)
		idC31 := b.EmitConstantUint(idUint, 31)
		idC62 := b.EmitConstantUint(idUint, 62)
		idNeg1 := b.EmitConstantUint(idUint, 0xFFFFFFFF)

		valLo, valHi := ctx.GetOperand64Value(b, details.Src0, instr.Literal)

		// 1. Check high 32 bits. bit 31 of valHi is the sign bit.
		// FindSMsb returns index of first bit != sign bit (0-30), or -1.
		msbHi := b.EmitExtInst(idInt, idGlsl, SpvGlslOpFindSMsb, b.EmitBitcast(idInt, valHi))
		isHiNotAllSame := b.EmitINotEqual(idBool, b.EmitBitcast(idUint, msbHi), idNeg1)
		resHi := b.EmitISub(idUint, idC30, b.EmitBitcast(idUint, msbHi))

		// 2. Check low 32 bits if high bits were all sign bits.
		// signMask = (int32(valHi) >> 31) -> all 0s or all 1s.
		signMask := b.EmitShiftRightArithmetic(idInt, b.EmitBitcast(idInt, valHi), idC31)
		// xLo = valLo ^ signMask -> bits are 1 where they differ from sign.
		xLo := b.EmitBitwiseXor(idUint, valLo, b.EmitBitcast(idUint, signMask))
		// FindUMsb returns index of first '1' (0-31), or -1.
		msbLo := b.EmitExtInst(idUint, idGlsl, SpvGlslOpFindUMsb, xLo)
		isLoNotAllSame := b.EmitINotEqual(idBool, msbLo, idNeg1)
		resLo := b.EmitISub(idUint, idC62, msbLo)

		// 3. Final result: Hi distance if found, else Lo distance if found, else -1.
		res := b.EmitSelect(idUint, isHiNotAllSame, resHi, b.EmitSelect(idUint, isLoNotAllSame, resLo, idNeg1))
		ctx.StoreRegisterPointer(b, details.Dst, res)
	default:
		panic(fmt.Sprintf("unknown sop1 op %s", Mnemotics[EncSOP1][details.Op]))
	}
}

func emitWqmDword(b *SpvBuilder, ctx SpirvBlockContext, val uint32) uint32 {
	idUint := ctx.GetId(BlockContextIdTypeUint)
	idC1 := ctx.GetConstId(ConstIdxUint1)
	idC2 := ctx.GetConstId(ConstIdxUint2)
	idC3 := ctx.GetConstId(ConstIdxUint3)
	idMask := ctx.GetConstId(ConstIdxUint11111111)

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
