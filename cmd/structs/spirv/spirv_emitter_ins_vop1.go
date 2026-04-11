package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitVOP1(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	switch instr.Details.(*VectorDetails).Op {
	case Vop1OpMovB32:
		val := ctx.GetOperandValue(b, instr.Details.(*VectorDetails).Src0, instr.Literal)
		ptr := ctx.GetRegisterPointer(instr.Details.(*VectorDetails).Dst + OpVgpr0)
		b.EmitStore(ptr, val)
	default:
		panic(fmt.Sprintf("unknown vop1 op %d", instr.Details.(*VectorDetails).Op))
	}
}
