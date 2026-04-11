package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitVOPC(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	switch instr.Details.(*VectorDetails).Op {
	default:
		panic(fmt.Sprintf("unknown vopc op %d", instr.Details.(*VectorDetails).Op))
	}
}
