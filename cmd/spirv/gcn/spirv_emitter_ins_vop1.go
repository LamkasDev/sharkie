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
		resF := b.EmitExtInst(ctx.GetId(BlockContextIdTypeFloat), ctx.GetId(BlockContextIdTypeGlsl), spec.SpvGlslOpExp2, val)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	case gcnSpec.Vop1OpFractF32:
		val := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		resF := b.EmitExtInst(ctx.GetId(BlockContextIdTypeFloat), ctx.GetId(BlockContextIdTypeGlsl), spec.SpvGlslOpFract, val)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	case gcnSpec.Vop1OpSqrtF32:
		val := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		resF := b.EmitExtInst(ctx.GetId(BlockContextIdTypeFloat), ctx.GetId(BlockContextIdTypeGlsl), spec.SpvGlslOpSqrt, val)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	case gcnSpec.Vop1OpRcpF32:
		val := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		resF := b.EmitFDiv(ctx.GetId(BlockContextIdTypeFloat), ctx.GetConstId(ConstIdFloat1), val)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	case gcnSpec.Vop1OpRsqClampF32:
		val := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		resF := b.EmitExtInst(ctx.GetId(BlockContextIdTypeFloat), ctx.GetId(BlockContextIdTypeGlsl), spec.SpvGlslOpInverseSqrt, val)
		resF = b.EmitExtInst(ctx.GetId(BlockContextIdTypeFloat), ctx.GetId(BlockContextIdTypeGlsl), spec.SpvGlslOpFClamp, resF, ctx.GetConstId(ConstIdFloatMin), ctx.GetConstId(ConstIdFloatMax))
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	case gcnSpec.Vop1OpRsqF32:
		val := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		resF := b.EmitExtInst(ctx.GetId(BlockContextIdTypeFloat), ctx.GetId(BlockContextIdTypeGlsl), spec.SpvGlslOpInverseSqrt, val)
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
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, res, false)
	case gcnSpec.Vop1OpCvtF16F32:
		typeVec2 := ctx.GetId(BlockContextIdTypeV2Float)

		valF := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		vec2 := b.EmitCompositeConstruct(typeVec2, valF, ctx.GetConstId(ConstIdFloat0))
		resU := b.EmitExtInst(ctx.GetId(BlockContextIdTypeUint), ctx.GetId(BlockContextIdTypeGlsl), spec.SpvGlslOpPackHalf2x16, vec2)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resU, false)
	case gcnSpec.Vop1OpCvtF32F16:
		typeVec2 := ctx.GetId(BlockContextIdTypeV2Float)

		valUint := ctx.GetOperandUintValue(b, details.Src0, instr.Literal)
		vec2 := b.EmitExtInst(typeVec2, ctx.GetId(BlockContextIdTypeGlsl), spec.SpvGlslOpUnpackHalf2x16, valUint)
		resF := b.EmitCompositeExtract(ctx.GetId(BlockContextIdTypeFloat), vec2, 0)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	case gcnSpec.Vop1OpCvtOffF32I4:
		typeUint := ctx.GetId(BlockContextIdTypeUint)
		typeInt := ctx.GetId(BlockContextIdTypeInt)
		typeFloat := ctx.GetId(BlockContextIdTypeFloat)

		// Fetch the source operand as a 32-bit integer.
		valInt := GetOperandIntValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)

		// Extract the lower 4 bits and sign-extend to a 32-bit integer.
		offset := b.EmitConstantUint(typeUint, 0)
		count := b.EmitConstantUint(typeUint, 4)
		signExtInt := b.EmitBitFieldSExtract(typeInt, valInt, offset, count)

		// Convert the sign-extended 32-bit integer to a 32-bit float.
		valF := b.EmitConvertSToF(typeFloat, signExtInt)

		// Multiply by 0.0625f (which is 1/16).
		// 0x3d800000 is the IEEE-754 binary representation of 0.0625f.
		factor := b.EmitBitcast(typeFloat, b.EmitConstantUint(typeUint, 0x3d800000))
		resF := b.EmitFMul(typeFloat, valF, factor)

		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	case gcnSpec.Vop1OpCvtF32U32:
		valUint := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		resF := b.EmitConvertUToF(ctx.GetId(BlockContextIdTypeFloat), valUint)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	case gcnSpec.Vop1OpCvtF32Ubyte0:
		typeUint := ctx.GetId(BlockContextIdTypeUint)
		valUint := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)

		// Extract byte 0 (S0.u[7:0]) using a bitwise AND mask.
		mask := b.EmitConstantUint(typeUint, 0xFF)
		byte0 := b.EmitBitwiseAnd(typeUint, valUint, mask)

		resF := b.EmitConvertUToF(ctx.GetId(BlockContextIdTypeFloat), byte0)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	case gcnSpec.Vop1OpCvtF32Ubyte3:
		typeUint := ctx.GetId(BlockContextIdTypeUint)
		valUint := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)

		// Extract byte 3 (S0.u[31:24]) using a logical right shift.
		shiftAmount := b.EmitConstantUint(typeUint, 24)
		byte3 := b.EmitShiftRightLogical(typeUint, valUint, shiftAmount)

		resF := b.EmitConvertUToF(ctx.GetId(BlockContextIdTypeFloat), byte3)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	case gcnSpec.Vop1OpFloorF32:
		valF := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		resF := b.EmitExtInst(ctx.GetId(BlockContextIdTypeFloat), ctx.GetId(BlockContextIdTypeGlsl), spec.SpvGlslOpFloor, valF)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	case gcnSpec.Vop1OpRndneF32:
		valF := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		resF := b.EmitExtInst(ctx.GetId(BlockContextIdTypeFloat), ctx.GetId(BlockContextIdTypeGlsl), spec.SpvGlslOpRoundEven, valF)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	case gcnSpec.Vop1OpLogF32:
		val := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		resF := b.EmitExtInst(ctx.GetId(BlockContextIdTypeFloat), ctx.GetId(BlockContextIdTypeGlsl), spec.SpvGlslOpLog2, val)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	case gcnSpec.Vop1OpSinF32:
		typeFloat := ctx.GetId(BlockContextIdTypeFloat)
		typeBool := ctx.GetId(BlockContextIdTypeBool)
		typeUint := ctx.GetId(BlockContextIdTypeUint)

		valF := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)

		// Convert normalized input to radians (valF * 2PI).
		// 0x40C90FDB is the IEEE-754 binary representation of 2*PI (6.2831855).
		twoPi := b.EmitBitcast(typeFloat, b.EmitConstantUint(typeUint, 0x40C90FDB))
		radians := b.EmitFMul(typeFloat, valF, twoPi)

		// Compute the sine value.
		sinVal := b.EmitExtInst(typeFloat, ctx.GetId(BlockContextIdTypeGlsl), spec.SpvGlslOpSin, radians)

		// Bounds check (valid domain is [-256.0, +256.0]).
		// We can check this efficiently by taking the absolute value: abs(valF) <= 256.0.
		absVal := b.EmitExtInst(typeFloat, ctx.GetId(BlockContextIdTypeGlsl), spec.SpvGlslOpFAbs, valF)

		// 0x43800000 is the IEEE-754 binary representation of 256.0.
		limit := b.EmitBitcast(typeFloat, b.EmitConstantUint(typeUint, 0x43800000))
		inBounds := b.EmitFOrdLessThanEqual(typeBool, absVal, limit)

		// Select the computed sine if in bounds, otherwise 0.0.
		resF := b.EmitSelect(typeFloat, inBounds, sinVal, ctx.GetConstId(ConstIdFloat0))

		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	default:
		panic(fmt.Sprintf("unknown vop1 op %s", gcnSpec.Mnemotics[gcnSpec.EncVOP1][details.Op]))
	}
}
