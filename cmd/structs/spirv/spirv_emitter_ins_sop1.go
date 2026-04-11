package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitSOP1(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	switch instr.Details.(*ScalarDetails).Op {
	case Sop1OpMovB32:
		val := ctx.GetOperandValue(b, instr.Details.(*ScalarDetails).Src0, instr.Literal)
		ptr := ctx.GetRegisterPointer(instr.Details.(*ScalarDetails).Dst)
		b.EmitStore(ptr, val)
	default:
		panic(fmt.Sprintf("unknown sop1 op %d", instr.Details.(*ScalarDetails).Op))
	}
}
