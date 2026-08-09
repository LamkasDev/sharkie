package gcn

import (
	"fmt"
	"math"

	"github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/reg"
	vk "github.com/goki/vulkan"
	"github.com/x448/float16"
)

func TranslateLogicOp(rop3 uint32) vk.LogicOp {
	switch rop3 {
	case 0x00: // BLACKNESS
		return vk.LogicOpClear
	case 0x05:
		return vk.LogicOpAndReverse
	case 0x0A:
		return vk.LogicOpAndInverted
	case 0x0F:
		return vk.LogicOpNoOp // Not exactly, but close to some ROPs
	case 0x11: // NOTSRCERASE
		return vk.LogicOpNor
	case 0x22:
		return vk.LogicOpAndReverse
	case 0x33: // NOTSRCCOPY
		return vk.LogicOpCopyInverted
	case 0x44: // SRCERASE
		return vk.LogicOpAndInverted
	case 0x50:
		return vk.LogicOpNoOp
	case 0x55: // DSTINVERT
		return vk.LogicOpInvert
	case 0x5A: // PATINVERT
		return vk.LogicOpXor
	case 0x5F:
		return vk.LogicOpNoOp
	case 0x66: // SRCINVERT
		return vk.LogicOpXor
	case 0x77:
		return vk.LogicOpNand
	case 0x88: // SRCAND
		return vk.LogicOpAnd
	case 0x99:
		return vk.LogicOpEquivalent
	case 0xA0:
		return vk.LogicOpAnd
	case 0xA5:
		return vk.LogicOpXor
	case 0xAA:
		return vk.LogicOpNoOp
	case 0xAF:
		return vk.LogicOpNoOp
	case 0xBB: // MERGEPAINT
		return vk.LogicOpOrReverse
	case 0xCC: // SRCCOPY
		return vk.LogicOpCopy
	case 0xDD:
		return vk.LogicOpOrInverted
	case 0xEE: // SRCPAINT
		return vk.LogicOpOr
	case 0xF0: // PATCOPY
		return vk.LogicOpCopy
	case 0xF5:
		return vk.LogicOpNoOp
	case 0xFA:
		return vk.LogicOpNoOp
	case 0xFF: // WHITENESS
		return vk.LogicOpSet
	default:
		return vk.LogicOpCopy
	}
}

func ColorBufferPitch(pitch reg.CbColorPitch) uint32 {
	return (pitch.TileMax() + 1) * 8
}

func ColorBufferHeight(pitch reg.CbColorPitch, sliceReg uint32) uint32 {
	sliceTileMax := sliceReg & 0x3FFFFF // TILE_MAX for slice
	totalTiles := sliceTileMax + 1
	p := ColorBufferPitch(pitch)
	if p == 0 {
		return 1080
	}

	return (totalTiles * 64) / p
}

func TranslateClearColor(word0 uint32, word1 uint32, dataFormat uint32, numFormat uint32, compSwap uint32) []float32 {
	if compSwap != 0 && dataFormat != gcn.GcnDataFormat2_10_10_10 &&
		dataFormat != gcn.GcnDataFormat8_8_8_8 && dataFormat != gcn.GcnDataFormat16_16_16_16 {
		panic(fmt.Sprintf("unhandled comp swap %d for format %d", compSwap, dataFormat))
	}

	var r, g, b, a float32 = 0.0, 0.0, 0.0, 1.0
	switch dataFormat {
	case gcn.GcnDataFormat8:
		switch numFormat {
		case gcn.GcnNumFormatUnorm:
			r = float32(word0&0xFF) / 255.0
		default:
			panic("unhandled")
		}
	case gcn.GcnDataFormat16:
		switch numFormat {
		case gcn.GcnNumFormatUnorm:
			r = float32(word0&0xFFFF) / 255.0
		default:
			panic("unhandled")
		}
	case gcn.GcnDataFormat32:
		switch numFormat {
		case gcn.GcnNumFormatSfloat:
			r = math.Float32frombits(word0)
		default:
			panic("unhandled")
		}
	case gcn.GcnDataFormat16_16:
		switch numFormat {
		case gcn.GcnNumFormatSfloat:
			r = float32(float16.Frombits(uint16(word0 & 0xFFFF)))
			g = float32(float16.Frombits(uint16((word0 >> 16) & 0xFFFF)))
		default:
			panic("unhandled")
		}
	case gcn.GcnDataFormat11_11_10:
		switch numFormat {
		case gcn.GcnNumFormatSfloat:
			// TODO: fix this.
			_ = "todo"
		default:
			panic("unhandled")
		}
	case gcn.GcnDataFormat2_10_10_10:
		switch numFormat {
		case gcn.GcnNumFormatUnorm:
			r = float32(word0&0x3) / 255.0
			g = float32((word0>>2)&0x3FF) / 255.0
			b = float32((word0>>12)&0x3FF) / 255.0
			a = float32((word0>>22)&0x3FF) / 255.0
		default:
			panic("unhandled")
		}
		if compSwap == 1 {
			r, b = b, r
		}
	case gcn.GcnDataFormat8_8_8_8:
		switch numFormat {
		case gcn.GcnNumFormatUnorm, gcn.GcnNumFormatSnormOgl:
			r = float32(word0&0xFF) / 255.0
			g = float32((word0>>8)&0xFF) / 255.0
			b = float32((word0>>16)&0xFF) / 255.0
			a = float32((word0>>24)&0xFF) / 255.0
		case gcn.GcnNumFormatUint, gcn.GcnNumFormatSint:
			r = float32(word0 & 0xFF)
			g = float32((word0 >> 8) & 0xFF)
			b = float32((word0 >> 16) & 0xFF)
			a = float32((word0 >> 24) & 0xFF)
		default:
			panic("unhandled")
		}
		if compSwap == 1 {
			r, b = b, r
		}
	case gcn.GcnDataFormat16_16_16_16:
		switch numFormat {
		case gcn.GcnNumFormatSfloat:
			r = float32(float16.Frombits(uint16(word0 & 0xFFFF)))
			g = float32(float16.Frombits(uint16((word0 >> 16) & 0xFFFF)))
			b = float32(float16.Frombits(uint16(word1 & 0xFFFF)))
			a = float32(float16.Frombits(uint16((word1 >> 16) & 0xFFFF)))
		default:
			panic("unhandled")
		}
		if compSwap == 1 {
			r, b = b, r
		}
	default:
		panic("unhandled")
	}

	return []float32{r, g, b, a}
}
