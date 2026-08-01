package gcn

import (
	"fmt"

	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
)

func EmitVOPC(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*gcnSpec.VectorDetails)
	typeBool := ctx.GetId(BlockContextIdTypeBool)
	var cond SpirvId
	switch details.Op {
	case gcnSpec.VopcOpCmpEqF32:
		val0 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src1, 0, 1)
		cond = b.EmitFOrdEqual(typeBool, val0, val1)
	case gcnSpec.VopcOpCmpNeqF32:
		val0 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src1, 0, 1)
		cond = b.EmitFUnordNotEqual(typeBool, val0, val1)
	case gcnSpec.VopcOpCmpLtF32:
		val0 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src1, 0, 1)
		cond = b.EmitFOrdLessThan(typeBool, val0, val1)
	case gcnSpec.VopcOpCmpGtF32:
		val0 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src1, 0, 1)
		cond = b.EmitFOrdGreaterThan(typeBool, val0, val1)
	case gcnSpec.VopcOpCmpGeF32:
		val0 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandFloatValueModified(b, ctx, details.Abs, details.Neg, details.Src1, 0, 1)
		cond = b.EmitFOrdGreaterThanEqual(typeBool, val0, val1)
	case gcnSpec.VopcOpCmpGtU32, gcnSpec.VopcOpCmpxGtU32:
		val0 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandUintValueModified(b, ctx, details.Abs, details.Neg, details.Src1, 0, 1)
		cond = b.EmitUGreaterThan(typeBool, val0, val1)
	case gcnSpec.VopcOpCmpLgI32:
		val0 := GetOperandIntValueModified(b, ctx, details.Abs, details.Neg, details.Src0, instr.Literal, 0)
		val1 := GetOperandIntValueModified(b, ctx, details.Abs, details.Neg, details.Src1, 0, 1)
		cond = b.EmitINotEqual(typeBool, val0, val1)
	default:
		panic(fmt.Sprintf("unknown vopc op %s", gcnSpec.Mnemotics[gcnSpec.EncVOPC][details.Op]))
	}

	emitComparisonResultUpdate(b, ctx, cond, details.Sdst)

	// V_CMPX updates EXEC as well.
	if details.IsVopcCmpx() {
		resLo := ctx.GetGcnSgprId(b, details.Sdst)
		resHi := ctx.GetGcnSgprId(b, details.Sdst+1)
		ctx.StoreRegisterPointer(b, gcnSpec.OpExecLo, resLo)
		ctx.StoreRegisterPointer(b, gcnSpec.OpExecHi, resHi)
	}
}

func emitComparisonResultUpdate(b *SpvBuilder, ctx *SpirvBlockContext, cond SpirvId, dst uint32) {
	typeV4Uint := ctx.GetId(BlockContextIdTypeV4Uint)
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	idC3 := ctx.GetConstId(ConstIdUint3)

	// Combine boolean results into comparison result.
	ballot := b.EmitGroupNonUniformBallot(typeV4Uint, idC3, cond)
	rawResLo := b.EmitCompositeExtract(typeUint, ballot, 0)
	rawResHi := b.EmitCompositeExtract(typeUint, ballot, 1)

	// Mask out the inactive lanes.
	execLo, execHi := ctx.GetOperand64Value(b, gcnSpec.OpExecLo, 0)
	resLo := b.EmitBitwiseAnd(typeUint, rawResLo, execLo)
	resHi := b.EmitBitwiseAnd(typeUint, rawResHi, execHi)

	ctx.StoreRegisterPointer(b, dst, resLo)
	ctx.StoreRegisterPointer(b, dst+1, resHi)
}
