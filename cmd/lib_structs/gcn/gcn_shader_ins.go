package gcn

import (
	"fmt"

	"github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
)

type InstructionDecodeFunc func(instr *spec.Instruction)

var InstructionDecodeMap = map[spec.Encoding]InstructionDecodeFunc{
	spec.EncSOP2:   (*spec.Instruction).DecodeSOP2,
	spec.EncSOPK:   (*spec.Instruction).DecodeSOPK,
	spec.EncSOP1:   (*spec.Instruction).DecodeSOP1,
	spec.EncSOPC:   (*spec.Instruction).DecodeSOPC,
	spec.EncSOPP:   (*spec.Instruction).DecodeSOPP,
	spec.EncVOP2:   (*spec.Instruction).DecodeVOP2,
	spec.EncVOP1:   (*spec.Instruction).DecodeVOP1,
	spec.EncVOPC:   (*spec.Instruction).DecodeVOPC,
	spec.EncVOP3:   (*spec.Instruction).DecodeVOP3,
	spec.EncVINTRP: (*spec.Instruction).DecodeVINTRP,
	spec.EncSMRD:   (*spec.Instruction).DecodeSMRD,
	spec.EncMTBUF:  (*spec.Instruction).DecodeMTBUF,
	spec.EncMUBUF:  (*spec.Instruction).DecodeMUBUF,
	spec.EncMIMG:   (*spec.Instruction).DecodeMIMG,
	spec.EncDS:     (*spec.Instruction).DecodeDS,
	spec.EncEXP:    (*spec.Instruction).DecodeEXP,
}

func NewInstruction(dwordOffset uintptr, enc spec.Encoding, dwords []uint32) (spec.Instruction, error) {
	instr := spec.Instruction{
		Encoding:    enc,
		DwordOffset: dwordOffset,
		DwordLen:    len(dwords),
	}
	copy(instr.Dwords[:], dwords)
	decodeFunc, ok := InstructionDecodeMap[instr.Encoding]
	if !ok {
		return instr, fmt.Errorf("unknown encoding %s", enc)
	}
	decodeFunc(&instr)

	return instr, nil
}
