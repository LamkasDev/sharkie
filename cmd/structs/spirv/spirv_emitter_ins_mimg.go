package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitMIMG(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	details := instr.Details.(*MimgDetails)
	switch details.Op {
	default:
		panic(fmt.Sprintf("unknown mimg op %s", Mnemotics[EncMIMG][details.Op]))
	}
}
