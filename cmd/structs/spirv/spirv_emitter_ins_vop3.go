package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitVOP3(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	details := instr.Details.(*Vop3Details)
	switch details.Op {
	default:
		panic(fmt.Sprintf("unknown vop3 op %d", details.Op))
	}
}
