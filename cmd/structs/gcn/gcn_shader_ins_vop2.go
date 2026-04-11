package gcn

const (
	Vop2OpCndmaskB32      = 0x00
	Vop2OpReadlaneB32     = 0x01
	Vop2OpWritelaneB32    = 0x02
	Vop2OpAddF32          = 0x03
	Vop2OpSubF32          = 0x04
	Vop2OpSubrevF32       = 0x05
	Vop2OpMacLegacyF32    = 0x06
	Vop2OpMulLegacyF32    = 0x07
	Vop2OpMulF32          = 0x08
	Vop2OpMulI32I24       = 0x09
	Vop2OpMulHiI32I24     = 0x0A
	Vop2OpMulU32U24       = 0x0B
	Vop2OpMulHiU32U24     = 0x0C
	Vop2OpMinLegacyF32    = 0x0D
	Vop2OpMaxLegacyF32    = 0x0E
	Vop2OpMinF32          = 0x0F
	Vop2OpMaxF32          = 0x10
	Vop2OpMinI32          = 0x11
	Vop2OpMaxI32          = 0x12
	Vop2OpMinU32          = 0x13
	Vop2OpMaxU32          = 0x14
	Vop2OpLshrB32         = 0x15
	Vop2OpLshrrevB32      = 0x16
	Vop2OpAshrI32         = 0x17
	Vop2OpAshrrevI32      = 0x18
	Vop2OpLshlB32         = 0x19
	Vop2OpLshlrevB32      = 0x1A
	Vop2OpAndB32          = 0x1B
	Vop2OpOrB32           = 0x1C
	Vop2OpXorB32          = 0x1D
	Vop2OpBfmB32          = 0x1E
	Vop2OpMacF32          = 0x1F
	Vop2OpMadmkF32        = 0x20
	Vop2OpMadakF32        = 0x21
	Vop2OpBcntU32B32      = 0x22
	Vop2OpMbcntLoU32B32   = 0x23
	Vop2OpMbcntHiU32B32   = 0x24
	Vop2OpAddI32          = 0x25
	Vop2OpSubI32          = 0x26
	Vop2OpSubrevI32       = 0x27
	Vop2OpAddcU32         = 0x28
	Vop2OpSubbU32         = 0x29
	Vop2OpSubbrevU32      = 0x2A
	Vop2OpLdexpF32        = 0x2B
	Vop2OpCvtPkaccumU8F32 = 0x2C
	Vop2OpCvtPknormI16F32 = 0x2D
	Vop2OpCvtPknormU16F32 = 0x2E
	Vop2OpCvtPkrtzF16F32  = 0x2F
	Vop2OpCvtPkU16U32     = 0x30
	Vop2OpCvtPkI16I32     = 0x31
)

func (instr *Instruction) DecodeVOP2() {
	dw := instr.Dwords[0]
	instr.Details = &VectorDetails{
		Src0: dw & 0b1111_1111_1,
		Src1: (dw >> 9) & 0b1111_1111,
		Dst:  (dw >> 17) & 0b1111_1111,
		Op:   (dw >> 25) & 0b111_1111,
	}
	if instr.DwordLen > 1 {
		instr.HasLiteral = true
		instr.Literal = instr.Dwords[1]
	}
}
