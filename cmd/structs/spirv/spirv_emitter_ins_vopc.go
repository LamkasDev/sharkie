package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitVOPC(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	details := instr.Details.(*VectorDetails)
	idBool := ctx.GetId(BlockContextIdTypeBool)
	var resB uint32
	switch details.Op {
	case VopcOpCmpEqF32:
		val0 := ctx.GetOperandFloatValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandFloatValue(b, details.Src1+OpVgpr0, 0)
		resB = b.EmitFOrdEqual(idBool, val0, val1)
	case VopcOpCmpNeqF32:
		val0 := ctx.GetOperandFloatValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandFloatValue(b, details.Src1+OpVgpr0, 0)
		resB = b.EmitFUnordNotEqual(idBool, val0, val1)
	case VopcOpCmpGtU32:
		val0 := ctx.GetOperandUintValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandUintValue(b, details.Src1+OpVgpr0, 0)
		resB = b.EmitUGreaterThan(idBool, val0, val1)
	default:
		panic(fmt.Sprintf("unknown vopc op %s", Mnemotics[EncVOPC][details.Op]))
	}

	emitVCCUpdate(b, ctx, resB)
}

func emitVCCUpdate(b *SpvBuilder, ctx SpirvBlockContext, cond uint32) {
	idV4Uint := ctx.GetId(BlockContextIdTypeV4Uint)
	idUint := ctx.GetId(BlockContextIdTypeUint)
	idC3 := ctx.GetConstId(ConstIdxUint3)

	// Combine boolean results into VCC.
	ballot := b.EmitGroupNonUniformBallot(idV4Uint, idC3, cond)
	vccLo := b.EmitCompositeExtract(idUint, ballot, 0)
	vccHi := b.EmitCompositeExtract(idUint, ballot, 1)

	ctx.StoreRegisterPointer(b, OpVccLo, vccLo)
	ctx.StoreRegisterPointer(b, OpVccHi, vccHi)
}
