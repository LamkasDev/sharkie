package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitMIMG(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	switch instr.Details.(*MimgDetails).Op {
	default:
		panic(fmt.Sprintf("unknown mimg op %d", instr.Details.(*MimgDetails).Op))
	}
}
