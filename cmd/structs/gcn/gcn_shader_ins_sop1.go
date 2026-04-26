package gcn

const (
	Sop1OpMovB32           = 0x03
	Sop1OpMovB64           = 0x04
	Sop1OpCmovB32          = 0x05
	Sop1OpCmovB64          = 0x06
	Sop1OpNotB32           = 0x07
	Sop1OpNotB64           = 0x08
	Sop1OpWqmB32           = 0x09
	Sop1OpWqmB64           = 0x0A
	Sop1OpBrevB32          = 0x0B
	Sop1OpBrevB64          = 0x0C
	Sop1OpBcnt0I32B32      = 0x0D
	Sop1OpBcnt0I32B64      = 0x0E
	Sop1OpBcnt1I32B32      = 0x0F
	Sop1OpBcnt1I32B64      = 0x10
	Sop1OpFf0I32B32        = 0x11
	Sop1OpFf0I32B64        = 0x12
	Sop1OpFf1I32B32        = 0x13
	Sop1OpFf1I32B64        = 0x14
	Sop1OpFlbitI32B32      = 0x15
	Sop1OpFlbitI32B64      = 0x16
	Sop1OpFlbitI32         = 0x17
	Sop1OpFlbitI32I64      = 0x18
	Sop1OpSextI32I8        = 0x19
	Sop1OpSextI32I16       = 0x1A
	Sop1OpBitset0B32       = 0x1B
	Sop1OpBitset0B64       = 0x1C
	Sop1OpBitset1B32       = 0x1D
	Sop1OpBitset1B64       = 0x1E
	Sop1OpGetpcB64         = 0x1F
	Sop1OpSetpcB64         = 0x20
	Sop1OpSwappcB64        = 0x21
	Sop1OpRfeB64           = 0x22
	Sop1OpAndSaveexecB64   = 0x24
	Sop1OpOrSaveexecB64    = 0x25
	Sop1OpXorSaveexecB64   = 0x26
	Sop1OpAndn2SaveexecB64 = 0x27
	Sop1OpOrn2SaveexecB64  = 0x28
	Sop1OpNandSaveexecB64  = 0x29
	Sop1OpNorSaveexecB64   = 0x2A
	Sop1OpXnorSaveexecB64  = 0x2B
	Sop1OpQuadmaskB32      = 0x2C
	Sop1OpQuadmaskB64      = 0x2D
	Sop1OpMovrelsB32       = 0x2E
	Sop1OpMovrelsB64       = 0x2F
	Sop1OpMovreldB32       = 0x30
	Sop1OpMovreldB64       = 0x31
	Sop1OpCbranchJoin      = 0x32
	Sop1OpAbsI32           = 0x34
)

func (instr *Instruction) DecodeSOP1() {
	dw := instr.Dwords[0]
	instr.Details = &ScalarDetails{
		Src0: dw & 0b1111_1111,
		Op:   (dw >> 8) & 0b1111_1111,
		Dst:  (dw >> 16) & 0b1111_111,
	}
	if instr.DwordLen > 1 {
		instr.HasLiteral = true
		instr.Literal = instr.Dwords[1]
	}
}
