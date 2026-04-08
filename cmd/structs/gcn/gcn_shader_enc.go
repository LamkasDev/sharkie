package gcn

type Encoding uint8

const (
	EncUnknown Encoding = iota

	// [Scalar ALU Operations] 5.1 SALU Instruction Formats.
	EncSOP2 // Scalar ALU, 2 sources
	EncSOPK // Scalar ALU with inline 16-bit constant
	EncSOP1 // Scalar ALU, 1 source
	EncSOPC // Scalar ALU compare
	EncSOPP // Scalar ALU program flow

	// [Vector ALU Operations] 6.1 Microcode Encodings.
	EncVOP2  // Vector ALU, 2 sources
	EncVOP1  // Vector ALU, 1 source
	EncVOPC  // Vector ALU compare
	EncINTRP // Vector ALU interpolate
	EncVOP3  // Vector ALU 3-source (classic or scalar destination)

	// [Scalar Memory Operations] 7.1 Microcode Encoding
	EncSMRD // Scalar Memory read

	// [Vector Memory Operations] 8.1 Vector Memory Buffer Instructions.
	EncMTBUF // Memory Typed Buffer
	EncMUBUF // Memory Untyped Buffer

	// [Vector Memory Operations] 8.2 Vector Memory (VM) Image Instructions.
	EncMIMG // Memory Image

	// [Flat Memory Instructions] 9.1 Flat Memory Instructions.

	// [Data Share Operations] 10.3 LDS Access.
	EncDS // Global / Local Data Share

	// [Exporting Pixel Color and Vertex Shader Parameters] 11.1 Microcode Encoding.
	EncEXP // Export
)

var EncodingNames = map[Encoding]string{
	EncUnknown: "UNKNOWN",
	EncSOP2:    "SOP2",
	EncSOPK:    "SOPK",
	EncSOP1:    "SOP1",
	EncSOPC:    "SOPC",
	EncSOPP:    "SOPP",
	EncVOP2:    "VOP2",
	EncVOP1:    "VOP1",
	EncVOPC:    "VOPC",
	EncINTRP:   "INTRP",
	EncVOP3:    "VOP3",
	EncSMRD:    "SMRD",
	EncMTBUF:   "MTBUF",
	EncMUBUF:   "MUBUF",
	EncMIMG:    "MIMG",
	EncDS:      "DS",
	EncEXP:     "EXP",
}

func (e Encoding) String() string {
	return EncodingNames[e]
}

func NewEncoding(dw uint32) Encoding {
	top9 := (dw >> 23) & 0b111111111
	switch top9 {
	case 0b101111101:
		return EncSOP1
	case 0b101111110:
		return EncSOPC
	case 0b101111111:
		return EncSOPP
	}

	top5 := dw >> 27
	if top5 == 0b11000 || top5 == 0b11001 {
		return EncSMRD
	}

	top6 := (dw >> 26) & 0b111111
	switch top6 {
	case 0b110100:
		return EncVOP3
	case 0b110110:
		return EncDS
	case 0b111110:
		return EncEXP
	case 0b111000:
		return EncMUBUF
	case 0b111010:
		return EncMTBUF
	case 0b111100:
		return EncMIMG
	}

	top7 := (dw >> 25) & 0b1111111
	switch top7 {
	case 0b0111111:
		return EncVOP1
	case 0b0111110:
		return EncVOPC
	}

	if (dw >> 28) == 0b1011 {
		return EncSOPK
	}
	if (dw >> 30) == 0b10 {
		return EncSOP2
	}
	if (dw >> 31) == 0b0 {
		return EncVOP2
	}

	return EncUnknown
}

func GetEncodingDwordLen(dw uint32) int {
	switch NewEncoding(dw) {
	case EncVOP3, EncEXP, EncDS, EncMUBUF, EncMTBUF, EncMIMG:
		return 2
	case EncVOP1, EncVOPC, EncVOP2:
		if dw&0x1FF == 0xFF { // SRC0 == 0xFF
			return 2
		}
	case EncSOP1:
		if dw&0xFF == 0xFF { // SSRC0 == 0xFF
			return 2
		}
	case EncSOPC, EncSOP2:
		if dw&0xFF == 0xFF || (dw>>8)&0xFF == 0xFF { // SSRC0/SSRC1 == 0xFF
			return 2
		}
	}

	return 1
}
