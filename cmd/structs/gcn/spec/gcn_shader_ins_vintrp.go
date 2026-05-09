package spec

const (
	VintrpOpInterpP1F32  = 0x0
	VintrpOpInterpP2F32  = 0x1
	VintrpOpInterpMovF32 = 0x2
)

type VintrpDetails struct {
	Vsrc uint32
	Attr uint32
	Chan uint32
	Vdst uint32
	Op   uint32
}

func (instr *Instruction) DecodeVINTRP() {
	dw := instr.Dwords[0]
	instr.Details = &VintrpDetails{
		Vsrc: dw & 0xFF,
		Attr: (dw >> 10) & 0x3F,
		Chan: (dw >> 8) & 0x3,
		Vdst: (dw >> 18) & 0xFF,
		Op:   (dw >> 16) & 0x3,
	}
}
