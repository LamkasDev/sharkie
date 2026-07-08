package gcn

import (
	"fmt"

	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
)

func EmitVOP2(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*gcnSpec.VectorDetails)
	switch details.Op {
	case gcnSpec.Vop2OpAddF32:
		val0 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src1, 0, 1)
		resF := b.EmitFAdd(ctx.GetId(BlockContextIdTypeFloat), val0, val1)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	case gcnSpec.Vop2OpAddI32:
		val0 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src1, 0, 1)

		resStruct := b.EmitIAddCarry(ctx.GetId(BlockContextIdTypeStructUintUint), val0, val1)
		resSum := b.EmitCompositeExtract(ctx.GetId(BlockContextIdTypeUint), resStruct, 0)
		resCarry := b.EmitCompositeExtract(ctx.GetId(BlockContextIdTypeUint), resStruct, 1)

		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resSum, false)
		cond := b.EmitINotEqual(ctx.GetId(BlockContextIdTypeBool), resCarry, ctx.GetConstId(ConstIdUint0))
		emitComparisonResultUpdate(b, ctx, cond, details.Sdst)
	case gcnSpec.Vop2OpSubF32:
		val0 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src1, 0, 1)
		resF := b.EmitFSub(ctx.GetId(BlockContextIdTypeFloat), val0, val1)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	case gcnSpec.Vop2OpSubrevF32:
		val0 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src1, 0, 1)
		resF := b.EmitFSub(ctx.GetId(BlockContextIdTypeFloat), val1, val0)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	case gcnSpec.Vop2OpMulF32:
		val0 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src1, 0, 1)
		resF := b.EmitFMul(ctx.GetId(BlockContextIdTypeFloat), val0, val1)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	case gcnSpec.Vop2OpMulLegacyF32:
		typeFloat := ctx.GetId(BlockContextIdTypeFloat)
		typeBool := ctx.GetId(BlockContextIdTypeBool)
		idZeroF := ctx.GetConstId(ConstIdFloat0)

		val0 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src1, 0, 1)

		isZero0 := b.EmitFOrdEqual(typeBool, val0, idZeroF)
		isZero1 := b.EmitFOrdEqual(typeBool, val1, idZeroF)
		anyZero := b.EmitLogicalOr(typeBool, isZero0, isZero1)

		mulF := b.EmitFMul(typeFloat, val0, val1)
		resF := b.EmitSelect(typeFloat, anyZero, idZeroF, mulF)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	case gcnSpec.Vop2OpMinF32:
		val0 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src1, 0, 1)
		resF := b.EmitExtInst(ctx.GetId(BlockContextIdTypeFloat), ctx.GetId(BlockContextIdGlsl), spec.SpvGlslOpFMin, val0, val1)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	case gcnSpec.Vop2OpMaxF32:
		val0 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src1, 0, 1)
		resF := b.EmitExtInst(ctx.GetId(BlockContextIdTypeFloat), ctx.GetId(BlockContextIdGlsl), spec.SpvGlslOpFMax, val0, val1)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	case gcnSpec.Vop2OpMacF32:
		val0 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src1, 0, 1)
		valD := b.EmitBitcast(ctx.GetId(BlockContextIdTypeFloat), ctx.LoadRegisterPointer(b, details.Vdst+gcnSpec.OpVgpr0))
		valD = applyVop3Modifiers(b, ctx, valD, details.Abs, details.Neg, 2)
		resF := b.EmitExtInst(ctx.GetId(BlockContextIdTypeFloat), ctx.GetId(BlockContextIdGlsl), spec.SpvGlslOpFma, val0, val1, valD)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	case gcnSpec.Vop2OpCvtPkrtzF16F32:
		val0 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src1, 0, 1)
		vec := b.EmitCompositeConstruct(ctx.GetId(BlockContextIdTypeV2Float), val0, val1)
		resU := b.EmitExtInst(ctx.GetId(BlockContextIdTypeUint), ctx.GetId(BlockContextIdGlsl), spec.SpvGlslOpPackHalf2x16, vec)
		ctx.StoreRegisterPointerMasked(b, details.Vdst+gcnSpec.OpVgpr0, resU)
	case gcnSpec.Vop2OpLshlrevB32:
		val0 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src1, 0, 1)
		shift := b.EmitBitwiseAnd(ctx.GetId(BlockContextIdTypeUint), val0, ctx.GetConstId(ConstIdUint31))
		resU := b.EmitShiftLeftLogical(ctx.GetId(BlockContextIdTypeUint), val1, shift)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resU, false)
	case gcnSpec.Vop2OpMinU32:
		val0 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src1, 0, 1)
		resU := b.EmitExtInst(ctx.GetId(BlockContextIdTypeUint), ctx.GetId(BlockContextIdGlsl), spec.SpvGlslOpUMin, val0, val1)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resU, false)
	case gcnSpec.Vop2OpAndB32:
		val0 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src1, 0, 1)
		resU := b.EmitBitwiseAnd(ctx.GetId(BlockContextIdTypeUint), val0, val1)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resU, false)
	default:
		panic(fmt.Sprintf("unknown vop2 op %s", gcnSpec.Mnemotics[gcnSpec.EncVOP2][details.Op]))
	}
}
