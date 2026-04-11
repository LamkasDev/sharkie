package gcn

const (
	MimgOpLoad           = 0x00
	MimgOpLoadMip        = 0x01
	MimgOpLoadPck        = 0x02
	MimgOpLoadPckSgn     = 0x03
	MimgOpLoadMipPck     = 0x04
	MimgOpLoadMipPckSgn  = 0x05
	MimgOpStore          = 0x08
	MimgOpStoreMip       = 0x09
	MimgOpStorePck       = 0x0A
	MimgOpStoreMipPck    = 0x0B
	MimgOpGetResinfo     = 0x0E
	MimgOpAtomicSwap     = 0x0F
	MimgOpAtomicCmpswap  = 0x10
	MimgOpAtomicAdd      = 0x11
	MimgOpAtomicSub      = 0x12
	MimgOpAtomicSmin     = 0x14
	MimgOpAtomicUmin     = 0x15
	MimgOpAtomicSmax     = 0x16
	MimgOpAtomicUmax     = 0x17
	MimgOpAtomicAnd      = 0x18
	MimgOpAtomicOr       = 0x19
	MimgOpAtomicXor      = 0x1A
	MimgOpAtomicInc      = 0x1B
	MimgOpAtomicDec      = 0x1C
	MimgOpAtomicFcmpswap = 0x1D
	MimgOpAtomicFmin     = 0x1E
	MimgOpAtomicFmax     = 0x1F
	MimgOpSample         = 0x20
	MimgOpSampleCl       = 0x21
	MimgOpSampleD        = 0x22
	MimgOpSampleDCl      = 0x23
	MimgOpSampleL        = 0x24
	MimgOpSampleB        = 0x25
	MimgOpSampleBCl      = 0x26
	MimgOpSampleLz       = 0x27
	MimgOpSampleC        = 0x28
	MimgOpSampleCCl      = 0x29
	MimgOpSampleCD       = 0x2A
	MimgOpSampleCDCl     = 0x2B
	MimgOpSampleCL       = 0x2C
	MimgOpSampleCB       = 0x2D
	MimgOpSampleCBCl     = 0x2E
	MimgOpSampleCLz      = 0x2F
	MimgOpSampleO        = 0x30
	MimgOpSampleClO      = 0x31
	MimgOpSampleDO       = 0x32
	MimgOpSampleDClO     = 0x33
	MimgOpSampleLO       = 0x34
	MimgOpSampleBO       = 0x35
	MimgOpSampleBClO     = 0x36
	MimgOpSampleLzO      = 0x37
	MimgOpSampleCO       = 0x38
	MimgOpSampleCClO     = 0x39
	MimgOpSampleCDO      = 0x3A
	MimgOpSampleCDClO    = 0x3B
	MimgOpSampleCLO      = 0x3C
	MimgOpSampleCBO      = 0x3D
	MimgOpSampleCBClO    = 0x3E
	MimgOpSampleCLzO     = 0x3F
	MimgOpGather4        = 0x40
	MimgOpGather4Cl      = 0x41
	MimgOpGather4L       = 0x42
	MimgOpGather4B       = 0x43
	MimgOpGather4BCl     = 0x44
	MimgOpGather4Lz      = 0x45
	MimgOpGather4C       = 0x46
	MimgOpGather4CCl     = 0x47
	MimgOpGather4CL      = 0x4C
	MimgOpGather4CB      = 0x4D
	MimgOpGather4CBCl    = 0x4E
	MimgOpGather4CLz     = 0x4F
	MimgOpGather4O       = 0x50
	MimgOpGather4ClO     = 0x51
	MimgOpGather4LO      = 0x54
	MimgOpGather4BO      = 0x55
	MimgOpGather4BClO    = 0x56
	MimgOpGather4LzO     = 0x57
	MimgOpGather4CO      = 0x58
	MimgOpGather4CClO    = 0x59
	MimgOpGather4CLO     = 0x5C
	MimgOpGather4CBO     = 0x5D
	MimgOpGather4CBClO   = 0x5E
	MimgOpGather4CLzO    = 0x5F
	MimgOpGetLod         = 0x60
	MimgOpSampleCd       = 0x68
	MimgOpSampleCdCl     = 0x69
	MimgOpSampleCCd      = 0x6A
	MimgOpSampleCCdCl    = 0x6B
	MimgOpSampleCdO      = 0x6C
	MimgOpSampleCdClO    = 0x6D
	MimgOpSampleCCdO     = 0x6E
	MimgOpSampleCCdClO   = 0x6F
)

// Vector memory image instructions.
type MimgDetails struct {
	Op    uint32
	Vaddr uint32
	Vdata uint32
	Srsrc uint32
	Ssamp uint32
	Dmask uint32
	Unorm bool
	Glc   bool
	Da    bool
	R128  bool
	Tfe   bool
	Lwe   bool
	Slc   bool
}

func (instr *Instruction) DecodeMIMG() {
	dw0 := instr.Dwords[0]
	dw1 := instr.Dwords[1]
	instr.Details = &MimgDetails{
		Dmask: (dw0 >> 8) & 0b1111,
		Unorm: (dw0>>12)&0b1 == 1,
		Glc:   (dw0>>13)&0b1 == 1,
		Da:    (dw0>>14)&0b1 == 1,
		R128:  (dw0>>15)&0b1 == 1,
		Tfe:   (dw0>>16)&0b1 == 1,
		Lwe:   (dw0>>17)&0b1 == 1,
		Op:    (dw0 >> 18) & 0b1111_111,
		Slc:   (dw0>>25)&0b1 == 1,

		Vaddr: dw1 & 0b1111_1111,
		Vdata: (dw1 >> 8) & 0b1111_1111,
		Srsrc: (dw1 >> 16) & 0b1111_1,
		Ssamp: (dw1 >> 21) & 0b1111_1,
	}
}
