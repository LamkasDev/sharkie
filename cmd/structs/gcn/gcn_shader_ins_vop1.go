package gcn

const (
	Vop1OpNop              = 0x00
	Vop1OpMovB32           = 0x01
	Vop1OpReadfirstlaneB32 = 0x02
	Vop1OpCvtI32F64        = 0x03
	Vop1OpCvtF64I32        = 0x04
	Vop1OpCvtF32I32        = 0x05
	Vop1OpCvtF32U32        = 0x06
	Vop1OpCvtU32F32        = 0x07
	Vop1OpCvtI32F32        = 0x08
	Vop1OpCvtF16F32        = 0x0A
	Vop1OpCvtF32F16        = 0x0B
	Vop1OpCvtRpiI32F32     = 0x0C
	Vop1OpCvtFlrI32F32     = 0x0D
	Vop1OpCvtOffF32I4      = 0x0E
	Vop1OpCvtF32F64        = 0x0F
	Vop1OpCvtF64F32        = 0x10
	Vop1OpCvtF32Ubyte0     = 0x11
	Vop1OpCvtF32Ubyte1     = 0x12
	Vop1OpCvtF32Ubyte2     = 0x13
	Vop1OpCvtF32Ubyte3     = 0x14
	Vop1OpCvtU32F64        = 0x15
	Vop1OpCvtF64U32        = 0x16
	Vop1OpTruncF64         = 0x17
	Vop1OpCeilF64          = 0x18
	Vop1OpRndneF64         = 0x19
	Vop1OpFloorF64         = 0x1A
	Vop1OpFractF32         = 0x20
	Vop1OpTruncF32         = 0x21
	Vop1OpCeilF32          = 0x22
	Vop1OpRndneF32         = 0x23
	Vop1OpFloorF32         = 0x24
	Vop1OpExpF32           = 0x25
	Vop1OpLogClampF32      = 0x26
	Vop1OpLogF32           = 0x27
	Vop1OpRcpF32           = 0x2A
	Vop1OpRcpClampF32      = 0x2B
	Vop1OpRsqLegacyF32     = 0x2D
	Vop1OpRsqF32           = 0x2E
	Vop1OpRcpF64           = 0x2F
	Vop1OpRcpClampF64      = 0x30
	Vop1OpRsqF64           = 0x31
	Vop1OpRsqClampF64      = 0x32
	Vop1OpSqrtF32          = 0x33
	Vop1OpSqrtF64          = 0x34
	Vop1OpSinF32           = 0x35
	Vop1OpCosF32           = 0x36
	Vop1OpNotB32           = 0x37
	Vop1OpBfrevB32         = 0x38
	Vop1OpFfbhU32          = 0x39
	Vop1OpFfblB32          = 0x3A
	Vop1OpFfbhI32          = 0x3B
	Vop1OpFrexpExpI32F64   = 0x3C
	Vop1OpFrexpMantF64     = 0x3D
	Vop1OpFractF64         = 0x3E
	Vop1OpFrexpExpI32F32   = 0x3F
	Vop1OpFrexpMantF32     = 0x40
	Vop1OpClrexcp          = 0x41
	Vop1OpMovreldB32       = 0x42
	Vop1OpMovrelsB32       = 0x43
	Vop1OpMovrelsdB32      = 0x44
	Vop1OpLogLegacyF32     = 0x45
	Vop1OpExpLegacyF32     = 0x46
)

func (instr *Instruction) DecodeVOP1() {
	dw := instr.Dwords[0]
	instr.Details = &VectorDetails{
		Src0: dw & 0b1111_1111_1,
		Op:   (dw >> 9) & 0b1111_1111,
		Dst:  (dw >> 17) & 0b1111_1111,
	}
	if instr.DwordLen > 1 {
		instr.HasLiteral = true
		instr.Literal = instr.Dwords[1]
	}
}
