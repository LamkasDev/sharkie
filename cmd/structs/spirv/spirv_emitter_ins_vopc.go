package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitVOPC(b *SpvBuilder, instr *Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*VectorDetails)
	typeBool := ctx.GetId(BlockContextIdTypeBool)
	var resB uint32
	switch details.Op {
	case VopcOpCmpEqF32:
		val0 := ctx.GetOperandFloatValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandFloatValue(b, details.Src1+OpVgpr0, 0)
		resB = b.EmitFOrdEqual(typeBool, val0, val1)
	case VopcOpCmpNeqF32:
		val0 := ctx.GetOperandFloatValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandFloatValue(b, details.Src1+OpVgpr0, 0)
		resB = b.EmitFUnordNotEqual(typeBool, val0, val1)
	case VopcOpCmpGtU32:
		val0 := ctx.GetOperandUintValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandUintValue(b, details.Src1+OpVgpr0, 0)
		resB = b.EmitUGreaterThan(typeBool, val0, val1)
	default:
		panic(fmt.Sprintf("unknown vopc op %s", Mnemotics[EncVOPC][details.Op]))
	}

	emitVCCUpdate(b, ctx, resB)
}

func emitVCCUpdate(b *SpvBuilder, ctx *SpirvBlockContext, cond uint32) {
	typeV4Uint := ctx.GetId(BlockContextIdTypeV4Uint)
	typeUint := ctx.GetId(BlockContextIdTypeUint)
	idC3 := ctx.GetConstId(ConstIdxUint3)

	// Combine boolean results into VCC.
	ballot := b.EmitGroupNonUniformBallot(typeV4Uint, idC3, cond)
	vccLo := b.EmitCompositeExtract(typeUint, ballot, 0)
	vccHi := b.EmitCompositeExtract(typeUint, ballot, 1)

	ctx.StoreRegisterPointer(b, OpVccLo, vccLo)
	ctx.StoreRegisterPointer(b, OpVccHi, vccHi)
}
