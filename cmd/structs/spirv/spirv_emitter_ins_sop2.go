package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitSOP2(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	switch instr.Details.(*ScalarDetails).Op {
	default:
		panic(fmt.Sprintf("unknown sop2 op %d", instr.Details.(*ScalarDetails).Op))
	}
}
