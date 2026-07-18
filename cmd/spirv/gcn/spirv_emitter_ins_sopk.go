package gcn

import (
	"fmt"

	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
)

func EmitSOPK(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*gcnSpec.ScalarDetails)
	switch details.Op {
	default:
		panic(fmt.Sprintf("unknown sopk op %s", gcnSpec.Mnemotics[gcnSpec.EncSOPK][details.Op]))
	}
}
