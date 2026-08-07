package gcn

import (
	"fmt"

	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
)

func EmitSOP2(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*gcnSpec.ScalarDetails)
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	typeBool := ctx.GetId(BlockContextIdTypeBool)
	idC0 := ctx.GetConstId(ConstIdUint0)

	switch details.Op {
	case gcnSpec.Sop2OpCselectB32:
		val0 := ctx.GetOperandUintValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandUintValue(b, details.Src1, instr.Literal)
		scc := ctx.LoadRegisterPointer(b, gcnSpec.OpScc)

		isSccNonZero := b.EmitINotEqual(typeBool, scc, idC0)
		res := b.EmitSelect(typeUint, isSccNonZero, val0, val1)
		ctx.StoreRegisterPointer(b, details.Dst, res)
	/* case gcnSpec.Sop2OpCselectB64:
	val0Lo, val0Hi := ctx.GetOperand64Value(b, details.Src0, instr.Literal)
	val1Lo, val1Hi := ctx.GetOperand64Value(b, details.Src1, instr.Literal)
	scc := ctx.LoadRegisterPointer(b, gcnSpec.OpScc)

	isSccNonZero := b.EmitINotEqual(typeBool, scc, idC0)
	resLo := b.EmitSelect(typeUint, isSccNonZero, val0Lo, val1Lo)
	resHi := b.EmitSelect(typeUint, isSccNonZero, val0Hi, val1Hi)
	ctx.StoreRegisterPointer(b, details.Dst, resLo)
	ctx.StoreRegisterPointer(b, details.Dst+1, resHi) */
	case gcnSpec.Sop2OpLshlB32:
		val0 := ctx.GetOperandUintValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandUintValue(b, details.Src1, instr.Literal)

		// D.u = S0.u << S1.u[4:0]
		shift := b.EmitBitwiseAnd(typeUint, val1, ctx.GetConstId(ConstIdUint31))
		res := b.EmitShiftLeftLogical(typeUint, val0, shift)
		ctx.StoreRegisterPointer(b, details.Dst, res)

		// SCC = 1 if result is non-zero.
		emitSccUpdateNonZero(b, ctx, res)
	case gcnSpec.Sop2OpAndB32:
		val0 := ctx.GetOperandUintValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandUintValue(b, details.Src1, instr.Literal)

		res := b.EmitBitwiseAnd(typeUint, val0, val1)
		ctx.StoreRegisterPointer(b, details.Dst, res)

		// SCC = 1 if result is non-zero.
		emitSccUpdateNonZero(b, ctx, res)
	case gcnSpec.Sop2OpAndB64:
		val0Lo, val0Hi := ctx.GetOperand64Value(b, details.Src0, instr.Literal)
		val1Lo, val1Hi := ctx.GetOperand64Value(b, details.Src1, instr.Literal)

		resLo := b.EmitBitwiseAnd(typeUint, val0Lo, val1Lo)
		resHi := b.EmitBitwiseAnd(typeUint, val0Hi, val1Hi)
		ctx.StoreRegisterPointer(b, details.Dst, resLo)
		ctx.StoreRegisterPointer(b, details.Dst+1, resHi)

		// SCC = 1 if result is non-zero.
		emitSccUpdateNonZero64(b, ctx, resLo, resHi)
	case gcnSpec.Sop2OpOrB64:
		val0Lo, val0Hi := ctx.GetOperand64Value(b, details.Src0, instr.Literal)
		val1Lo, val1Hi := ctx.GetOperand64Value(b, details.Src1, instr.Literal)

		resLo := b.EmitBitwiseOr(typeUint, val0Lo, val1Lo)
		resHi := b.EmitBitwiseOr(typeUint, val0Hi, val1Hi)
		ctx.StoreRegisterPointer(b, details.Dst, resLo)
		ctx.StoreRegisterPointer(b, details.Dst+1, resHi)

		// SCC = 1 if result is non-zero.
		emitSccUpdateNonZero64(b, ctx, resLo, resHi)
	case gcnSpec.Sop2OpNorB64:
		val0Lo, val0Hi := ctx.GetOperand64Value(b, details.Src0, instr.Literal)
		val1Lo, val1Hi := ctx.GetOperand64Value(b, details.Src1, instr.Literal)

		orLo := b.EmitBitwiseOr(typeUint, val0Lo, val1Lo)
		orHi := b.EmitBitwiseOr(typeUint, val0Hi, val1Hi)
		resLo := b.EmitNot(typeUint, orLo)
		resHi := b.EmitNot(typeUint, orHi)
		ctx.StoreRegisterPointer(b, details.Dst, resLo)
		ctx.StoreRegisterPointer(b, details.Dst+1, resHi)

		// SCC = 1 if result is non-zero.
		emitSccUpdateNonZero64(b, ctx, resLo, resHi)
	/* case gcnSpec.Sop2OpXorB64:
	val0Lo, val0Hi := ctx.GetOperand64Value(b, details.Src0, instr.Literal)
	val1Lo, val1Hi := ctx.GetOperand64Value(b, details.Src1, instr.Literal)

	resLo := b.EmitBitwiseXor(typeUint, val0Lo, val1Lo)
	resHi := b.EmitBitwiseXor(typeUint, val0Hi, val1Hi)
	ctx.StoreRegisterPointer(b, details.Dst, resLo)
	ctx.StoreRegisterPointer(b, details.Dst+1, resHi)

	// SCC = 1 if result is non-zero.
	emitSccUpdateNonZero64(b, ctx, resLo, resHi) */
	case gcnSpec.Sop2OpAndn2B64:
		val0Lo, val0Hi := ctx.GetOperand64Value(b, details.Src0, instr.Literal)
		val1Lo, val1Hi := ctx.GetOperand64Value(b, details.Src1, instr.Literal)

		not1Lo := b.EmitNot(typeUint, val1Lo)
		not1Hi := b.EmitNot(typeUint, val1Hi)

		resLo := b.EmitBitwiseAnd(typeUint, val0Lo, not1Lo)
		resHi := b.EmitBitwiseAnd(typeUint, val0Hi, not1Hi)
		ctx.StoreRegisterPointer(b, details.Dst, resLo)
		ctx.StoreRegisterPointer(b, details.Dst+1, resHi)

		// SCC = 1 if result is non-zero.
		emitSccUpdateNonZero64(b, ctx, resLo, resHi)
	case gcnSpec.Sop2OpBfeU32:
		val0 := ctx.GetOperandUintValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandUintValue(b, details.Src1, instr.Literal)

		// offset = src1[6:0]
		// width = src1[22:16]
		offset := b.EmitBitwiseAnd(typeUint, val1, ctx.GetConstId(ConstIdUint7F))
		width := b.EmitBitwiseAnd(typeUint, b.EmitShiftRightLogical(typeUint, val1, ctx.GetConstId(ConstIdUint16)), ctx.GetConstId(ConstIdUint7F))

		// If (width == 0) dst = 0
		// Else if (width + offset <= 32) dst = bitfieldUExtract(src0, offset, width)
		// Else dst = src0 >> offset
		isWidthZero := b.EmitIEqual(typeBool, width, idC0)
		isShortExtract := b.EmitULessThan(typeBool, b.EmitIAdd(typeUint, width, offset), ctx.GetConstId(ConstIdUint33))

		// Short extract: bitfieldUExtract(src0, offset, width)
		resShort := b.EmitBitFieldUExtract(typeUint, val0, offset, width)

		// Long extract: src0 >> offset
		resLong := b.EmitShiftRightLogical(typeUint, val0, offset)

		res := b.EmitSelect(typeUint, isWidthZero, idC0, b.EmitSelect(typeUint, isShortExtract, resShort, resLong))
		ctx.StoreRegisterPointer(b, details.Dst, res)

		// SCC = 1 if result is non-zero.
		emitSccUpdateNonZero(b, ctx, res)
	case gcnSpec.Sop2OpMulI32:
		val0 := ctx.GetOperandIntValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandIntValue(b, details.Src1, instr.Literal)
		res := b.EmitIMul(typeUint, val0, val1)
		ctx.StoreRegisterPointer(b, details.Dst, res)
	case gcnSpec.Sop2OpAddU32:
		val0 := ctx.GetOperandUintValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandUintValue(b, details.Src1, instr.Literal)

		// D.u = S0.u + S1.u
		res := b.EmitIAdd(typeUint, val0, val1)
		ctx.StoreRegisterPointer(b, details.Dst, res)

		// SCC = unsigned carry out
		// carry occurs if res < val0
		isCarry := b.EmitULessThan(typeBool, res, val0)
		sccVal := b.EmitSelect(typeUint, isCarry, b.EmitConstantUint(typeUint, 1), idC0)
		ctx.StoreRegisterPointer(b, gcnSpec.OpScc, sccVal)
	case gcnSpec.Sop2OpAddI32:
		val0 := ctx.GetOperandUintValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandUintValue(b, details.Src1, instr.Literal)

		// D.u = S0.i + S1.i
		res := b.EmitIAdd(typeUint, val0, val1)
		ctx.StoreRegisterPointer(b, details.Dst, res)

		// SCC = signed overflow
		// Overflow logic: ((~(val0 ^ val1)) & (val0 ^ res)) & 0x80000000 != 0
		xor01 := b.EmitBitwiseXor(typeUint, val0, val1)
		notXor01 := b.EmitNot(typeUint, xor01)
		xor0Res := b.EmitBitwiseXor(typeUint, val0, res)

		andMask := b.EmitBitwiseAnd(typeUint, notXor01, xor0Res)
		signBit := b.EmitConstantUint(typeUint, 0x80000000)
		overflowBits := b.EmitBitwiseAnd(typeUint, andMask, signBit)

		isOverflow := b.EmitINotEqual(typeBool, overflowBits, idC0)

		// Store 1 into SCC if overflowed, else 0.
		sccVal := b.EmitSelect(typeUint, isOverflow, b.EmitConstantUint(typeUint, 1), idC0)
		ctx.StoreRegisterPointer(b, gcnSpec.OpScc, sccVal)
	case gcnSpec.Sop2OpLshrB32:
		val0 := ctx.GetOperandUintValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandUintValue(b, details.Src1, instr.Literal)
		shift := b.EmitBitwiseAnd(typeUint, val1, ctx.GetConstId(ConstIdUint31))
		res := b.EmitShiftRightLogical(typeUint, val0, shift)
		ctx.StoreRegisterPointer(b, details.Dst, res)

		// SCC = 1 if result is non-zero.
		emitSccUpdateNonZero(b, ctx, res)
	default:
		panic(fmt.Sprintf("unknown sop2 op %s", gcnSpec.Mnemotics[gcnSpec.EncSOP2][details.Op]))
	}
}
