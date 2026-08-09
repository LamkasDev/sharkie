package spirv

import (
	"fmt"

	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	"github.com/LamkasDev/sharkie/cmd/spirv/gcn"
)

type InstructionEmitFunc func(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext)

var InstructionEmitMap = map[gcnSpec.Encoding]InstructionEmitFunc{
	gcnSpec.EncSOP2:   gcn.EmitSOP2,
	gcnSpec.EncSOPK:   gcn.EmitSOPK,
	gcnSpec.EncSOP1:   gcn.EmitSOP1,
	gcnSpec.EncSOPC:   gcn.EmitSOPC,
	gcnSpec.EncSOPP:   gcn.EmitSOPP,
	gcnSpec.EncVOP2:   gcn.EmitVOP2,
	gcnSpec.EncVOP1:   gcn.EmitVOP1,
	gcnSpec.EncVOPC:   gcn.EmitVOPC,
	gcnSpec.EncVOP3:   gcn.EmitVOP3,
	gcnSpec.EncVINTRP: gcn.EmitVINTRP,
	gcnSpec.EncSMRD:   gcn.EmitSMRD,
	gcnSpec.EncMTBUF:  gcn.EmitMTBUF,
	gcnSpec.EncMUBUF:  gcn.EmitMUBUF,
	gcnSpec.EncMIMG:   gcn.EmitMIMG,
	gcnSpec.EncDS:     gcn.EmitDS,
	gcnSpec.EncEXP:    gcn.EmitEXP,
}

// emitInstruction emits the SPIR-V for a single instruction.
func emitInstruction(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext) {
	b.EmitLine(b.EmitString(instr.String()), uint32(instr.DwordOffset), 0)
	emitFunc, ok := InstructionEmitMap[instr.Encoding]
	if !ok {
		panic(fmt.Errorf("unknown encoding %s", instr.Encoding))
	}
	emitFunc(b, instr, ctx)
}
