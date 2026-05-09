package gcn

import "github.com/LamkasDev/sharkie/cmd/structs/gcn/spec"

func NewEncoding(dw uint32) spec.Encoding {
	top9 := (dw >> 23) & 0b111111111
	switch top9 {
	case 0b101111101:
		return spec.EncSOP1
	case 0b101111110:
		return spec.EncSOPC
	case 0b101111111:
		return spec.EncSOPP
	}

	top5 := dw >> 27
	if top5 == 0b11000 {
		return spec.EncSMRD
	}

	top6 := (dw >> 26) & 0b111111
	switch top6 {
	case 0b110010:
		return spec.EncVINTRP
	case 0b110100:
		return spec.EncVOP3
	case 0b110110:
		return spec.EncDS
	case 0b111110:
		return spec.EncEXP
	case 0b111000:
		return spec.EncMUBUF
	case 0b111010:
		return spec.EncMTBUF
	case 0b111100:
		return spec.EncMIMG
	}

	top7 := (dw >> 25) & 0b1111111
	switch top7 {
	case 0b0111111:
		return spec.EncVOP1
	case 0b0111110:
		return spec.EncVOPC
	}

	if (dw >> 28) == 0b1011 {
		return spec.EncSOPK
	}
	if (dw >> 30) == 0b10 {
		return spec.EncSOP2
	}
	if (dw >> 31) == 0b0 {
		return spec.EncVOP2
	}

	return spec.EncUnknown
}

func GetEncodingDwordLen(dw uint32) int {
	switch NewEncoding(dw) {
	case spec.EncVOP3, spec.EncEXP, spec.EncDS, spec.EncMUBUF, spec.EncMTBUF, spec.EncMIMG:
		return 2
	case spec.EncVOP1, spec.EncVOPC, spec.EncVOP2:
		if dw&0x1FF == 0xFF { // SRC0 == 0xFF
			return 2
		}
	case spec.EncSOP1:
		if dw&0xFF == 0xFF { // SSRC0 == 0xFF
			return 2
		}
	case spec.EncSOPC, spec.EncSOP2, spec.EncVINTRP:
		if dw&0xFF == 0xFF || (dw>>8)&0xFF == 0xFF { // SSRC0/SSRC1 == 0xFF
			return 2
		}
	}

	return 1
}
