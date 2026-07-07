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
		if details.Abs != 0 || details.Neg != 0 || details.OMod != 0 || details.Clamp {
			panic(fmt.Sprintf("vop3 modifiers not implemented for %s", gcnSpec.Mnemotics[gcnSpec.EncVOP3][details.Op]))
		}
		instr.Details = &gcnSpec.VectorDetails{
			Op:   details.Op,
			Vdst: details.Vdst,
			Src0: details.Src0,
			Src1: details.Src1,
			Sdst: details.Sdst,
		}
		EmitVOPC(b, instr, ctx)
	case details.Op >= gcnSpec.Vop3OpCndmaskB32 && details.Op <= gcnSpec.Vop3OpCvtPkI16I32: // VOP2
		if details.Abs != 0 || details.Neg != 0 || details.OMod != 0 || details.Clamp {
			panic(fmt.Sprintf("vop3 modifiers not implemented for %s", gcnSpec.Mnemotics[gcnSpec.EncVOP3][details.Op]))
		}
		instr.Details = &gcnSpec.VectorDetails{
			Op:   details.Op - gcnSpec.Vop3OpCndmaskB32,
			Vdst: details.Vdst,
			Src0: details.Src0,
			Src1: details.Src1,
			Sdst: details.Sdst,
		}
		EmitVOP2(b, instr, ctx)
	case details.Op >= gcnSpec.Vop3OpNop && details.Op <= gcnSpec.Vop3OpMovrelsdB32: // VOP1
		if details.Abs != 0 || details.Neg != 0 || details.OMod != 0 || details.Clamp {
			panic(fmt.Sprintf("vop3 modifiers not implemented for %s", gcnSpec.Mnemotics[gcnSpec.EncVOP3][details.Op]))
		}
		instr.Details = &gcnSpec.VectorDetails{
			Op:   details.Op - gcnSpec.Vop3OpNop,
			Vdst: details.Vdst,
			Src0: details.Src0,
			Sdst: details.Sdst,
		}
		EmitVOP1(b, instr, ctx)
	case details.Op == gcnSpec.Vop3OpMadF32:
		val0 := applyVop3Modifiers(b, ctx, ctx.GetOperandFloatValue(b, details.Src0, instr.Literal), details, 0)
		val1 := applyVop3Modifiers(b, ctx, ctx.GetOperandFloatValue(b, details.Src1, instr.Literal), details, 1)
		val2 := applyVop3Modifiers(b, ctx, ctx.GetOperandFloatValue(b, details.Src2, instr.Literal), details, 2)

		resF := b.EmitExtInst(ctx.GetId(BlockContextIdTypeFloat), ctx.GetId(BlockContextIdGlsl), spec.SpvGlslOpFma, val0, val1, val2)
		resF = applyVop3OutputModifiers(b, ctx, resF, details.Clamp, details.OMod)
		ctx.StoreRegisterPointerMasked(b, details.Vdst+gcnSpec.OpVgpr0, b.EmitBitcast(ctx.GetId(BlockContextIdTypeUint), resF))
	case details.Op == gcnSpec.Vop3OpMed3F32:
		// TODO: add SPV_AMD_shader_trinary_minmax optimized version.
		typeFloat := ctx.GetId(BlockContextIdTypeFloat)
		typeBool := ctx.GetId(BlockContextIdTypeBool)
		idGlsl := ctx.GetId(BlockContextIdGlsl)

		val0 := applyVop3Modifiers(b, ctx, ctx.GetOperandFloatValue(b, details.Src0, instr.Literal), details, 0)
		val1 := applyVop3Modifiers(b, ctx, ctx.GetOperandFloatValue(b, details.Src1, instr.Literal), details, 1)
		val2 := applyVop3Modifiers(b, ctx, ctx.GetOperandFloatValue(b, details.Src2, instr.Literal), details, 2)

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
		finalRes = applyVop3OutputModifiers(b, ctx, finalRes, details.Clamp, details.OMod)
		ctx.StoreRegisterPointerMasked(b, details.Vdst+gcnSpec.OpVgpr0, b.EmitBitcast(ctx.GetId(BlockContextIdTypeUint), finalRes))
	default:
		panic(fmt.Sprintf("unknown vop3 op %s", gcnSpec.Mnemotics[gcnSpec.EncVOP3][details.Op]))
	}
}

func applyVop3Modifiers(b *SpvBuilder, ctx *SpirvBlockContext, val SpirvId, details *gcnSpec.Vop3Details, i int) SpirvId {
	typeFloat := ctx.GetId(BlockContextIdTypeFloat)
	if (details.Abs>>i)&1 == 1 {
		val = b.EmitExtInst(typeFloat, ctx.GetId(BlockContextIdGlsl), spec.SpvGlslOpFAbs, val)
	}
	if (details.Neg>>i)&1 == 1 {
		val = b.EmitFNegate(typeFloat, val)
	}

	return val
}

func applyVop3OutputModifiers(b *SpvBuilder, ctx *SpirvBlockContext, val SpirvId, clamp bool, omod uint8) SpirvId {
	typeFloat := ctx.GetId(BlockContextIdTypeFloat)
	idGlsl := ctx.GetId(BlockContextIdGlsl)

	// TODO: use built-ins.
	// Apply omod.
	switch omod {
	case 1:
		idC2 := b.EmitConstantFloat(typeFloat, 2.0)
		val = b.EmitFMul(typeFloat, val, idC2)
	case 2:
		idC4 := b.EmitConstantFloat(typeFloat, 4.0)
		val = b.EmitFMul(typeFloat, val, idC4)
	case 3:
		idC05 := b.EmitConstantFloat(typeFloat, 0.5)
		val = b.EmitFMul(typeFloat, val, idC05)
	}

	// TODO: use built-ins.
	// Apply clamp.
	if clamp {
		idC0 := b.EmitConstantFloat(typeFloat, 0.0)
		idC1 := b.EmitConstantFloat(typeFloat, 1.0)
		val = b.EmitExtInst(typeFloat, idGlsl, spec.SpvGlslOpFClamp, val, idC0, idC1)
	}

	return val
}
