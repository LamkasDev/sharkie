package gcn

import "fmt"

type InstructionDecodeFunc func(instr *Instruction)

var InstructionDecodeMap = map[Encoding]InstructionDecodeFunc{
	EncSOP2:  (*Instruction).DecodeSOP2,
	EncSOP1:  (*Instruction).DecodeSOP1,
	EncSOPC:  (*Instruction).DecodeSOPC,
	EncSOPP:  (*Instruction).DecodeSOPP,
	EncVOP2:  (*Instruction).DecodeVOP2,
	EncVOP1:  (*Instruction).DecodeVOP1,
	EncVOPC:  (*Instruction).DecodeVOPC,
	EncVOP3:  (*Instruction).DecodeVOP3,
	EncSMRD:  (*Instruction).DecodeSMRD,
	EncMUBUF: (*Instruction).DecodeMUBUF,
	EncMIMG:  (*Instruction).DecodeMIMG,
	EncEXP:   (*Instruction).DecodeEXP,
}

// Following based on this doc:
// https://docs.amd.com/v/u/en-US/sea-islands-instruction-set-architecture_0
type Instruction struct {
	Encoding    Encoding
	DwordOffset uintptr
	Dwords      [2]uint32
	DwordLen    int

	// Follows some instructions when SRC0/SSRC0 == 0xFF.
	HasLiteral bool
	Literal    uint32

	Details any
}

// Scalar instructions.
type ScalarDetails struct {
	Op    uint32
	Dst   uint32
	Src0  uint32
	Src1  uint32
	Imm16 int16
}

// Vector instructions.
type VectorDetails struct {
	Op   uint32
	Dst  uint32
	Src0 uint32
	Src1 uint32
}

func NewInstruction(dwordOffset uintptr, enc Encoding, dwords []uint32) (Instruction, error) {
	instr := Instruction{
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
