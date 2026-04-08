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
	Encoding Encoding
	Dwords   [2]uint32
	DwordLen int

	// Follows some instructions when SRC0/SSRC0 == 0xFF.
	HasLiteral bool
	Literal    uint32

	// Scalar instructions.
	SOp    uint32 // instruction opcode
	SDst   uint32 // scalar destination register index (SOP2 / SOPK / SOP1)
	SSrc0  uint32 // scalar source 0 (SOP2 / SOP1 / SOPC)
	SSrc1  uint32 // scalar source 1 (SOP2 / SOPC)
	Simm16 uint32 // signed immediate (SOPK / SOPP)

	// Vector instructions.
	VOp   uint32 // instruction opcode
	VDst  uint32 // vector destination
	VSrc0 uint32 // vector source 0
	VSrc1 uint32 // vector source 1 (VOP2 / VOPC / VOP3)

	// VOP3 modifiers.
	VNeg   uint8  // negate bits for src2/src1/src0
	VOMod  uint8  // output modifier
	VSrc2  uint32 // vector source 2
	VClamp bool
	VAbs   uint8 // abs bits for src2/src1/src0
	VSdst  uint32

	// Scalar memory instructions.
	SmOp     uint32 // instruction
	SmDst    uint32 // scalar destination
	SmBase   uint32 // base SGPR pair index (actual SGPR = SBase*2)
	SmImmOff bool   // true = offset is immediate, false = offset is in M0
	SmOffset uint32 // byte or dword offset

	// Vector memory buffer instructions.
	VmbOffset  uint32
	VmbOffen   bool
	VmbIdxen   bool
	VmbGlc     bool
	VmbAddr64  bool
	VmbLds     bool
	VmbOp      uint32
	VmbVaddr   uint32
	VmbVdata   uint32
	VmbSrsrc   uint32
	VmbSlc     bool
	VmbTfe     bool
	VmbSoffset uint32

	// Vector memory image instructions.
	VmiDmask uint32
	VmiUnorm bool
	VmiGlc   bool
	VmiDa    bool
	VmiR128  bool
	VmiTfe   bool
	VmiLwe   bool
	VmiOp    uint32
	VmiSlc   bool
	VmiVaddr uint32
	VmiVdata uint32
	VmiSrsrc uint32
	VmiSsamp uint32

	// Export instructions.
	ExpEn     uint32
	ExpTarget uint32
	ExpCompr  bool
	ExpDone   bool
	ExpVm     bool
	ExpVSrcs  [4]uint32
}

func NewInstruction(enc Encoding, dwords []uint32) (Instruction, error) {
	instr := Instruction{
		Encoding: enc,
		DwordLen: len(dwords),
	}
	copy(instr.Dwords[:], dwords)
	decodeFunc, ok := InstructionDecodeMap[instr.Encoding]
	if !ok {
		return instr, fmt.Errorf("unknown encoding %s", enc)
	}
	decodeFunc(&instr)

	return instr, nil
}

func (instr *Instruction) DecodeSOP2() {
	dw := instr.Dwords[0]
	instr.SSrc0 = dw & 0b1111_1111
	instr.SSrc1 = (dw >> 8) & 0b1111_1111
	instr.SDst = (dw >> 16) & 0b1111_111
	instr.SOp = (dw >> 23) & 0b1111_111
	if instr.DwordLen > 1 {
		instr.HasLiteral = true
		instr.Literal = instr.Dwords[1]
	}
}

func (instr *Instruction) DecodeSOP1() {
	dw := instr.Dwords[0]
	instr.SSrc0 = dw & 0b1111_1111
	instr.SOp = (dw >> 8) & 0b1111_1111
	instr.SDst = (dw >> 16) & 0b1111_111
	if instr.DwordLen > 1 {
		instr.HasLiteral = true
		instr.Literal = instr.Dwords[1]
	}
}

func (instr *Instruction) DecodeSOPC() {
	dw := instr.Dwords[0]
	instr.SSrc0 = dw & 0b1111_1111
	instr.SSrc1 = (dw >> 8) & 0b1111_1111
	instr.SOp = (dw >> 16) & 0b1111_111
	if instr.DwordLen > 1 {
		instr.HasLiteral = true
		instr.Literal = instr.Dwords[1]
	}
}

func (instr *Instruction) DecodeSOPP() {
	dw := instr.Dwords[0]
	instr.Simm16 = dw & 0b1111_1111_1111_1111
	instr.SOp = (dw >> 16) & 0b1111_111
}

func (instr *Instruction) DecodeVOP2() {
	dw := instr.Dwords[0]
	instr.VSrc0 = dw & 0b1111_1111_1
	instr.VSrc1 = (dw >> 9) & 0b1111_1111
	instr.VDst = (dw >> 17) & 0b1111_1111
	instr.VOp = (dw >> 25) & 0b1111_111
	if instr.DwordLen > 1 {
		instr.HasLiteral = true
		instr.Literal = instr.Dwords[1]
	}
}

func (instr *Instruction) DecodeVOP1() {
	dw := instr.Dwords[0]
	instr.VSrc0 = dw & 0b1111_1111_1
	instr.VOp = (dw >> 9) & 0b1111_1111
	instr.VDst = (dw >> 17) & 0b1111_1111
	if instr.DwordLen > 1 {
		instr.HasLiteral = true
		instr.Literal = instr.Dwords[1]
	}
}

func (instr *Instruction) DecodeVOPC() {
	dw := instr.Dwords[0]
	instr.VSrc0 = dw & 0b1111_1111_1
	instr.VSrc1 = (dw >> 9) & 0b1111_1111
	instr.VOp = (dw >> 17) & 0b1111_1111
	if instr.DwordLen > 1 {
		instr.HasLiteral = true
		instr.Literal = instr.Dwords[1]
	}
}

func (instr *Instruction) DecodeVOP3() {
	dw0 := instr.Dwords[0]
	instr.VDst = dw0 & 0b1111_1111
	instr.VSdst = (dw0 >> 8) & 0b1111_1111
	instr.VAbs = uint8((dw0 >> 8) & 0b111)
	instr.VClamp = (dw0>>11)&0b1 == 1
	instr.VOp = (dw0 >> 17) & 0b1111_1111_1

	dw1 := instr.Dwords[1]
	instr.VSrc0 = dw1 & 0b1111_1111_1
	instr.VSrc1 = (dw1 >> 9) & 0b1111_1111_1
	instr.VSrc2 = (dw1 >> 18) & 0b1111_1111_1
	instr.VOMod = uint8((dw1 >> 27) & 0b11)
	instr.VNeg = uint8((dw1 >> 29) & 0b111)
}

func (instr *Instruction) DecodeSMRD() {
	dw := instr.Dwords[0]
	instr.SmOffset = dw & 0b1111_1111
	instr.SmImmOff = (dw>>8)&0b1 == 1
	instr.SmBase = (dw >> 9) & 0b1111_111
	instr.SmDst = (dw >> 16) & 0b1111_11
	instr.SmOp = (dw >> 22) & 0b1111_1
	if instr.DwordLen > 1 {
		instr.HasLiteral = true
		instr.Literal = instr.Dwords[1]
	}
}

func (instr *Instruction) DecodeMUBUF() {
	// Dword 0.
	dw0 := instr.Dwords[0]
	instr.VmbOffset = dw0 & 0b1111_1111_1111
	instr.VmbOffen = (dw0>>12)&0b1 == 1
	instr.VmbIdxen = (dw0>>13)&0b1 == 1
	instr.VmbGlc = (dw0>>14)&0b1 == 1
	instr.VmbAddr64 = (dw0>>15)&0b1 == 1
	instr.VmbLds = (dw0>>16)&0b1 == 1
	instr.VmbOp = (dw0 >> 18) & 0b1111_111

	// Dword 1.
	dw1 := instr.Dwords[1]
	instr.VmbVaddr = dw1 & 0b1111_1111
	instr.VmbVdata = (dw1 >> 8) & 0b1111_1111
	instr.VmbSrsrc = (dw1 >> 16) & 0b1111_1
	instr.VmbSlc = (dw1>>22)&0b1 == 1
	instr.VmbTfe = (dw1>>23)&0b1 == 1
	instr.VmbSoffset = (dw1 >> 24) & 0b1111_1111
}

func (instr *Instruction) DecodeMIMG() {
	// Dword 0.
	dw0 := instr.Dwords[0]
	instr.VmiDmask = (dw0 >> 8) & 0b1111
	instr.VmiUnorm = (dw0>>12)&0b1 == 1
	instr.VmiGlc = (dw0>>13)&0b1 == 1
	instr.VmiDa = (dw0>>14)&0b1 == 1
	instr.VmiR128 = (dw0>>15)&0b1 == 1
	instr.VmiTfe = (dw0>>16)&0b1 == 1
	instr.VmiLwe = (dw0>>17)&0b1 == 1
	instr.VmiOp = (dw0 >> 18) & 0b1111_111
	instr.VmiSlc = (dw0>>25)&0b1 == 1

	// Dword 1.
	dw1 := instr.Dwords[1]
	instr.VmiVaddr = dw1 & 0b1111_1111
	instr.VmiVdata = (dw1 >> 8) & 0b1111_1111
	instr.VmiSrsrc = (dw1 >> 16) & 0b1111_1
	instr.VmiSsamp = (dw1 >> 21) & 0b1111_1
}

func (instr *Instruction) DecodeEXP() {
	// Dword 0.
	dw0 := instr.Dwords[0]
	instr.ExpEn = dw0 & 0b1111
	instr.ExpTarget = (dw0 >> 4) & 0b1111_11
	instr.ExpCompr = (dw0>>10)&0b1 == 1
	instr.ExpDone = (dw0>>11)&0b1 == 1
	instr.ExpVm = (dw0>>12)&0b1 == 1

	// Dword 1.
	dw1 := instr.Dwords[1]
	instr.ExpVSrcs[0] = dw1 & 0b1111_1111
	instr.ExpVSrcs[1] = (dw1 >> 8) & 0b1111_1111
	instr.ExpVSrcs[2] = (dw1 >> 16) & 0b1111_1111
	instr.ExpVSrcs[3] = (dw1 >> 24) & 0b1111_1111
}
