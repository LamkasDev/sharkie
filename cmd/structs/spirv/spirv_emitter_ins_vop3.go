package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitVOP3(b *SpvBuilder, instr *Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*Vop3Details)
	switch {
	case details.Op <= 0xFF: // VOPC
		instr.Details = &VectorDetails{
			Src0: details.Src0,
			Src1: details.Src1,
			Op:   details.Op,
		}
		emitVOPC(b, instr, ctx)
	case details.Op >= Vop3OpCndmaskB32 && details.Op <= Vop3OpCvtPkI16I32: // VOP2
		instr.Details = &VectorDetails{
			Src0: details.Src0,
			Src1: details.Src1,
			Dst:  details.Dst,
			Op:   details.Op - Vop3OpCndmaskB32,
		}
		emitVOP2(b, instr, ctx)
	case details.Op >= Vop3OpNop && details.Op <= Vop3OpMovrelsdB32: // VOP1
		instr.Details = &VectorDetails{
			Src0: details.Src0,
			Dst:  details.Dst,
			Op:   details.Op - Vop3OpNop,
		}
		emitVOP1(b, instr, ctx)
	case details.Op == Vop3OpMadF32:
		val0 := ctx.GetOperandFloatValue(b, details.Src0, 0)
		val1 := ctx.GetOperandFloatValue(b, details.Src1, 0)
		val2 := ctx.GetOperandFloatValue(b, details.Src2, 0)
		resF := b.EmitExtInst(ctx.GetId(BlockContextIdTypeFloat), ctx.GetId(BlockContextIdGlsl), SpvGlslOpFma, val0, val1, val2)
		ctx.StoreRegisterPointer(b, details.Dst+OpVgpr0, b.EmitBitcast(ctx.GetId(BlockContextIdTypeUint), resF))
	case details.Op == Vop3OpMed3F32:
		// TODO: add SPV_AMD_shader_trinary_minmax optimized version.
		typeFloat := ctx.GetId(BlockContextIdTypeFloat)
		typeBool := ctx.GetId(BlockContextIdTypeBool)
		idGlsl := ctx.GetId(BlockContextIdGlsl)

		val0 := ctx.GetOperandFloatValue(b, details.Src0, 0)
		val1 := ctx.GetOperandFloatValue(b, details.Src1, 0)
		val2 := ctx.GetOperandFloatValue(b, details.Src2, 0)

		// isNan(S0.f) || isNan(S1.f) || isNan(S2.f)
		nan0 := b.EmitIsNan(typeBool, val0)
		nan1 := b.EmitIsNan(typeBool, val1)
		nan2 := b.EmitIsNan(typeBool, val2)
		anyNan := b.EmitLogicalOr(typeBool, nan0, b.EmitLogicalOr(typeBool, nan1, nan2))

		// MIN3(S0.f, S1.f, S2.f)
		min01 := b.EmitExtInst(typeFloat, idGlsl, SpvGlslOpFMin, val0, val1)
		min3 := b.EmitExtInst(typeFloat, idGlsl, SpvGlslOpFMin, min01, val2)

		// MAX3(S0.f, S1.f, S2.f)
		max01 := b.EmitExtInst(typeFloat, idGlsl, SpvGlslOpFMax, val0, val1)
		max3 := b.EmitExtInst(typeFloat, idGlsl, SpvGlslOpFMax, max01, val2)

		// MAX3 == S0.f, MAX3 == S1.f
		isMax0 := b.EmitFOrdEqual(typeBool, max3, val0)
		isMax1 := b.EmitFOrdEqual(typeBool, max3, val1)

		// D.f = MAX(S1.f, S2.f) if MAX3 == S0
		// D.f = MAX(S0.f, S2.f) if MAX3 == S1
		// Else D.f = MAX(S0.f, S1.f)
		max12 := b.EmitExtInst(typeFloat, idGlsl, SpvGlslOpFMax, val1, val2)
		max02 := b.EmitExtInst(typeFloat, idGlsl, SpvGlslOpFMax, val0, val2)
		max01_2 := b.EmitExtInst(typeFloat, idGlsl, SpvGlslOpFMax, val0, val1)

		res := b.EmitSelect(typeFloat, isMax0, max12, b.EmitSelect(typeFloat, isMax1, max02, max01_2))

		// Final result based on anyNan
		finalRes := b.EmitSelect(typeFloat, anyNan, min3, res)
		ctx.StoreRegisterPointer(b, details.Dst+OpVgpr0, b.EmitBitcast(ctx.GetId(BlockContextIdTypeUint), finalRes))
	default:
		panic(fmt.Sprintf("unknown vop3 op %s", Mnemotics[EncVOP3][details.Op]))
	}
}
