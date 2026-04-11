package gcn

const (
	// VOP2 opcodes.
	Vop3OpCndmaskB32      = 0x100
	Vop3OpReadlaneB32     = 0x101
	Vop3OpWritelaneB32    = 0x102
	Vop3OpAddF32          = 0x103
	Vop3OpSubF32          = 0x104
	Vop3OpSubrevF32       = 0x105
	Vop3OpMacLegacyF32    = 0x106
	Vop3OpMulLegacyF32    = 0x107
	Vop3OpMulF32          = 0x108
	Vop3OpMulI32I24       = 0x109
	Vop3OpMulHiI32I24     = 0x10A
	Vop3OpMulU32U24       = 0x10B
	Vop3OpMulHiU32U24     = 0x10C
	Vop3OpMinLegacyF32    = 0x10D
	Vop3OpMaxLegacyF32    = 0x10E
	Vop3OpMinF32          = 0x10F
	Vop3OpMaxF32          = 0x110
	Vop3OpMinI32          = 0x111
	Vop3OpMaxI32          = 0x112
	Vop3OpMinU32          = 0x113
	Vop3OpMaxU32          = 0x114
	Vop3OpLshrB32         = 0x115
	Vop3OpLshrrevB32      = 0x116
	Vop3OpAshrI32         = 0x117
	Vop3OpAshrrevI32      = 0x118
	Vop3OpLshlB32         = 0x119
	Vop3OpLshlrevB32      = 0x11A
	Vop3OpAndB32          = 0x11B
	Vop3OpOrB32           = 0x11C
	Vop3OpXorB32          = 0x11D
	Vop3OpBfmB32          = 0x11E
	Vop3OpMacF32          = 0x11F
	Vop3OpBcntU32B32      = 0x122
	Vop3OpMbcntHiU32B32   = 0x124
	Vop3OpAddI32          = 0x125
	Vop3OpSubI32          = 0x126
	Vop3OpSubrevI32       = 0x127
	Vop3OpAddcU32         = 0x128
	Vop3OpSubbU32         = 0x129
	Vop3OpSubbrevU32      = 0x12A
	Vop3OpLdexpF32        = 0x12B
	Vop3OpCvtPkaccumU8F32 = 0x12C
	Vop3OpCvtPknormI16F32 = 0x12D
	Vop3OpCvtPknormU16F32 = 0x12E
	Vop3OpCvtPkrtzF16F32  = 0x12F
	Vop3OpCvtPkU16U32     = 0x130
	Vop3OpCvtPkI16I32     = 0x131

	// VOP3 opcodes.
	Vop3OpMadLegacyF32 = 0x140
	Vop3OpMadF32       = 0x141
	Vop3OpMadI32I24    = 0x142
	Vop3OpMadU32U24    = 0x143
	Vop3OpCubeidF32    = 0x144
	Vop3OpCubescF32    = 0x145
	Vop3OpCubetcF32    = 0x146
	Vop3OpCubemaF32    = 0x147
	Vop3OpBfeU32       = 0x148
	Vop3OpBfeI32       = 0x149
	Vop3OpBfiB32       = 0x14A
	Vop3OpFmaF32       = 0x14B
	Vop3OpFmaF64       = 0x14C
	Vop3OpLerpU8       = 0x14D
	Vop3OpAlignbitB32  = 0x14E
	Vop3OpAlignbyteB32 = 0x14F
	Vop3OpMullitF32    = 0x150
	Vop3OpMin3F32      = 0x151
	Vop3OpMin3I32      = 0x152
	Vop3OpMin3U32      = 0x153
	Vop3OpMax3F32      = 0x154
	Vop3OpMax3I32      = 0x155
	Vop3OpMax3U32      = 0x156
	Vop3OpMed3F32      = 0x157
	Vop3OpMed3I32      = 0x158
	Vop3OpMed3U32      = 0x159
	Vop3OpSadU8        = 0x15A
	Vop3OpSadHiU8      = 0x15B
	Vop3OpSadU16       = 0x15C
	Vop3OpSadU32       = 0x15D
	Vop3OpCvtPkU8F32   = 0x15E
	Vop3OpDivFixupF32  = 0x15F
	Vop3OpDivFixupF64  = 0x160
	Vop3OpLshlB64      = 0x161
	Vop3OpLshrB64      = 0x162
	Vop3OpAshrI64      = 0x163
	Vop3OpAddF64       = 0x164
	Vop3OpMulF64       = 0x165
	Vop3OpMin3F64      = 0x166
	Vop3OpMaxF64       = 0x167
	Vop3OpLdexpF64     = 0x168
	Vop3OpMulLoU32     = 0x169
	Vop3OpMulHiU32     = 0x16A
	Vop3OpMulLoI32     = 0x16B
	Vop3OpMulHiI32     = 0x16C
	Vop3OpDivScaleF32  = 0x16D
	Vop3OpDivScaleF64  = 0x16E
	Vop3OpDivFmasF32   = 0x16F
	Vop3OpDivFmasF64   = 0x170
	Vop3OpMsadU8       = 0x171
	Vop3OpQsadPkU16U8  = 0x172
	Vop3OpMqsadPkU16U8 = 0x173
	Vop3OpTrigPreopF64 = 0x174
	Vop3OpMqsadU32U8   = 0x175
	Vop3OpMadU64U32    = 0x176
	Vop3OpMadI64I32    = 0x177

	// VOP1 opcodes.
	Vop3OpNop              = 0x180
	Vop3OpMovB32           = 0x181
	Vop3OpReadfirstlaneB32 = 0x182
	Vop3OpCvtI32F64        = 0x183
	Vop3OpCvtF64I32        = 0x184
	Vop3OpCvtF32I32        = 0x185
	Vop3OpCvtF32U32        = 0x186
	Vop3OpCvtU32F32        = 0x187
	Vop3OpCvtI32F32        = 0x188
	Vop3OpCvtF16F32        = 0x18A
	Vop3OpCvtF32F16        = 0x18B
	Vop3OpCvtRpiI32F32     = 0x18C
	Vop3OpCvtFlrI32F32     = 0x18D
	Vop3OpCvtOffF32I4      = 0x18E
	Vop3OpCvtF32F64        = 0x18F
	Vop3OpCvtF64F32        = 0x190
	Vop3OpCvtF32Ubyte0     = 0x191
	Vop3OpCvtF32Ubyte1     = 0x192
	Vop3OpCvtF32Ubyte2     = 0x193
	Vop3OpCvtF32Ubyte3     = 0x194
	Vop3OpCvtU32F64        = 0x195
	Vop3OpCvtF64U32        = 0x196
	Vop3OpFractF32         = 0x1A0
	Vop3OpTruncF32         = 0x1A1
	Vop3OpCeilF32          = 0x1A2
	Vop3OpRndneF32         = 0x1A3
	Vop3OpFloorF32         = 0x1A4
	Vop3OpExpF32           = 0x1A5
	Vop3OpLogClampF32      = 0x1A6
	Vop3OpLogF32           = 0x1A7
	Vop3OpRcpClampF32      = 0x1A8
	Vop3OpRcpLegacyF32     = 0x1A9
	Vop3OpRcpF32           = 0x1AA
	Vop3OpRcpIflagF32      = 0x1AB
	Vop3OpRsqClampF32      = 0x1AC
	Vop3OpRsqLegacyF32     = 0x1AD
	Vop3OpRsqF32           = 0x1AE
	Vop3OpRcpF64           = 0x1AF
	Vop3OpRcpClampF64      = 0x1B0
	Vop3OpRsqF64           = 0x1B1
	Vop3OpRsqClampF64      = 0x1B2
	Vop3OpSqrtF32          = 0x1B3
	Vop3OpSqrtF64          = 0x1B4
	Vop3OpSinF32           = 0x1B5
	Vop3OpCosF32           = 0x1B6
	Vop3OpNotB32           = 0x1B7
	Vop3OpBfrevB32         = 0x1B8
	Vop3OpFfbhU32          = 0x1B9
	Vop3OpFfblB32          = 0x1BA
	Vop3OpFfbhI32          = 0x1BB
	Vop3OpFrexpExpI32F64   = 0x1BC
	Vop3OpFrexpMantF64     = 0x1BD
	Vop3OpFractF64         = 0x1BE
	Vop3OpFrexpExpI32F32   = 0x1BF
	Vop3OpFrexpMantF32     = 0x1C0
	Vop3OpClrexcp          = 0x1C1
	Vop3OpMovreldB32       = 0x1C2
	Vop3OpMovrelsdB32      = 0x1C4
)

// VOP3 modifiers.
type Vop3Details struct {
	Op    uint32
	Dst   uint32
	Sdst  uint32
	Src0  uint32
	Src1  uint32
	Src2  uint32
	Abs   uint8
	Neg   uint8
	OMod  uint8
	Clamp bool
}

func (instr *Instruction) DecodeVOP3() {
	dw0 := instr.Dwords[0]
	dw1 := instr.Dwords[1]
	instr.Details = &Vop3Details{
		Dst:   dw0 & 0b1111_1111,
		Sdst:  (dw0 >> 8) & 0b1111_1111,
		Abs:   uint8((dw0 >> 8) & 0b111),
		Clamp: (dw0>>11)&0b1 == 1,
		Op:    (dw0 >> 17) & 0b1111_1111_1,

		Src0: dw1 & 0b1111_1111_1,
		Src1: (dw1 >> 9) & 0b1111_1111_1,
		Src2: (dw1 >> 18) & 0b1111_1111_1,
		OMod: uint8((dw1 >> 27) & 0b11),
		Neg:  uint8((dw1 >> 29) & 0b111),
	}
}
