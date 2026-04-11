package gcn

const (
	SmrdOpLoadDword          = 0x00
	SmrdOpLoadDwordx2        = 0x01
	SmrdOpLoadDwordx4        = 0x02
	SmrdOpLoadDwordx8        = 0x03
	SmrdOpLoadDwordx16       = 0x04
	SmrdOpBufferLoadDword    = 0x08
	SmrdOpBufferLoadDwordx2  = 0x09
	SmrdOpBufferLoadDwordx4  = 0x0A
	SmrdOpBufferLoadDwordx8  = 0x0B
	SmrdOpBufferLoadDwordx16 = 0x0C
	SmrdOpDcacheInvVol       = 0x1D
	SmrdOpMemtime            = 0x1E
	SmrdOpDcacheInv          = 0x1F
)

// Scalar memory instructions.
type SmrdDetails struct {
	Op     uint32
	Dst    uint32
	Base   uint32
	Offset uint32
	ImmOff bool
}

func (instr *Instruction) DecodeSMRD() {
	dw := instr.Dwords[0]
	instr.Details = &SmrdDetails{
		Offset: dw & 0b1111_1111,
		ImmOff: (dw>>8)&0b1 == 1,
		Base:   (dw >> 9) & 0b1111_111,
		Dst:    (dw >> 16) & 0b1111_11,
		Op:     (dw >> 22) & 0b1111_1,
	}
	if instr.DwordLen > 1 {
		instr.HasLiteral = true
		instr.Literal = instr.Dwords[1]
	}
}
