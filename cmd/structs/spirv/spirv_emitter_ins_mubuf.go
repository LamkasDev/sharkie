package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitMUBUF(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	switch instr.Details.(*MubufDetails).Op {
	default:
		panic(fmt.Sprintf("unknown mubuf op %d", instr.Details.(*MubufDetails).Op))
	}
}
