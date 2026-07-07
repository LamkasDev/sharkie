package gcn

import (
	"fmt"

	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
)

func EmitSOPP(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*gcnSpec.ScalarDetails)
	switch details.Op {
	case gcnSpec.SoppOpWaitcnt:
		// No-op in SPIR-V for now.
	case gcnSpec.SoppOpCbranchExecz:
		valLo, valHi := ctx.GetOperand64Value(b, gcnSpec.OpExecLo, 0)
		val64 := ctx.Pack64(b, valLo, valHi)
		ctx.GcnConditionId = b.EmitIEqual(ctx.GetId(BlockContextIdTypeBool), val64, ctx.GetConstId(ConstId64Uint0))
	case gcnSpec.SoppOpEndpgm:
		// Not sure about this lol.
	default:
		panic(fmt.Sprintf("unknown sopp op %s", gcnSpec.Mnemotics[gcnSpec.EncSOPP][details.Op]))
	}
}
