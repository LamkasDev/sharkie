package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitSOP1(b *SpvBuilder, instr *Instruction, ctx *SpirvBlockContext) {
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
		typeUint := ctx.GetId(BlockContextIdTypeUint)
		typeBool := ctx.GetId(BlockContextIdTypeBool)
		idC0 := ctx.GetConstId(ConstIdxUint0)
		idC1 := ctx.GetConstId(ConstIdxUint1)

		val := ctx.GetOperandUintValue(b, details.Src0, instr.Literal)
		res := emitWqmDword(b, ctx, val)
		ctx.StoreRegisterPointer(b, details.Dst, res)

		isNonZero := b.EmitINotEqual(typeBool, res, idC0)
		sccVal := b.EmitSelect(typeUint, isNonZero, idC1, idC0)
		ctx.StoreRegisterPointer(b, OpScc, sccVal)
	case Sop1OpWqmB64:
		typeUint := ctx.GetId(BlockContextIdTypeUint)
		typeBool := ctx.GetId(BlockContextIdTypeBool)
		idC0 := ctx.GetConstId(ConstIdxUint0)
		idC1 := ctx.GetConstId(ConstIdxUint1)

		valLo, valHi := ctx.GetOperand64Value(b, details.Src0, instr.Literal)
		resLo := emitWqmDword(b, ctx, valLo)
		resHi := emitWqmDword(b, ctx, valHi)
		ctx.StoreRegisterPointer(b, details.Dst, resLo)
		ctx.StoreRegisterPointer(b, details.Dst+1, resHi)

		isNonZeroLo := b.EmitINotEqual(typeBool, resLo, idC0)
		isNonZeroHi := b.EmitINotEqual(typeBool, resHi, idC0)
		isNonZero := b.EmitLogicalOr(typeBool, isNonZeroLo, isNonZeroHi)
		sccVal := b.EmitSelect(typeUint, isNonZero, idC1, idC0)
		ctx.StoreRegisterPointer(b, OpScc, sccVal)
	case Sop1OpFlbitI32I64:
		// TODO: this
		typeInt := ctx.GetId(BlockContextIdTypeInt)
		typeUint := ctx.GetId(BlockContextIdTypeUint)
		typeBool := ctx.GetId(BlockContextIdTypeBool)
		idGlsl := ctx.GetId(BlockContextIdGlsl)
		idC30 := ctx.GetConstId(ConstIdxUint30)
		idC31 := ctx.GetConstId(ConstIdxUint31)
		idC62 := ctx.GetConstId(ConstIdxUint62)
		idNeg1 := ctx.GetConstId(ConstIdxUintFFFFFFFF)

		valLo, valHi := ctx.GetOperand64Value(b, details.Src0, instr.Literal)

		// 1. Check high 32 bits. bit 31 of valHi is the sign bit.
		// FindSMsb returns index of first bit != sign bit (0-30), or -1.
		msbHi := b.EmitExtInst(typeInt, idGlsl, SpvGlslOpFindSMsb, b.EmitBitcast(typeInt, valHi))
		isHiNotAllSame := b.EmitINotEqual(typeBool, b.EmitBitcast(typeUint, msbHi), idNeg1)
		resHi := b.EmitISub(typeUint, idC30, b.EmitBitcast(typeUint, msbHi))

		// 2. Check low 32 bits if high bits were all sign bits.
		// signMask = (int32(valHi) >> 31) -> all 0s or all 1s.
		signMask := b.EmitShiftRightArithmetic(typeInt, b.EmitBitcast(typeInt, valHi), idC31)
		// xLo = valLo ^ signMask -> bits are 1 where they differ from sign.
		xLo := b.EmitBitwiseXor(typeUint, valLo, b.EmitBitcast(typeUint, signMask))
		// FindUMsb returns index of first '1' (0-31), or -1.
		msbLo := b.EmitExtInst(typeUint, idGlsl, SpvGlslOpFindUMsb, xLo)
		isLoNotAllSame := b.EmitINotEqual(typeBool, msbLo, idNeg1)
		resLo := b.EmitISub(typeUint, idC62, msbLo)

		// 3. Final result: Hi distance if found, else Lo distance if found, else -1.
		res := b.EmitSelect(typeUint, isHiNotAllSame, resHi, b.EmitSelect(typeUint, isLoNotAllSame, resLo, idNeg1))
		ctx.StoreRegisterPointer(b, details.Dst, res)
	case Sop1OpAndSaveexecB64:
		typeUint := ctx.GetId(BlockContextIdTypeUint)
		typeBool := ctx.GetId(BlockContextIdTypeBool)
		idC0 := ctx.GetConstId(ConstIdxUint0)
		idC1 := ctx.GetConstId(ConstIdxUint1)

		// Dst = EXEC
		execLo, execHi := ctx.GetOperand64Value(b, OpExecLo, 0)
		ctx.StoreRegisterPointer(b, details.Dst, execLo)
		ctx.StoreRegisterPointer(b, details.Dst+1, execHi)

		// EXEC = Src0 & EXEC
		src0Lo, src0Hi := ctx.GetOperand64Value(b, details.Src0, instr.Literal)
		newExecLo := b.EmitBitwiseAnd(typeUint, src0Lo, execLo)
		newExecHi := b.EmitBitwiseAnd(typeUint, src0Hi, execHi)
		ctx.StoreRegisterPointer(b, OpExecLo, newExecLo)
		ctx.StoreRegisterPointer(b, OpExecHi, newExecHi)

		// SCC = (EXEC != 0)
		isNonZeroLo := b.EmitINotEqual(typeBool, newExecLo, idC0)
		isNonZeroHi := b.EmitINotEqual(typeBool, newExecHi, idC0)
		isNonZero := b.EmitLogicalOr(typeBool, isNonZeroLo, isNonZeroHi)
		sccVal := b.EmitSelect(typeUint, isNonZero, idC1, idC0)
		ctx.StoreRegisterPointer(b, OpScc, sccVal)
	default:
		panic(fmt.Sprintf("unknown sop1 op %s", Mnemotics[EncSOP1][details.Op]))
	}
}

func emitWqmDword(b *SpvBuilder, ctx *SpirvBlockContext, val uint32) uint32 {
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	idC1 := ctx.GetConstId(ConstIdxUint1)
	idC2 := ctx.GetConstId(ConstIdxUint2)
	idC3 := ctx.GetConstId(ConstIdxUint3)
	idMask := ctx.GetConstId(ConstIdxUint11111111)

	// Whole quad mode checks each group of four bits in the bitmask;
	// if any bit is set to 1, all four bits are set to 1 in the result.
	// This operation is repeated for the entire bitmask.
	s1 := b.EmitShiftRightLogical(typeUint, val, idC1)
	s2 := b.EmitShiftRightLogical(typeUint, val, idC2)
	s3 := b.EmitShiftRightLogical(typeUint, val, idC3)
	t := b.EmitBitwiseOr(typeUint, val, s1)
	t = b.EmitBitwiseOr(typeUint, t, s2)
	t = b.EmitBitwiseOr(typeUint, t, s3)
	s0 := b.EmitBitwiseAnd(typeUint, t, idMask)
	l1 := b.EmitShiftLeftLogical(typeUint, s0, idC1)
	l2 := b.EmitShiftLeftLogical(typeUint, s0, idC2)
	l3 := b.EmitShiftLeftLogical(typeUint, s0, idC3)
	res := b.EmitBitwiseOr(typeUint, s0, l1)
	res = b.EmitBitwiseOr(typeUint, res, l2)
	res = b.EmitBitwiseOr(typeUint, res, l3)

	return res
}
