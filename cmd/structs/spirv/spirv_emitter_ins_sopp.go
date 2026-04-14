package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitSOPP(b *SpvBuilder, instr *Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*ScalarDetails)
	switch details.Op {
	case SoppOpWaitcnt:
		// No-op in SPIR-V for now.
	case SoppOpCbranchExecz:
		valLo, valHi := ctx.GetOperand64Value(b, OpExecLo, 0)
		val64 := ctx.Pack64(b, valLo, valHi)
		ctx.GcnConditionId = b.EmitIEqual(ctx.GetId(BlockContextIdTypeBool), val64, ctx.GetConstId(ConstIdx64Uint0))
	default:
		panic(fmt.Sprintf("unknown sopp op %s", Mnemotics[EncSOPP][details.Op]))
	}
}
