package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitSOPC(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	details := instr.Details.(*ScalarDetails)
	switch details.Op {
	default:
		panic(fmt.Sprintf("unknown sopc op %s", Mnemotics[EncSOPC][details.Op]))
	}
}
