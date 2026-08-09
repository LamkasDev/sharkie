package gcn

import (
	"fmt"

	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
)

func EmitSOPC(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*gcnSpec.ScalarDetails)
	switch details.Op {
	// Unsigned versions.
	case gcnSpec.SopcOpCmpEqU32:
		val0 := ctx.GetOperandUintValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandUintValue(b, details.Src1, instr.Literal)

		isEqual := b.EmitIEqual(ctx.GetId(BlockContextIdTypeBool), val0, val1)
		resScc := b.EmitSelect(ctx.GetId(BlockContextIdTypeUint), isEqual, ctx.GetConstId(ConstIdUint1), ctx.GetConstId(ConstIdUint0))
		ctx.StoreRegisterPointer(b, gcnSpec.OpScc, resScc)
	case gcnSpec.SopcOpCmpLgU32:
		val0 := ctx.GetOperandUintValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandUintValue(b, details.Src1, instr.Literal)

		isEqual := b.EmitINotEqual(ctx.GetId(BlockContextIdTypeBool), val0, val1)
		resScc := b.EmitSelect(ctx.GetId(BlockContextIdTypeUint), isEqual, ctx.GetConstId(ConstIdUint1), ctx.GetConstId(ConstIdUint0))
		ctx.StoreRegisterPointer(b, gcnSpec.OpScc, resScc)
	case gcnSpec.SopcOpCmpLeU32:
		val0 := ctx.GetOperandUintValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandUintValue(b, details.Src1, instr.Literal)

		isEqual := b.EmitULessThanEqual(ctx.GetId(BlockContextIdTypeBool), val0, val1)
		resScc := b.EmitSelect(ctx.GetId(BlockContextIdTypeUint), isEqual, ctx.GetConstId(ConstIdUint1), ctx.GetConstId(ConstIdUint0))
		ctx.StoreRegisterPointer(b, gcnSpec.OpScc, resScc)
	case gcnSpec.SopcOpCmpLtU32:
		val0 := ctx.GetOperandUintValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandUintValue(b, details.Src1, instr.Literal)

		isEqual := b.EmitULessThan(ctx.GetId(BlockContextIdTypeBool), val0, val1)
		resScc := b.EmitSelect(ctx.GetId(BlockContextIdTypeUint), isEqual, ctx.GetConstId(ConstIdUint1), ctx.GetConstId(ConstIdUint0))
		ctx.StoreRegisterPointer(b, gcnSpec.OpScc, resScc)
	// Signed versions.
	case gcnSpec.SopcOpCmpEqI32:
		val0 := ctx.GetOperandIntValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandIntValue(b, details.Src1, instr.Literal)

		isEqual := b.EmitIEqual(ctx.GetId(BlockContextIdTypeBool), val0, val1)
		resScc := b.EmitSelect(ctx.GetId(BlockContextIdTypeUint), isEqual, ctx.GetConstId(ConstIdUint1), ctx.GetConstId(ConstIdUint0))
		ctx.StoreRegisterPointer(b, gcnSpec.OpScc, resScc)
	case gcnSpec.SopcOpCmpLeI32:
		val0 := ctx.GetOperandIntValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandIntValue(b, details.Src1, instr.Literal)

		isLessEqual := b.EmitSLessThanEqual(ctx.GetId(BlockContextIdTypeBool), val0, val1)
		resScc := b.EmitSelect(ctx.GetId(BlockContextIdTypeUint), isLessEqual, ctx.GetConstId(ConstIdUint1), ctx.GetConstId(ConstIdUint0))
		ctx.StoreRegisterPointer(b, gcnSpec.OpScc, resScc)
	default:
		panic(fmt.Sprintf("unknown sopc op %s", gcnSpec.Mnemotics[gcnSpec.EncSOPC][details.Op]))
	}
}
