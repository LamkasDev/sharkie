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
		}
		EmitVOP1(b, instr, ctx)
	case details.Op == gcnSpec.Vop3OpMadF32:
		val0 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src1, instr.Literal, 1)
		val2 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src2, instr.Literal, 2)

		resF := b.EmitExtInst(ctx.GetId(BlockContextIdTypeFloat), ctx.GetId(BlockContextIdGlsl), spec.SpvGlslOpFma, val0, val1, val2)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	case details.Op == gcnSpec.Vop3OpMed3F32:
		// TODO: add SPV_AMD_shader_trinary_minmax optimized version.
		typeFloat := ctx.GetId(BlockContextIdTypeFloat)
		typeBool := ctx.GetId(BlockContextIdTypeBool)
		idGlsl := ctx.GetId(BlockContextIdGlsl)

		val0 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src1, instr.Literal, 1)
		val2 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src2, instr.Literal, 2)

		// isNan(S0.f) || isNan(S1.f) || isNan(S2.f)
		nan0 := b.EmitIsNan(typeBool, val0)
		nan1 := b.EmitIsNan(typeBool, val1)
		nan2 := b.EmitIsNan(typeBool, val2)
		anyNan := b.EmitLogicalOr(typeBool, nan0, b.EmitLogicalOr(typeBool, nan1, nan2))

		// MIN3(S0.f, S1.f, S2.f)
		min01 := b.EmitExtInst(typeFloat, idGlsl, spec.SpvGlslOpFMin, val0, val1)
		min3 := b.EmitExtInst(typeFloat, idGlsl, spec.SpvGlslOpFMin, min01, val2)

		// MAX3(S0.f, S1.f, S2.f)
		max01 := b.EmitExtInst(typeFloat, idGlsl, spec.SpvGlslOpFMax, val0, val1)
		max3 := b.EmitExtInst(typeFloat, idGlsl, spec.SpvGlslOpFMax, max01, val2)

		// MAX3 == S0.f, MAX3 == S1.f
		isMax0 := b.EmitFOrdEqual(typeBool, max3, val0)
		isMax1 := b.EmitFOrdEqual(typeBool, max3, val1)

		// D.f = MAX(S1.f, S2.f) if MAX3 == S0
		// D.f = MAX(S0.f, S2.f) if MAX3 == S1
		// Else D.f = MAX(S0.f, S1.f)
		max12 := b.EmitExtInst(typeFloat, idGlsl, spec.SpvGlslOpFMax, val1, val2)
		max02 := b.EmitExtInst(typeFloat, idGlsl, spec.SpvGlslOpFMax, val0, val2)
		max01_2 := b.EmitExtInst(typeFloat, idGlsl, spec.SpvGlslOpFMax, val0, val1)

		res := b.EmitSelect(typeFloat, isMax0, max12, b.EmitSelect(typeFloat, isMax1, max02, max01_2))

		// Final result based on anyNan
		finalRes := b.EmitSelect(typeFloat, anyNan, min3, res)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, finalRes, true)
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
		val = b.EmitExtInst(typeFloat, ctx.GetId(BlockContextIdGlsl), spec.SpvGlslOpFAbs, val)
	}
	if (neg>>i)&1 == 1 {
		val = b.EmitFNegate(typeFloat, val)
	}

	return val
}

func applyVop3OutputModifiers(b *SpvBuilder, ctx *SpirvBlockContext, val SpirvId, clamp bool, omod uint8) SpirvId {
	typeFloat := ctx.GetId(BlockContextIdTypeFloat)
	idGlsl := ctx.GetId(BlockContextIdGlsl)

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
