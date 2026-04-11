package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitSOPC(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	switch instr.Details.(*ScalarDetails).Op {
	default:
		panic(fmt.Sprintf("unknown sopc op %d", instr.Details.(*ScalarDetails).Op))
	}
}
