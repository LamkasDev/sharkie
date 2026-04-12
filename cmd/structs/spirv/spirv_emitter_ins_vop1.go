package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitVOP1(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	details := instr.Details.(*VectorDetails)
	switch details.Op {
	case Vop1OpMovB32:
		val := ctx.GetOperandValue(b, details.Src0, instr.Literal)
		ctx.StoreRegisterPointer(b, details.Dst+OpVgpr0, val)
	default:
		panic(fmt.Sprintf("unknown vop1 op %s", Mnemotics[EncVOP1][details.Op]))
	}
}
