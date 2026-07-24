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
	case gcnSpec.Vop1OpRsqClampF32:
		val := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		resF := b.EmitExtInst(ctx.GetId(BlockContextIdTypeFloat), ctx.GetId(BlockContextIdGlsl), spec.SpvGlslOpInverseSqrt, val)
		resF = b.EmitExtInst(ctx.GetId(BlockContextIdTypeFloat), ctx.GetId(BlockContextIdGlsl), spec.SpvGlslOpFClamp, resF, ctx.GetConstId(ConstIdFloatMin), ctx.GetConstId(ConstIdFloatMax))
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	case gcnSpec.Vop1OpCvtI32F32:
		typeUint := ctx.GetId(BlockContextIdTypeUint)
		typeInt := ctx.GetId(BlockContextIdTypeInt)
		typeFloat := ctx.GetId(BlockContextIdTypeFloat)
		typeBool := ctx.GetId(BlockContextIdTypeBool)

		// 2147483648.0 (0x4f000000). Any float >= this overflows positive int32.
		upperBound := b.EmitBitcast(typeFloat, b.EmitConstantUint(typeUint, 0x4f000000))
		// -2147483648.0 (0xcf000000). Any float <= this underflows negative int32.
		lowerBound := b.EmitBitcast(typeFloat, b.EmitConstantUint(typeUint, 0xcf000000))

		// Evaluate conditions.
		valF := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		isNan := b.EmitIsNan(typeBool, valF)
		isUpper := b.EmitFOrdGreaterThanEqual(typeBool, valF, upperBound)
		isLower := b.EmitFOrdLessThanEqual(typeBool, valF, lowerBound)

		// Base conversion (will be undefined for out-of-bounds, we mask it next).
		baseInt := b.EmitConvertFToS(typeInt, valF)
		baseUint := b.EmitBitcast(typeUint, baseInt)

		// Apply GCN saturation and NaN rules.
		res := b.EmitSelect(typeUint, isLower, b.EmitConstantUint(typeUint, 0x80000000), baseUint) // -max_int
		res = b.EmitSelect(typeUint, isUpper, b.EmitConstantUint(typeUint, 0x7FFFFFFF), res)       // max_int
		res = b.EmitSelect(typeUint, isNan, ctx.GetConstId(ConstIdUint0), res)                     // NaN -> 0

		// Store back as uint.
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, res, false)
	default:
		panic(fmt.Sprintf("unknown vop1 op %s", gcnSpec.Mnemotics[gcnSpec.EncVOP1][details.Op]))
	}
}
