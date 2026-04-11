package gcn

const (
	MubufOpLoadFormatX      = 0x00
	MubufOpLoadFormatXy     = 0x01
	MubufOpLoadFormatXyz    = 0x02
	MubufOpLoadFormatXyzw   = 0x03
	MubufOpStoreFormatX     = 0x04
	MubufOpStoreFormatXy    = 0x05
	MubufOpStoreFormatXyz   = 0x06
	MubufOpStoreFormatXyzw  = 0x07
	MubufOpLoadUbyte        = 0x08
	MubufOpLoadSbyte        = 0x09
	MubufOpLoadUshort       = 0x0A
	MubufOpLoadSshort       = 0x0B
	MubufOpLoadDword        = 0x0C
	MubufOpLoadDwordx2      = 0x0D
	MubufOpLoadDwordx4      = 0x0E
	MubufOpLoadDwordx3      = 0x0F
	MubufOpStoreByte        = 0x18
	MubufOpStoreShort       = 0x1A
	MubufOpStoreDword       = 0x1C
	MubufOpStoreDwordx2     = 0x1D
	MubufOpStoreDwordx4     = 0x1E
	MubufOpStoreDwordx3     = 0x1F
	MubufOpAtomicSwap       = 0x30
	MubufOpAtomicCmpswap    = 0x31
	MubufOpAtomicAdd        = 0x32
	MubufOpAtomicSub        = 0x33
	MubufOpAtomicSmin       = 0x35
	MubufOpAtomicUmin       = 0x36
	MubufOpAtomicSmax       = 0x37
	MubufOpAtomicUmax       = 0x38
	MubufOpAtomicAnd        = 0x39
	MubufOpAtomicOr         = 0x3A
	MubufOpAtomicXor        = 0x3B
	MubufOpAtomicInc        = 0x3C
	MubufOpAtomicDec        = 0x3D
	MubufOpAtomicFcmpswap   = 0x3E
	MubufOpAtomicFmin       = 0x3F
	MubufOpAtomicFmax       = 0x40
	MubufOpAtomicSwapX2     = 0x50
	MubufOpAtomicCmpswapX2  = 0x51
	MubufOpAtomicAddX2      = 0x52
	MubufOpAtomicSubX2      = 0x53
	MubufOpAtomicSminX2     = 0x55
	MubufOpAtomicUminX2     = 0x56
	MubufOpAtomicSmaxX2     = 0x57
	MubufOpAtomicUmaxX2     = 0x58
	MubufOpAtomicAndX2      = 0x59
	MubufOpAtomicOrX2       = 0x5A
	MubufOpAtomicXorX2      = 0x5B
	MubufOpAtomicIncX2      = 0x5C
	MubufOpAtomicDecX2      = 0x5D
	MubufOpAtomicFcmpswapX2 = 0x5E
	MubufOpAtomicFminX2     = 0x5F
	MubufOpAtomicFmaxX2     = 0x60
	MubufOpWbinvl1Vol       = 0x70
	MubufOpWbinvl1          = 0x71
)

// Vector memory buffer instructions.
type MubufDetails struct {
	Op      uint32
	Vaddr   uint32
	Vdata   uint32
	Srsrc   uint32
	Soffset uint32
	Offset  uint32
	Offen   bool
	Idxen   bool
	Glc     bool
	Addr64  bool
	Lds     bool
	Slc     bool
	Tfe     bool
}

func (instr *Instruction) DecodeMUBUF() {
	dw0 := instr.Dwords[0]
	dw1 := instr.Dwords[1]
	instr.Details = &MubufDetails{
		Offset: dw0 & 0b1111_1111_1111,
		Offen:  (dw0>>12)&0b1 == 1,
		Idxen:  (dw0>>13)&0b1 == 1,
		Glc:    (dw0>>14)&0b1 == 1,
		Addr64: (dw0>>15)&0b1 == 1,
		Lds:    (dw0>>16)&0b1 == 1,
		Op:     (dw0 >> 18) & 0b1111_111,

		Vaddr:   dw1 & 0b1111_1111,
		Vdata:   (dw1 >> 8) & 0b1111_1111,
		Srsrc:   (dw1 >> 16) & 0b1111_1,
		Slc:     (dw1>>22)&0b1 == 1,
		Tfe:     (dw1>>23)&0b1 == 1,
		Soffset: (dw1 >> 24) & 0b1111_1111,
	}
}
