package gcn

const (
	Sop2OpAddU32       = 0x00
	Sop2OpSubU32       = 0x01
	Sop2OpAddI32       = 0x02
	Sop2OpSubI32       = 0x03
	Sop2OpAddcU32      = 0x04
	Sop2OpSubbU32      = 0x05
	Sop2OpMinI32       = 0x06
	Sop2OpMinU32       = 0x07
	Sop2OpMaxI32       = 0x08
	Sop2OpMaxU32       = 0x09
	Sop2OpCselectB32   = 0x0A
	Sop2OpCselectB64   = 0x0B
	Sop2OpAndB32       = 0x0E
	Sop2OpAndB64       = 0x0F
	Sop2OpOrB32        = 0x10
	Sop2OpOrB64        = 0x11
	Sop2OpXorB32       = 0x12
	Sop2OpXorB64       = 0x13
	Sop2OpAndn2B32     = 0x14
	Sop2OpAndn2B64     = 0x15
	Sop2OpOrn2B32      = 0x16
	Sop2OpOrn2B64      = 0x17
	Sop2OpNandB32      = 0x18
	Sop2OpNandB64      = 0x19
	Sop2OpNorB32       = 0x1A
	Sop2OpNorB64       = 0x1B
	Sop2OpXnorB32      = 0x1C
	Sop2OpXnorB64      = 0x1D
	Sop2OpLshlB32      = 0x1E
	Sop2OpLshlB64      = 0x1F
	Sop2OpLshrB32      = 0x20
	Sop2OpLshrB64      = 0x21
	Sop2OpAshrI32      = 0x22
	Sop2OpAshrI64      = 0x23
	Sop2OpBfmB32       = 0x24
	Sop2OpBfmB64       = 0x25
	Sop2OpMulI32       = 0x26
	Sop2OpBfeU32       = 0x27
	Sop2OpBfeI32       = 0x28
	Sop2OpBfeU64       = 0x29
	Sop2OpBfeI64       = 0x2A
	Sop2OpCbranchGFork = 0x2B
	Sop2OpAbsdiffI32   = 0x2C
)

func (instr *Instruction) DecodeSOP2() {
	dw := instr.Dwords[0]
	instr.Details = &ScalarDetails{
		Src0: dw & 0b1111_1111,
		Src1: (dw >> 8) & 0b1111_1111,
		Dst:  (dw >> 16) & 0b1111_111,
		Op:   (dw >> 23) & 0b1111_111,
	}
	if instr.DwordLen > 1 {
		instr.HasLiteral = true
		instr.Literal = instr.Dwords[1]
	}
}
