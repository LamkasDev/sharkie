package gcn

const (
	SopcOpCmpEqI32   = 0x00
	SopcOpCmpLgI32   = 0x01
	SopcOpCmpGtI32   = 0x02
	SopcOpCmpGeI32   = 0x03
	SopcOpCmpLtI32   = 0x04
	SopcOpCmpLeI32   = 0x05
	SopcOpCmpEqU32   = 0x06
	SopcOpCmpLgU32   = 0x07
	SopcOpCmpGtU32   = 0x08
	SopcOpCmpGeU32   = 0x09
	SopcOpCmpLtU32   = 0x0A
	SopcOpCmpLeU32   = 0x0B
	SopcOpBitcmp0B32 = 0x0C
	SopcOpBitcmp1B32 = 0x0D
	SopcOpBitcmp0B64 = 0x0E
	SopcOpBitcmp1B64 = 0x0F
	SopcOpSetvskip   = 0x10
)

func (instr *Instruction) DecodeSOPC() {
	dw := instr.Dwords[0]
	instr.Details = &ScalarDetails{
		Src0: dw & 0b1111_1111,
		Src1: (dw >> 8) & 0b1111_1111,
		Op:   (dw >> 16) & 0b1111_111,
	}
	if instr.DwordLen > 1 {
		instr.HasLiteral = true
		instr.Literal = instr.Dwords[1]
	}
}
