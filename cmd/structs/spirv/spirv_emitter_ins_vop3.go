package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitVOP3(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	switch instr.Details.(*Vop3Details).Op {
	default:
		panic(fmt.Sprintf("unknown vop3 op %d", instr.Details.(*VectorDetails).Op))
	}
}
