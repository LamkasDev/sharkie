package spec

const (
	MtbufOpTbufferLoadFormatX     = 0x00
	MtbufOpTbufferLoadFormatXy    = 0x01
	MtbufOpTbufferLoadFormatXyz   = 0x02
	MtbufOpTbufferLoadFormatXyzw  = 0x03
	MtbufOpTbufferStoreFormatX    = 0x04
	MtbufOpTbufferStoreFormatXy   = 0x05
	MtbufOpTbufferStoreFormatXyz  = 0x06
	MtbufOpTbufferStoreFormatXyzw = 0x07
)

// Typed memory buffer instructions.
type MtbufDetails struct {
	Offset     uint32
	Offen      bool
	Idxen      bool
	Glc        bool
	Addr64     bool
	Op         uint32
	DataFormat uint32
	NumFormat  uint32

	Vaddr   uint32
	Vdata   uint32
	Srsrc   uint32
	Slc     bool
	Tfe     bool
	Soffset uint32
}

func (instr *Instruction) DecodeMTBUF() {
	dw0 := instr.Dwords[0]
	dw1 := instr.Dwords[1]
	instr.Details = &MtbufDetails{
		Offset:     dw0 & 0b1111_1111_1111,
		Offen:      (dw0>>12)&0b1 == 1,
		Idxen:      (dw0>>13)&0b1 == 1,
		Glc:        (dw0>>14)&0b1 == 1,
		Addr64:     (dw0>>15)&0b1 == 1,
		Op:         (dw0 >> 16) & 0b111,
		DataFormat: (dw0 >> 19) & 0b1111,
		NumFormat:  (dw0 >> 23) & 0b111,

		Vaddr:   dw1 & 0b1111_1111,
		Vdata:   (dw1 >> 8) & 0b1111_1111,
		Srsrc:   (dw1 >> 16) & 0b1111_1,
		Slc:     (dw1>>22)&0b1 == 1,
		Tfe:     (dw1>>23)&0b1 == 1,
		Soffset: (dw1 >> 24) & 0b1111_1111,
	}
}
