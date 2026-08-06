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
	case gcnSpec.Vop2OpSubI32:
		val0 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src1, 0, 1)

		resStruct := b.EmitISubBorrow(ctx.GetId(BlockContextIdTypeStructUintUint), val0, val1)
		resDiff := b.EmitCompositeExtract(ctx.GetId(BlockContextIdTypeUint), resStruct, 0)
		resBorrow := b.EmitCompositeExtract(ctx.GetId(BlockContextIdTypeUint), resStruct, 1)

		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resDiff, false)
		// AMD GCN V_SUB_I32 (carry-out is 1 if NO borrow). OpISubBorrow sets borrow to 1 if there IS a borrow.
		cond := b.EmitIEqual(ctx.GetId(BlockContextIdTypeBool), resBorrow, ctx.GetConstId(ConstIdUint0))
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
	case gcnSpec.Vop2OpMacLegacyF32:
		typeFloat := ctx.GetId(BlockContextIdTypeFloat)
		typeBool := ctx.GetId(BlockContextIdTypeBool)
		idZeroF := ctx.GetConstId(ConstIdFloat0)

		val0 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src1, 0, 1)
		valD := b.EmitBitcast(typeFloat, ctx.LoadRegisterPointer(b, details.Vdst+gcnSpec.OpVgpr0))

		isZero0 := b.EmitFOrdEqual(typeBool, val0, idZeroF)
		isZero1 := b.EmitFOrdEqual(typeBool, val1, idZeroF)
		anyZero := b.EmitLogicalOr(typeBool, isZero0, isZero1)

		mulF := b.EmitFMul(typeFloat, val0, val1)
		resMul := b.EmitSelect(typeFloat, anyZero, idZeroF, mulF)
		resF := b.EmitFAdd(typeFloat, resMul, valD)
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
		resF := b.EmitExtInst(ctx.GetId(BlockContextIdTypeFloat), ctx.GetId(BlockContextIdTypeGlsl), spec.SpvGlslOpFMin, val0, val1)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	case gcnSpec.Vop2OpMaxF32:
		val0 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src1, 0, 1)
		resF := b.EmitExtInst(ctx.GetId(BlockContextIdTypeFloat), ctx.GetId(BlockContextIdTypeGlsl), spec.SpvGlslOpFMax, val0, val1)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	case gcnSpec.Vop2OpMacF32:
		val0 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src1, 0, 1)
		valD := b.EmitBitcast(ctx.GetId(BlockContextIdTypeFloat), ctx.LoadRegisterPointer(b, details.Vdst+gcnSpec.OpVgpr0))
		resF := b.EmitExtInst(ctx.GetId(BlockContextIdTypeFloat), ctx.GetId(BlockContextIdTypeGlsl), spec.SpvGlslOpFma, val0, val1, valD)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	case gcnSpec.Vop2OpCvtPkrtzF16F32:
		val0 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src1, 0, 1)
		vec := b.EmitCompositeConstruct(ctx.GetId(BlockContextIdTypeV2Float), val0, val1)
		resU := b.EmitExtInst(ctx.GetId(BlockContextIdTypeUint), ctx.GetId(BlockContextIdTypeGlsl), spec.SpvGlslOpPackHalf2x16, vec)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resU, false)
	case gcnSpec.Vop2OpLshlrevB32:
		val0 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src1, 0, 1)
		shift := b.EmitBitwiseAnd(ctx.GetId(BlockContextIdTypeUint), val0, ctx.GetConstId(ConstIdUint31))
		resU := b.EmitShiftLeftLogical(ctx.GetId(BlockContextIdTypeUint), val1, shift)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resU, false)
	case gcnSpec.Vop2OpLshrrevB32:
		val0 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src1, 0, 1)
		shift := b.EmitBitwiseAnd(ctx.GetId(BlockContextIdTypeUint), val0, ctx.GetConstId(ConstIdUint31))
		resU := b.EmitShiftRightLogical(ctx.GetId(BlockContextIdTypeUint), val1, shift)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resU, false)
	case gcnSpec.Vop2OpAshrrevI32:
		val0 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandIntValueModified(b, ctx, details.Abs, details.Neg, details.Src1, 0, 1)
		shift := b.EmitBitwiseAnd(ctx.GetId(BlockContextIdTypeUint), val0, ctx.GetConstId(ConstIdUint31))
		resI := b.EmitShiftRightArithmetic(ctx.GetId(BlockContextIdTypeInt), val1, shift)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resI, false)
	case gcnSpec.Vop2OpBcntU32B32:
		typeUint := ctx.GetId(BlockContextIdTypeUint)

		val0 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src1, 0, 1)
		cnt := b.EmitBitCount(typeUint, val0)
		resU := b.EmitIAdd(typeUint, cnt, val1)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resU, false)
	case gcnSpec.Vop2OpMinU32:
		val0 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src1, 0, 1)
		resU := b.EmitExtInst(ctx.GetId(BlockContextIdTypeUint), ctx.GetId(BlockContextIdTypeGlsl), spec.SpvGlslOpUMin, val0, val1)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resU, false)
	case gcnSpec.Vop2OpAndB32:
		val0 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src1, 0, 1)
		resU := b.EmitBitwiseAnd(ctx.GetId(BlockContextIdTypeUint), val0, val1)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resU, false)
	case gcnSpec.Vop2OpCndmaskB32:
		typeUint := ctx.GetId(BlockContextIdTypeUint)
		typeBool := ctx.GetId(BlockContextIdTypeBool)
		idZeroU := ctx.GetConstId(ConstIdUint0)
		idOneU := ctx.GetConstId(ConstIdUint1)

		// D.u = VCC[i] ? S1.u : S0.u (i = threadID in wave); VOP3: specify VCC as a scalar GPR in S2.
		val0 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src1, 0, 1)
		var maskVal SpirvId
		if instr.Encoding == gcnSpec.EncVOP3 {
			maskVal = ctx.GetOperandValue(b, details.Src2, 0)
		} else {
			maskVal = ctx.GetOperandValue(b, gcnSpec.OpVccLo, 0)
		}

		// Get mask bit based on thread ID.
		threadId := b.EmitLoad(typeUint, ctx.GetId(BlockContextIdSubgroupLocalInvocationId))
		shiftedMask := b.EmitShiftRightLogical(typeUint, maskVal, threadId)
		bitVal := b.EmitBitwiseAnd(typeUint, shiftedMask, idOneU)

		// D.u = mask[i] ? S1.u : S0.u.
		cond := b.EmitINotEqual(typeBool, bitVal, idZeroU)
		resU := b.EmitSelect(typeUint, cond, val1, val0)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resU, false)
	case gcnSpec.Vop2OpMadmkF32:
		typeUint := ctx.GetId(BlockContextIdTypeUint)
		typeFloat := ctx.GetId(BlockContextIdTypeFloat)

		val0 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src1, instr.Literal, 1)
		valK := b.EmitBitcast(typeFloat, b.EmitConstantUint(typeUint, instr.Literal))

		// D.f = S0.f * K + S1.f
		resF := b.EmitExtInst(typeFloat, ctx.GetId(BlockContextIdTypeGlsl), spec.SpvGlslOpFma, val0, valK, val1)
		StoreRegisterPointerMaskedModified(b, ctx, details.Clamp, details.OMod, details.Vdst+gcnSpec.OpVgpr0, resF, true)
	default:
		panic(fmt.Sprintf("unknown vop2 op %s", gcnSpec.Mnemotics[gcnSpec.EncVOP2][details.Op]))
	}
}
