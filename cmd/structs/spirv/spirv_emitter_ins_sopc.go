package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitSOPC(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	details := instr.Details.(*ScalarDetails)
	switch details.Op {
	case SopcOpCmpEqU32:
		val0 := ctx.GetOperandUintValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandUintValue(b, details.Src1, instr.Literal)

		isEqual := b.EmitIEqual(ctx.GetId(BlockContextIdTypeBool), val0, val1)
		resScc := b.EmitSelect(ctx.GetId(BlockContextIdTypeUint), isEqual, ctx.GetConstId(ConstIdxUint1), ctx.GetConstId(ConstIdxUint0))
		ctx.StoreRegisterPointer(b, OpScc, resScc)
	default:
		panic(fmt.Sprintf("unknown sopc op %s", Mnemotics[EncSOPC][details.Op]))
	}
}
