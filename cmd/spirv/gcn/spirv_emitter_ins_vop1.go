package gcn

import (
	"fmt"

	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
)

func EmitVOP1(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*gcnSpec.VectorDetails)
	switch details.Op {
	case gcnSpec.Vop1OpMovB32:
		val := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, val, false)
	case gcnSpec.Vop1OpCvtF32I32:
		valInt := GetOperandIntValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		resF := b.EmitConvertSToF(ctx.GetId(BlockContextIdTypeFloat), valInt)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	case gcnSpec.Vop1OpExpF32:
		val := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		resF := b.EmitExtInst(ctx.GetId(BlockContextIdTypeFloat), ctx.GetId(BlockContextIdGlsl), spec.SpvGlslOpExp2, val)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	case gcnSpec.Vop1OpFractF32:
		val := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		resF := b.EmitExtInst(ctx.GetId(BlockContextIdTypeFloat), ctx.GetId(BlockContextIdGlsl), spec.SpvGlslOpFract, val)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	case gcnSpec.Vop1OpSqrtF32:
		val := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		resF := b.EmitExtInst(ctx.GetId(BlockContextIdTypeFloat), ctx.GetId(BlockContextIdGlsl), spec.SpvGlslOpSqrt, val)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	case gcnSpec.Vop1OpRcpF32:
		val := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		resF := b.EmitFDiv(ctx.GetId(BlockContextIdTypeFloat), ctx.GetConstId(ConstIdFloat1), val)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	default:
		panic(fmt.Sprintf("unknown vop1 op %s", gcnSpec.Mnemotics[gcnSpec.EncVOP1][details.Op]))
	}
}
