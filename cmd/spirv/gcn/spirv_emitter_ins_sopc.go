package gcn

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	gcnSpec "github.com/LamkasDev/sharkie/cmd/structs/gcn/spec"
)

func EmitSOPC(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*gcnSpec.ScalarDetails)
	switch details.Op {
	case gcnSpec.SopcOpCmpEqU32:
		val0 := ctx.GetOperandUintValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandUintValue(b, details.Src1, instr.Literal)

		isEqual := b.EmitIEqual(ctx.GetId(BlockContextIdTypeBool), val0, val1)
		resScc := b.EmitSelect(ctx.GetId(BlockContextIdTypeUint), isEqual, ctx.GetConstId(ConstIdUint1), ctx.GetConstId(ConstIdUint0))
		ctx.StoreRegisterPointer(b, gcnSpec.OpScc, resScc)
	default:
		panic(fmt.Sprintf("unknown sopc op %s", gcnSpec.Mnemotics[gcnSpec.EncSOPC][details.Op]))
	}
}
