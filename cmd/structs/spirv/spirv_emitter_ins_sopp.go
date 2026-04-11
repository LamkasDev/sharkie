package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitSOPP(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	switch instr.Details.(*ScalarDetails).Op {
	case SoppOpWaitcnt:
		// No-op in SPIR-V for now.
	default:
		panic(fmt.Sprintf("unknown sopp op %d", instr.Details.(*ScalarDetails).Op))
	}
}
