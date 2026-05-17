package vulkan

import (
	vk "github.com/goki/vulkan"
)

func translateLogicOp(rop3 uint32) vk.LogicOp {
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

func translateColorFormat(format uint32, numberType uint32, compSwap uint32) vk.Format {
	// TODO: fix this.
	return vk.FormatR8g8b8a8Unorm

	switch format {
	case 1: // COLOR_8
		switch numberType {
		case 0:
			return vk.FormatR8Unorm
		case 1:
			return vk.FormatR8Snorm
		case 4:
			return vk.FormatR8Uint
		case 5:
			return vk.FormatR8Sint
		case 6:
			return vk.FormatR8Srgb
		}
	case 2: // COLOR_16
		switch numberType {
		case 0:
			return vk.FormatR16Unorm
		case 1:
			return vk.FormatR16Snorm
		case 4:
			return vk.FormatR16Uint
		case 5:
			return vk.FormatR16Sint
		case 7:
			return vk.FormatR16Sfloat
		}
	case 3: // COLOR_8_8
		switch numberType {
		case 0:
			return vk.FormatR8g8Unorm
		case 1:
			return vk.FormatR8g8Snorm
		case 4:
			return vk.FormatR8g8Uint
		case 5:
			return vk.FormatR8g8Sint
		case 6:
			return vk.FormatR8g8Srgb
		}
	case 4: // COLOR_32
		switch numberType {
		case 4:
			return vk.FormatR32Uint
		case 5:
			return vk.FormatR32Sint
		case 7:
			return vk.FormatR32Sfloat
		}
	case 5: // COLOR_16_16
		switch numberType {
		case 0:
			return vk.FormatR16g16Unorm
		case 1:
			return vk.FormatR16g16Snorm
		case 4:
			return vk.FormatR16g16Uint
		case 5:
			return vk.FormatR16g16Sint
		case 7:
			return vk.FormatR16g16Sfloat
		}
	case 10: // COLOR_8_8_8_8
		if compSwap == 1 { // SWAP_ALT (BGRA)
			switch numberType {
			case 0:
				return vk.FormatB8g8r8a8Unorm
			case 1:
				return vk.FormatB8g8r8a8Snorm
			case 4:
				return vk.FormatB8g8r8a8Uint
			case 5:
				return vk.FormatB8g8r8a8Sint
			case 6:
				return vk.FormatB8g8r8a8Srgb
			}
		}
		switch numberType {
		case 0:
			return vk.FormatR8g8b8a8Unorm
		case 1:
			return vk.FormatR8g8b8a8Snorm
		case 4:
			return vk.FormatR8g8b8a8Uint
		case 5:
			return vk.FormatR8g8b8a8Sint
		case 6:
			return vk.FormatR8g8b8a8Srgb
		}
	case 11: // COLOR_32_32
		switch numberType {
		case 4:
			return vk.FormatR32g32Uint
		case 5:
			return vk.FormatR32g32Sint
		case 7:
			return vk.FormatR32g32Sfloat
		}
	case 12: // COLOR_16_16_16_16
		switch numberType {
		case 0:
			return vk.FormatR16g16b16a16Unorm
		case 1:
			return vk.FormatR16g16b16a16Snorm
		case 4:
			return vk.FormatR16g16b16a16Uint
		case 5:
			return vk.FormatR16g16b16a16Sint
		case 7:
			return vk.FormatR16g16b16a16Sfloat
		}
	case 14: // COLOR_32_32_32_32
		switch numberType {
		case 4:
			return vk.FormatR32g32b32a32Uint
		case 5:
			return vk.FormatR32g32b32a32Sint
		case 7:
			return vk.FormatR32g32b32a32Sfloat
		}
	case 16: // COLOR_5_6_5
		return vk.FormatR5g6b5UnormPack16
	case 17: // COLOR_1_5_5_5
		return vk.FormatA1r5g5b5UnormPack16
	case 18: // COLOR_5_5_5_1
		return vk.FormatR5g5b5a1UnormPack16
	case 19: // COLOR_4_4_4_4
		return vk.FormatR4g4b4a4UnormPack16
	case 20: // COLOR_8_24
		return vk.FormatD24UnormS8Uint
	case 21: // COLOR_24_8
		return vk.FormatD24UnormS8Uint
	case 22: // COLOR_X24_8_32_FLOAT
		return vk.FormatD32SfloatS8Uint
	}
	return vk.FormatR8g8b8a8Unorm
}
