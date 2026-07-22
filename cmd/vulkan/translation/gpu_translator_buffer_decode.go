package translation

import (
	"encoding/binary"
	"math"
)

func DecodeGcnBufferFormat(src []byte, dataFormat uint8, numFormat uint8) [4]uint32 {
	var dst [4]uint32
	// Default W to 1.0f for floats, 1 for ints
	isFloat := numFormat == 0 || numFormat == 1 || numFormat == 7 || numFormat == 9
	if isFloat {
		dst[3] = math.Float32bits(1.0)
	} else {
		dst[3] = 1
	}

	switch dataFormat {
	case 13: // 32_32_32
		if len(src) < 12 {
			return dst
		}
		if numFormat == 7 { // Float
			dst[0] = binary.LittleEndian.Uint32(src[0:4])
			dst[1] = binary.LittleEndian.Uint32(src[4:8])
			dst[2] = binary.LittleEndian.Uint32(src[8:12])
		}
	case 11: // 32_32
		if len(src) < 8 {
			return dst
		}
		if numFormat == 7 { // Float
			dst[0] = binary.LittleEndian.Uint32(src[0:4])
			dst[1] = binary.LittleEndian.Uint32(src[4:8])
		}
	case 10: // 8_8_8_8
		if len(src) < 4 {
			return dst
		}
		if numFormat == 0 || numFormat == 1 {
			dst[0] = math.Float32bits(float32(src[0]) / 255.0)
			dst[1] = math.Float32bits(float32(src[1]) / 255.0)
			dst[2] = math.Float32bits(float32(src[2]) / 255.0)
			dst[3] = math.Float32bits(float32(src[3]) / 255.0)
		} else if numFormat == 4 || numFormat == 5 {
			dst[0] = uint32(src[0])
			dst[1] = uint32(src[1])
			dst[2] = uint32(src[2])
			dst[3] = uint32(src[3])
		}
	case 5: // 16_16
		if len(src) < 4 {
			return dst
		}
		if numFormat == 4 || numFormat == 5 {
			dst[0] = uint32(binary.LittleEndian.Uint16(src[0:2]))
			dst[1] = uint32(binary.LittleEndian.Uint16(src[2:4]))
		}
	}
	return dst
}
