package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitSOP2(b *SpvBuilder, instr *Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*ScalarDetails)
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	typeBool := ctx.GetId(BlockContextIdTypeBool)
	idC0 := ctx.GetConstId(ConstIdxUint0)
	idC1 := ctx.GetConstId(ConstIdxUint1)

	switch details.Op {
	case Sop2OpCselectB32:
		val0 := ctx.GetOperandUintValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandUintValue(b, details.Src1, instr.Literal)
		scc := ctx.LoadRegisterPointer(b, OpScc)

		isSccNonZero := b.EmitINotEqual(typeBool, scc, idC0)
		res := b.EmitSelect(typeUint, isSccNonZero, val0, val1)
		ctx.StoreRegisterPointer(b, details.Dst, res)
	case Sop2OpAndB32:
		val0 := ctx.GetOperandUintValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandUintValue(b, details.Src1, instr.Literal)

		res := b.EmitBitwiseAnd(typeUint, val0, val1)
		ctx.StoreRegisterPointer(b, details.Dst, res)

		// SCC = 1 if result is non-zero.
		isNonZero := b.EmitINotEqual(typeBool, res, idC0)
		resScc := b.EmitSelect(typeUint, isNonZero, idC1, idC0)
		ctx.StoreRegisterPointer(b, OpScc, resScc)
	case Sop2OpAndn2B64:
		val0Lo, val0Hi := ctx.GetOperand64Value(b, details.Src0, instr.Literal)
		val1Lo, val1Hi := ctx.GetOperand64Value(b, details.Src1, instr.Literal)

		not1Lo := b.EmitNot(typeUint, val1Lo)
		not1Hi := b.EmitNot(typeUint, val1Hi)

		resLo := b.EmitBitwiseAnd(typeUint, val0Lo, not1Lo)
		resHi := b.EmitBitwiseAnd(typeUint, val0Hi, not1Hi)

		ctx.StoreRegisterPointer(b, details.Dst, resLo)
		ctx.StoreRegisterPointer(b, details.Dst+1, resHi)

		// SCC = 1 if result is non-zero.
		nzLo := b.EmitINotEqual(typeBool, resLo, idC0)
		nzHi := b.EmitINotEqual(typeBool, resHi, idC0)
		anyNz := b.EmitLogicalOr(typeBool, nzLo, nzHi)

		resScc := b.EmitSelect(typeUint, anyNz, idC1, idC0)
		ctx.StoreRegisterPointer(b, OpScc, resScc)
	case Sop2OpBfeU32:
		val0 := ctx.GetOperandUintValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandUintValue(b, details.Src1, instr.Literal)

		// offset = src1[6:0]
		// width = src1[22:16]
		offset := b.EmitBitwiseAnd(typeUint, val1, ctx.GetConstId(ConstIdxUint7F))
		width := b.EmitBitwiseAnd(typeUint, b.EmitShiftRightLogical(typeUint, val1, ctx.GetConstId(ConstIdxUint16)), ctx.GetConstId(ConstIdxUint7F))

		// If (width == 0) dst = 0
		// Else if (width + offset <= 32) dst = bitfieldUExtract(src0, offset, width)
		// Else dst = src0 >> offset
		isWidthZero := b.EmitIEqual(typeBool, width, idC0)
		isShortExtract := b.EmitULessThan(typeBool, b.EmitIAdd(typeUint, width, offset), ctx.GetConstId(ConstIdxUint33))

		// Short extract: bitfieldUExtract(src0, offset, width)
		resShort := b.EmitBitFieldUExtract(typeUint, val0, offset, width)

		// Long extract: src0 >> offset
		resLong := b.EmitShiftRightLogical(typeUint, val0, offset)

		res := b.EmitSelect(typeUint, isWidthZero, idC0, b.EmitSelect(typeUint, isShortExtract, resShort, resLong))
		ctx.StoreRegisterPointer(b, details.Dst, res)

		// SCC = 1 if result is non-zero.
		isNonZero := b.EmitINotEqual(typeBool, res, idC0)
		resScc := b.EmitSelect(typeUint, isNonZero, idC1, idC0)
		ctx.StoreRegisterPointer(b, OpScc, resScc)
	default:
		panic(fmt.Sprintf("unknown sop2 op %s", Mnemotics[EncSOP2][details.Op]))
	}
}
