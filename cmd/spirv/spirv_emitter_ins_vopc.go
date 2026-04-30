package spirv

import (
	"fmt"

	gcnSpec "github.com/LamkasDev/sharkie/cmd/structs/gcn/spec"
)

func emitVOPC(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*gcnSpec.VectorDetails)
	typeBool := ctx.GetId(BlockContextIdTypeBool)
	var resCond SpirvId
	switch details.Op {
	case gcnSpec.VopcOpCmpEqF32:
		val0 := ctx.GetOperandFloatValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandFloatValue(b, details.Src1+gcnSpec.OpVgpr0, 0)
		resCond = b.EmitFOrdEqual(typeBool, val0, val1)
	case gcnSpec.VopcOpCmpNeqF32:
		val0 := ctx.GetOperandFloatValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandFloatValue(b, details.Src1+gcnSpec.OpVgpr0, 0)
		resCond = b.EmitFUnordNotEqual(typeBool, val0, val1)
	case gcnSpec.VopcOpCmpGtU32:
		val0 := ctx.GetOperandUintValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandUintValue(b, details.Src1+gcnSpec.OpVgpr0, 0)
		resCond = b.EmitUGreaterThan(typeBool, val0, val1)
	default:
		panic(fmt.Sprintf("unknown vopc op %s", gcnSpec.Mnemotics[gcnSpec.EncVOPC][details.Op]))
	}

	emitVccUpdate(b, ctx, resCond)
}

func emitVccUpdate(b *SpvBuilder, ctx *SpirvBlockContext, cond SpirvId) {
	typeV4Uint := ctx.GetId(BlockContextIdTypeV4Uint)
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	idC3 := ctx.GetConstId(ConstIdUint3)

	// Combine boolean results into VCC.
	ballot := b.EmitGroupNonUniformBallot(typeV4Uint, idC3, cond)
	rawVccLo := b.EmitCompositeExtract(typeUint, ballot, 0)
	rawVccHi := b.EmitCompositeExtract(typeUint, ballot, 1)

	// Mask out the inactive lanes.
	execLo, execHi := ctx.GetOperand64Value(b, gcnSpec.OpExecLo, 0)
	vccLo := b.EmitBitwiseAnd(typeUint, rawVccLo, execLo)
	vccHi := b.EmitBitwiseAnd(typeUint, rawVccHi, execHi)

	ctx.StoreRegisterPointer(b, gcnSpec.OpVccLo, vccLo)
	ctx.StoreRegisterPointer(b, gcnSpec.OpVccHi, vccHi)
}
