package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitSOP2(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	details := instr.Details.(*ScalarDetails)
	switch details.Op {
	default:
		panic(fmt.Sprintf("unknown sop2 op %s", Mnemotics[EncSOP2][details.Op]))
	}
}
