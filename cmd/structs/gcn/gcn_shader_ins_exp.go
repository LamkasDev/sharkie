package gcn

// Export instructions.
type ExpDetails struct {
	En     uint32
	Target uint32
	VSrcs  [4]uint32
	Compr  bool
	Done   bool
	Vm     bool
}

func (instr *Instruction) DecodeEXP() {
	dw0 := instr.Dwords[0]
	dw1 := instr.Dwords[1]
	instr.Details = &ExpDetails{
		En:     dw0 & 0b1111,
		Target: (dw0 >> 4) & 0b1111_11,
		Compr:  (dw0>>10)&0b1 == 1,
		Done:   (dw0>>11)&0b1 == 1,
		Vm:     (dw0>>12)&0b1 == 1,

		VSrcs: [4]uint32{
			dw1 & 0b1111_1111,
			(dw1 >> 8) & 0b1111_1111,
			(dw1 >> 16) & 0b1111_1111,
			(dw1 >> 24) & 0b1111_1111,
		},
	}
}
