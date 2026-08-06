package gcn

import (
	"fmt"

	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
)

func EmitSOP1(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*gcnSpec.ScalarDetails)
	switch details.Op {
	case gcnSpec.Sop1OpMovB32:
		val := ctx.GetOperandUintValue(b, details.Src0, instr.Literal)
		ctx.StoreRegisterPointer(b, details.Dst, val)
	case gcnSpec.Sop1OpMovB64:
		valLo, valHi := ctx.GetOperand64Value(b, details.Src0, instr.Literal)
		ctx.StoreRegisterPointer(b, details.Dst, valLo)
		ctx.StoreRegisterPointer(b, details.Dst+1, valHi)
	case gcnSpec.Sop1OpSwappcB64, gcnSpec.Sop1OpSetpcB64:
		// Already handled via fetch shader.
	/* case gcnSpec.Sop1OpNotB64:
	valLo, valHi := ctx.GetOperand64Value(b, details.Src0, instr.Literal)
	resLo := b.EmitNot(ctx.GetId(BlockContextIdTypeUint), valLo)
	resHi := b.EmitNot(ctx.GetId(BlockContextIdTypeUint), valHi)
	ctx.StoreRegisterPointer(b, details.Dst, resLo)
	ctx.StoreRegisterPointer(b, details.Dst+1, resHi)

	emitSccUpdateNonZero64(b, ctx, resLo, resHi) */
	case gcnSpec.Sop1OpWqmB32:
		val := ctx.GetOperandUintValue(b, details.Src0, instr.Literal)
		res := emitWqmDword(b, ctx, val)
		ctx.StoreRegisterPointer(b, details.Dst, res)

		emitSccUpdateNonZero(b, ctx, res)
	case gcnSpec.Sop1OpWqmB64:
		valLo, valHi := ctx.GetOperand64Value(b, details.Src0, instr.Literal)
		resLo := emitWqmDword(b, ctx, valLo)
		resHi := emitWqmDword(b, ctx, valHi)
		ctx.StoreRegisterPointer(b, details.Dst, resLo)
		ctx.StoreRegisterPointer(b, details.Dst+1, resHi)

		emitSccUpdateNonZero64(b, ctx, resLo, resHi)
	case gcnSpec.Sop1OpFlbitI32I64:
		// TODO: this
		typeInt := ctx.GetId(BlockContextIdTypeInt)
		typeUint := ctx.GetId(BlockContextIdTypeUint)
		typeBool := ctx.GetId(BlockContextIdTypeBool)
		idGlsl := ctx.GetId(BlockContextIdTypeGlsl)
		idC30 := ctx.GetConstId(ConstIdUint30)
		idC31 := ctx.GetConstId(ConstIdUint31)
		idC62 := ctx.GetConstId(ConstIdUint62)
		idNeg1 := ctx.GetConstId(ConstIdUintFFFFFFFF)

		valLo, valHi := ctx.GetOperand64Value(b, details.Src0, instr.Literal)

		// 1. Check high 32 bits. bit 31 of valHi is the sign bit.
		// FindSMsb returns index of first bit != sign bit (0-30), or -1.
		msbHi := b.EmitExtInst(typeInt, idGlsl, spec.SpvGlslOpFindSMsb, b.EmitBitcast(typeInt, valHi))
		isHiNotAllSame := b.EmitINotEqual(typeBool, b.EmitBitcast(typeUint, msbHi), idNeg1)
		resHi := b.EmitISub(typeUint, idC30, b.EmitBitcast(typeUint, msbHi))

		// 2. Check low 32 bits if high bits were all sign bits.
		// signMask = (int32(valHi) >> 31) -> all 0s or all 1s.
		signMask := b.EmitShiftRightArithmetic(typeInt, b.EmitBitcast(typeInt, valHi), idC31)
		// xLo = valLo ^ signMask -> bits are 1 where they differ from sign.
		xLo := b.EmitBitwiseXor(typeUint, valLo, b.EmitBitcast(typeUint, signMask))
		// FindUMsb returns index of first '1' (0-31), or -1.
		msbLo := b.EmitExtInst(typeUint, idGlsl, spec.SpvGlslOpFindUMsb, xLo)
		isLoNotAllSame := b.EmitINotEqual(typeBool, msbLo, idNeg1)
		resLo := b.EmitISub(typeUint, idC62, msbLo)

		// 3. Final result: Hi distance if found, else Lo distance if found, else -1.
		res := b.EmitSelect(typeUint, isHiNotAllSame, resHi, b.EmitSelect(typeUint, isLoNotAllSame, resLo, idNeg1))
		ctx.StoreRegisterPointer(b, details.Dst, res)
	case gcnSpec.Sop1OpAndSaveexecB64:
		typeUint := ctx.GetId(BlockContextIdTypeUint)

		// Dst = EXEC
		execLo, execHi := ctx.GetOperand64Value(b, gcnSpec.OpExecLo, 0)
		ctx.StoreRegisterPointer(b, details.Dst, execLo)
		ctx.StoreRegisterPointer(b, details.Dst+1, execHi)

		// EXEC = Src0 & EXEC
		src0Lo, src0Hi := ctx.GetOperand64Value(b, details.Src0, instr.Literal)
		newExecLo := b.EmitBitwiseAnd(typeUint, src0Lo, execLo)
		newExecHi := b.EmitBitwiseAnd(typeUint, src0Hi, execHi)
		ctx.StoreRegisterPointer(b, gcnSpec.OpExecLo, newExecLo)
		ctx.StoreRegisterPointer(b, gcnSpec.OpExecHi, newExecHi)

		// SCC = (EXEC != 0)
		emitSccUpdateNonZero64(b, ctx, newExecLo, newExecHi)
	/* case gcnSpec.Sop1OpOrSaveexecB64:
		typeUint := ctx.GetId(BlockContextIdTypeUint)

		// Dst = EXEC
		execLo, execHi := ctx.GetOperand64Value(b, gcnSpec.OpExecLo, 0)
		ctx.StoreRegisterPointer(b, details.Dst, execLo)
		ctx.StoreRegisterPointer(b, details.Dst+1, execHi)

		// EXEC = Src0 | EXEC
		src0Lo, src0Hi := ctx.GetOperand64Value(b, details.Src0, instr.Literal)
		newExecLo := b.EmitBitwiseOr(typeUint, src0Lo, execLo)
		newExecHi := b.EmitBitwiseOr(typeUint, src0Hi, execHi)
		ctx.StoreRegisterPointer(b, gcnSpec.OpExecLo, newExecLo)
		ctx.StoreRegisterPointer(b, gcnSpec.OpExecHi, newExecHi)

		// SCC = (EXEC != 0)
		emitSccUpdateNonZero64(b, ctx, newExecLo, newExecHi)
	case gcnSpec.Sop1OpXorSaveexecB64:
		typeUint := ctx.GetId(BlockContextIdTypeUint)

		// Dst = EXEC
		execLo, execHi := ctx.GetOperand64Value(b, gcnSpec.OpExecLo, 0)
		ctx.StoreRegisterPointer(b, details.Dst, execLo)
		ctx.StoreRegisterPointer(b, details.Dst+1, execHi)

		// EXEC = Src0 ^ EXEC
		src0Lo, src0Hi := ctx.GetOperand64Value(b, details.Src0, instr.Literal)
		newExecLo := b.EmitBitwiseXor(typeUint, src0Lo, execLo)
		newExecHi := b.EmitBitwiseXor(typeUint, src0Hi, execHi)
		ctx.StoreRegisterPointer(b, gcnSpec.OpExecLo, newExecLo)
		ctx.StoreRegisterPointer(b, gcnSpec.OpExecHi, newExecHi)

		// SCC = (EXEC != 0)
		emitSccUpdateNonZero64(b, ctx, newExecLo, newExecHi)
	case gcnSpec.Sop1OpAndn2SaveexecB64:
		typeUint := ctx.GetId(BlockContextIdTypeUint)

		// Dst = EXEC
		execLo, execHi := ctx.GetOperand64Value(b, gcnSpec.OpExecLo, 0)
		ctx.StoreRegisterPointer(b, details.Dst, execLo)
		ctx.StoreRegisterPointer(b, details.Dst+1, execHi)

		// EXEC = Src0 & ~EXEC
		src0Lo, src0Hi := ctx.GetOperand64Value(b, details.Src0, instr.Literal)
		newExecLo := b.EmitBitwiseAnd(typeUint, src0Lo, b.EmitNot(typeUint, execLo))
		newExecHi := b.EmitBitwiseAnd(typeUint, src0Hi, b.EmitNot(typeUint, execHi))
		ctx.StoreRegisterPointer(b, gcnSpec.OpExecLo, newExecLo)
		ctx.StoreRegisterPointer(b, gcnSpec.OpExecHi, newExecHi)

		// SCC = (EXEC != 0)
		emitSccUpdateNonZero64(b, ctx, newExecLo, newExecHi)
	case gcnSpec.Sop1OpOrn2SaveexecB64:
		typeUint := ctx.GetId(BlockContextIdTypeUint)

		// Dst = EXEC
		execLo, execHi := ctx.GetOperand64Value(b, gcnSpec.OpExecLo, 0)
		ctx.StoreRegisterPointer(b, details.Dst, execLo)
		ctx.StoreRegisterPointer(b, details.Dst+1, execHi)

		// EXEC = Src0 | ~EXEC
		src0Lo, src0Hi := ctx.GetOperand64Value(b, details.Src0, instr.Literal)
		newExecLo := b.EmitBitwiseOr(typeUint, src0Lo, b.EmitNot(typeUint, execLo))
		newExecHi := b.EmitBitwiseOr(typeUint, src0Hi, b.EmitNot(typeUint, execHi))
		ctx.StoreRegisterPointer(b, gcnSpec.OpExecLo, newExecLo)
		ctx.StoreRegisterPointer(b, gcnSpec.OpExecHi, newExecHi)

		// SCC = (EXEC != 0)
		emitSccUpdateNonZero64(b, ctx, newExecLo, newExecHi) */
	default:
		panic(fmt.Sprintf("unknown sop1 op %s", gcnSpec.Mnemotics[gcnSpec.EncSOP1][details.Op]))
	}
}

func emitSccUpdateNonZero(b *SpvBuilder, ctx *SpirvBlockContext, val SpirvId) {
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	typeBool := ctx.GetId(BlockContextIdTypeBool)
	idC0 := ctx.GetConstId(ConstIdUint0)
	idC1 := ctx.GetConstId(ConstIdUint1)

	isNonZero := b.EmitINotEqual(typeBool, val, idC0)
	ctx.StoreRegisterPointer(b, gcnSpec.OpScc, b.EmitSelect(typeUint, isNonZero, idC1, idC0))
}

func emitSccUpdateNonZero64(b *SpvBuilder, ctx *SpirvBlockContext, valLo, valHi SpirvId) {
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	typeBool := ctx.GetId(BlockContextIdTypeBool)
	idC0 := ctx.GetConstId(ConstIdUint0)
	idC1 := ctx.GetConstId(ConstIdUint1)

	isNonZeroLo := b.EmitINotEqual(typeBool, valLo, idC0)
	isNonZeroHi := b.EmitINotEqual(typeBool, valHi, idC0)
	isNonZero := b.EmitLogicalOr(typeBool, isNonZeroLo, isNonZeroHi)
	ctx.StoreRegisterPointer(b, gcnSpec.OpScc, b.EmitSelect(typeUint, isNonZero, idC1, idC0))
}

func emitWqmDword(b *SpvBuilder, ctx *SpirvBlockContext, val SpirvId) SpirvId {
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	idC1 := ctx.GetConstId(ConstIdUint1)
	idC2 := ctx.GetConstId(ConstIdUint2)
	idC3 := ctx.GetConstId(ConstIdUint3)
	idMask := ctx.GetConstId(ConstIdUint11111111)

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
