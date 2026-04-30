package spec

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
	EncVOP2   // Vector ALU, 2 sources
	EncVOP1   // Vector ALU, 1 source
	EncVOPC   // Vector ALU compare
	EncVINTRP // Vector ALU interpolate
	EncVOP3   // Vector ALU 3-source (classic or scalar destination)

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
	EncVINTRP:  "VINTRP",
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
