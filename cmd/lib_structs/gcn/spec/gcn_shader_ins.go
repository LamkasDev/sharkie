package spec

import "fmt"

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
	Vdst uint32
	Src0 uint32
	Src1 uint32
	Sdst uint32 // VCC for most instructions, SDST for VOP3b.

	// Details for redirected VOP3 instructions.
	Abs   uint8
	Neg   uint8
	OMod  uint8
	Clamp bool
	Src2  uint32
}

func (details *VectorDetails) IsVopcCmpx() bool {
	if details.Op >= 0x10 && details.Op <= 0x1F ||
		details.Op >= 0x30 && details.Op <= 0x3F ||
		details.Op >= 0x50 && details.Op <= 0x5F ||
		details.Op >= 0x70 && details.Op <= 0x7F ||
		details.Op >= 0x90 && details.Op <= 0x97 ||
		details.Op >= 0xB0 && details.Op <= 0xB7 ||
		details.Op >= 0xD0 && details.Op <= 0xD7 ||
		details.Op >= 0xF0 && details.Op <= 0xF7 {
		return true
	}

	return false
}

func (instr *Instruction) String() string {
	if instr.Encoding == EncUnknown {
		return fmt.Sprintf("%-6s  0x%08X                                   ; UNKNOWN", "?", instr.Dwords[0])
	}
	rawHex := fmt.Sprintf("0x%08X", instr.Dwords[0])
	if instr.DwordLen == 2 {
		rawHex += fmt.Sprintf(" 0x%08X", instr.Dwords[1])
	}

	return fmt.Sprintf("%-6s  %-22s  %-24s  %s", instr.Encoding, rawHex, instr.GetMnemotic(), instr.GetFieldsString())
}
