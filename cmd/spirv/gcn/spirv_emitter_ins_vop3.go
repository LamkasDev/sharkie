package gcn

import (
	"fmt"

	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
)

func EmitVOP3(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*gcnSpec.Vop3Details)
	origDetails := instr.Details
	defer func() { instr.Details = origDetails }()
	switch {
	case details.Op <= 0xFF: // VOPC
		instr.Details = &gcnSpec.VectorDetails{
			Op:    details.Op,
			Vdst:  details.Vdst,
			Src0:  details.Src0,
			Src1:  details.Src1,
			Sdst:  details.Sdst,
			Abs:   details.Abs,
			Neg:   details.Neg,
			OMod:  details.OMod,
			Clamp: details.Clamp,
			Src2:  details.Src2,
		}
		EmitVOPC(b, instr, ctx)
	case details.Op >= gcnSpec.Vop3OpCndmaskB32 && details.Op <= gcnSpec.Vop3OpCvtPkI16I32: // VOP2
		instr.Details = &gcnSpec.VectorDetails{
			Op:    details.Op - gcnSpec.Vop3OpCndmaskB32,
			Vdst:  details.Vdst,
			Src0:  details.Src0,
			Src1:  details.Src1,
			Sdst:  details.Sdst,
			Abs:   details.Abs,
			Neg:   details.Neg,
			OMod:  details.OMod,
			Clamp: details.Clamp,
			Src2:  details.Src2,
		}
		EmitVOP2(b, instr, ctx)
	case details.Op >= gcnSpec.Vop3OpNop && details.Op <= gcnSpec.Vop3OpMovrelsdB32: // VOP1
		instr.Details = &gcnSpec.VectorDetails{
			Op:    details.Op - gcnSpec.Vop3OpNop,
			Vdst:  details.Vdst,
			Src0:  details.Src0,
			Sdst:  details.Sdst,
			Abs:   details.Abs,
			Neg:   details.Neg,
			OMod:  details.OMod,
			Clamp: details.Clamp,
			Src2:  details.Src2,
		}
		EmitVOP1(b, instr, ctx)
	case details.Op == gcnSpec.Vop3OpMadF32:
		val0 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src1, instr.Literal, 1)
		val2 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src2, instr.Literal, 2)

		resF := b.EmitExtInst(ctx.GetId(BlockContextIdTypeFloat), ctx.GetId(BlockContextIdTypeGlsl), spec.SpvGlslOpFma, val0, val1, val2)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	case details.Op == gcnSpec.Vop3OpMin3F32:
		// TODO: add SPV_AMD_shader_trinary_minmax optimized version.
		typeFloat := ctx.GetId(BlockContextIdTypeFloat)
		typeGlsl := ctx.GetId(BlockContextIdTypeGlsl)

		val0 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src1, instr.Literal, 1)
		val2 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src2, instr.Literal, 2)

		// D.f = min(S0.f, min(S1.f, S2.f))
		min01 := b.EmitExtInst(typeFloat, typeGlsl, spec.SpvGlslOpFMin, val0, val1)
		resF := b.EmitExtInst(typeFloat, typeGlsl, spec.SpvGlslOpFMin, min01, val2)

		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	case details.Op == gcnSpec.Vop3OpMax3F32:
		// TODO: add SPV_AMD_shader_trinary_minmax optimized version.
		typeFloat := ctx.GetId(BlockContextIdTypeFloat)
		typeGlsl := ctx.GetId(BlockContextIdTypeGlsl)

		val0 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src1, instr.Literal, 1)
		val2 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src2, instr.Literal, 2)

		// D.f = max(S0.f, max(S1.f, S2.f))
		max01 := b.EmitExtInst(typeFloat, typeGlsl, spec.SpvGlslOpFMax, val0, val1)
		resF := b.EmitExtInst(typeFloat, typeGlsl, spec.SpvGlslOpFMax, max01, val2)

		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	case details.Op == gcnSpec.Vop3OpMulLoI32:
		typeInt := ctx.GetId(BlockContextIdTypeInt)

		val0 := GetOperandIntValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandIntValueModified(b, ctx, details.Abs, details.Neg, details.Src1, instr.Literal, 1)

		resI := b.EmitIMul(typeInt, val0, val1)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resI, false)
	case details.Op == gcnSpec.Vop3OpMulLoU32:
		typeUint := ctx.GetId(BlockContextIdTypeUint)

		val0 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src1, instr.Literal, 1)

		resI := b.EmitIMul(typeUint, val0, val1)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resI, false)
	case details.Op == gcnSpec.Vop3OpMed3F32:
		// TODO: add SPV_AMD_shader_trinary_minmax optimized version.
		typeFloat := ctx.GetId(BlockContextIdTypeFloat)
		typeBool := ctx.GetId(BlockContextIdTypeBool)
		typeGlsl := ctx.GetId(BlockContextIdTypeGlsl)

		val0 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src1, instr.Literal, 1)
		val2 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src2, instr.Literal, 2)

		// isNan(S0.f) || isNan(S1.f) || isNan(S2.f)
		nan0 := b.EmitIsNan(typeBool, val0)
		nan1 := b.EmitIsNan(typeBool, val1)
		nan2 := b.EmitIsNan(typeBool, val2)
		anyNan := b.EmitLogicalOr(typeBool, nan0, b.EmitLogicalOr(typeBool, nan1, nan2))

		// MIN3(S0.f, S1.f, S2.f)
		min01 := b.EmitExtInst(typeFloat, typeGlsl, spec.SpvGlslOpFMin, val0, val1)
		min3 := b.EmitExtInst(typeFloat, typeGlsl, spec.SpvGlslOpFMin, min01, val2)

		// MAX3(S0.f, S1.f, S2.f)
		max01 := b.EmitExtInst(typeFloat, typeGlsl, spec.SpvGlslOpFMax, val0, val1)
		max3 := b.EmitExtInst(typeFloat, typeGlsl, spec.SpvGlslOpFMax, max01, val2)

		// MAX3 == S0.f, MAX3 == S1.f
		isMax0 := b.EmitFOrdEqual(typeBool, max3, val0)
		isMax1 := b.EmitFOrdEqual(typeBool, max3, val1)

		// D.f = MAX(S1.f, S2.f) if MAX3 == S0
		// D.f = MAX(S0.f, S2.f) if MAX3 == S1
		// Else D.f = MAX(S0.f, S1.f)
		max12 := b.EmitExtInst(typeFloat, typeGlsl, spec.SpvGlslOpFMax, val1, val2)
		max02 := b.EmitExtInst(typeFloat, typeGlsl, spec.SpvGlslOpFMax, val0, val2)
		max01_2 := b.EmitExtInst(typeFloat, typeGlsl, spec.SpvGlslOpFMax, val0, val1)

		res := b.EmitSelect(typeFloat, isMax0, max12, b.EmitSelect(typeFloat, isMax1, max02, max01_2))

		// Final result based on anyNan
		finalRes := b.EmitSelect(typeFloat, anyNan, min3, res)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, finalRes, true)
	case details.Op == gcnSpec.Vop3OpBfeU32:
		typeUint := ctx.GetId(BlockContextIdTypeUint)
		typeBool := ctx.GetId(BlockContextIdTypeBool)
		idC0 := ctx.GetConstId(ConstIdUint0)

		val0 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src1, instr.Literal, 1)
		val2 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src2, instr.Literal, 2)

		// offset = src1[4:0]
		// width = src2[4:0]
		mask31 := ctx.GetConstId(ConstIdUint31)
		offset := b.EmitBitwiseAnd(typeUint, val1, mask31)
		width := b.EmitBitwiseAnd(typeUint, val2, mask31)

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
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, res, false)
	case details.Op == gcnSpec.Vop3OpBfeI32:
		typeInt := ctx.GetId(BlockContextIdTypeInt)
		typeUint := ctx.GetId(BlockContextIdTypeUint)
		typeBool := ctx.GetId(BlockContextIdTypeBool)

		val0 := GetOperandIntValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src1, instr.Literal, 1)
		val2 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src2, instr.Literal, 2)

		// offset = src1[4:0]
		// width = src2[4:0]
		mask31 := ctx.GetConstId(ConstIdUint31)
		offset := b.EmitBitwiseAnd(typeUint, val1, mask31)
		width := b.EmitBitwiseAnd(typeUint, val2, mask31)

		// If (width == 0) dst = 0
		// Else if (width + offset <= 32) dst = bitfieldSExtract(src0, offset, width)
		// Else dst = src0 >> offset
		isWidthZero := b.EmitIEqual(typeBool, width, ctx.GetConstId(ConstIdUint0))
		isShortExtract := b.EmitULessThan(typeBool, b.EmitIAdd(typeUint, width, offset), ctx.GetConstId(ConstIdUint33))

		// Short extract: bitfieldSExtract(src0, offset, width)
		resShort := b.EmitBitFieldSExtract(typeInt, val0, offset, width)

		// Long extract: src0 >> offset
		resLong := b.EmitShiftRightArithmetic(typeInt, val0, offset)

		res := b.EmitSelect(typeInt, isWidthZero, b.EmitConstantUint(typeInt, 0), b.EmitSelect(typeInt, isShortExtract, resShort, resLong))
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, b.EmitBitcast(typeUint, res), false)
	case details.Op == gcnSpec.Vop3OpBfiB32:
		typeUint := ctx.GetId(BlockContextIdTypeUint)

		val0 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src1, instr.Literal, 1)
		val2 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src2, instr.Literal, 2)

		// D.u = (S0.u & S1.u) | (~S0.u & S2.u)
		and01 := b.EmitBitwiseAnd(typeUint, val0, val1)
		not0 := b.EmitNot(typeUint, val0)
		andNot02 := b.EmitBitwiseAnd(typeUint, not0, val2)
		res := b.EmitBitwiseOr(typeUint, and01, andNot02)

		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, res, false)
	case details.Op == gcnSpec.Vop3OpSadU32:
		typeUint := ctx.GetId(BlockContextIdTypeUint)
		typeBool := ctx.GetId(BlockContextIdTypeBool)

		val0 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src1, instr.Literal, 1)
		val2 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src2, instr.Literal, 2)

		// (A > B)
		isGreater := b.EmitUGreaterThan(typeBool, val0, val1)

		// (A - B) and (B - A)
		subAB := b.EmitISub(typeUint, val0, val1)
		subBA := b.EmitISub(typeUint, val1, val0)

		// ABS_DIFF(A,B) = (A > B) ? (A - B) : (B - A)
		absDiff := b.EmitSelect(typeUint, isGreater, subAB, subBA)

		// D.u = ABS_DIFF(S0.u, S1.u) + S2.u
		res := b.EmitIAdd(typeUint, absDiff, val2)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, res, false)
	case details.Op == gcnSpec.Vop3OpMadU32U24:
		typeUint := ctx.GetId(BlockContextIdTypeUint)
		idMask := b.EmitConstantUint(typeUint, 0xFFFFFF)

		val0 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src1, instr.Literal, 1)
		val2 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src2, instr.Literal, 2)

		val0 = b.EmitBitwiseAnd(typeUint, val0, idMask)
		val1 = b.EmitBitwiseAnd(typeUint, val1, idMask)

		res := b.EmitIAdd(typeUint, b.EmitIMul(typeUint, val0, val1), val2)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, res, false)
	case details.Op == gcnSpec.Vop3OpMadI32I24:
		typeInt := ctx.GetId(BlockContextIdTypeInt)
		typeUint := ctx.GetId(BlockContextIdTypeUint)

		val0 := GetOperandIntValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandIntValueModified(b, ctx, details.Abs, details.Neg, details.Src1, instr.Literal, 1)
		val2 := GetOperandIntValueModified(b, ctx, details.Abs, details.Neg, details.Src2, instr.Literal, 2)

		// Set up offset (0) and count (24) for the 24-bit bitfield extraction.
		offset := b.EmitConstantUint(typeUint, 0)
		count := b.EmitConstantUint(typeUint, 24)

		// Extract the lower 24 bits and sign-extend them to 32 bits.
		val0Ext := b.EmitBitFieldSExtract(typeInt, val0, offset, count)
		val1Ext := b.EmitBitFieldSExtract(typeInt, val1, offset, count)

		// Result = (S0 * S1) + S2.
		mul := b.EmitIMul(typeInt, val0Ext, val1Ext)
		res := b.EmitIAdd(typeInt, mul, val2)

		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, res, false)
	case details.Op == gcnSpec.Vop3OpFmaF32:
		val0 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src1, instr.Literal, 1)
		val2 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src2, instr.Literal, 2)

		resF := b.EmitExtInst(ctx.GetId(BlockContextIdTypeFloat), ctx.GetId(BlockContextIdTypeGlsl), spec.SpvGlslOpFma, val0, val1, val2)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	default:
		panic(fmt.Sprintf("unknown vop3 op %s", gcnSpec.Mnemotics[gcnSpec.EncVOP3][details.Op]))
	}
}

func GetOperandFloatValueModified(b *SpvBuilder, ctx *SpirvBlockContext, abs uint8, neg uint8, src uint32, literal uint32, i int) SpirvId {
	val := ctx.GetOperandFloatValue(b, src, literal)
	return applyVop3Modifiers(b, ctx, val, abs, neg, i)
}

func GetOperandUintValueModified(b *SpvBuilder, ctx *SpirvBlockContext, abs uint8, neg uint8, src uint32, literal uint32, i int) SpirvId {
	typeFloat := ctx.GetId(BlockContextIdTypeFloat)
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	val := ctx.GetOperandUintValue(b, src, literal)
	val = b.EmitBitcast(typeUint, applyVop3Modifiers(b, ctx, b.EmitBitcast(typeFloat, val), abs, neg, i))
	return val
}

func GetOperandIntValueModified(b *SpvBuilder, ctx *SpirvBlockContext, abs uint8, neg uint8, src uint32, literal uint32, i int) SpirvId {
	typeFloat := ctx.GetId(BlockContextIdTypeFloat)
	typeInt := ctx.GetId(BlockContextIdTypeInt)
	val := ctx.GetOperandIntValue(b, src, literal)
	val = b.EmitBitcast(typeInt, applyVop3Modifiers(b, ctx, b.EmitBitcast(typeFloat, val), abs, neg, i))
	return val
}

func StoreRegisterPointerMaskedModified(b *SpvBuilder, ctx *SpirvBlockContext, clamp bool, omod uint8, dst uint32, res SpirvId, isFloat bool) {
	typeFloat := ctx.GetId(BlockContextIdTypeFloat)
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	if !isFloat {
		res = b.EmitBitcast(typeFloat, res)
	}
	res = applyVop3OutputModifiers(b, ctx, res, clamp, omod)
	ctx.StoreRegisterPointerMasked(b, dst, b.EmitBitcast(typeUint, res))
}

func applyVop3Modifiers(b *SpvBuilder, ctx *SpirvBlockContext, val SpirvId, abs, neg uint8, i int) SpirvId {
	typeFloat := ctx.GetId(BlockContextIdTypeFloat)
	if (abs>>i)&1 == 1 {
		val = b.EmitExtInst(typeFloat, ctx.GetId(BlockContextIdTypeGlsl), spec.SpvGlslOpFAbs, val)
	}
	if (neg>>i)&1 == 1 {
		val = b.EmitFNegate(typeFloat, val)
	}

	return val
}

func applyVop3OutputModifiers(b *SpvBuilder, ctx *SpirvBlockContext, val SpirvId, clamp bool, omod uint8) SpirvId {
	typeFloat := ctx.GetId(BlockContextIdTypeFloat)
	idGlsl := ctx.GetId(BlockContextIdTypeGlsl)

	// Apply omod.
	switch omod {
	case 1:
		val = b.EmitFMul(typeFloat, val, ctx.GetConstId(ConstIdFloat2))
	case 2:
		val = b.EmitFMul(typeFloat, val, ctx.GetConstId(ConstIdFloat4))
	case 3:
		val = b.EmitFMul(typeFloat, val, ctx.GetConstId(ConstIdFloat05))
	}

	// Apply clamp.
	if clamp {
		idC0 := ctx.GetConstId(ConstIdFloat0)
		idC1 := ctx.GetConstId(ConstIdFloat1)
		val = b.EmitExtInst(typeFloat, idGlsl, spec.SpvGlslOpFClamp, val, idC0, idC1)
	}

	return val
}
