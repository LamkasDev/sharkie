package vulkan

import (
	vk "github.com/goki/vulkan"
)

func translateBorderColorType(borderColor uint8) vk.BorderColor {
	switch borderColor {
	case 0: // opaque-black
		return vk.BorderColorIntOpaqueBlack
	case 1: // transparent-black
		return vk.BorderColorIntTransparentBlack
	case 2: // white
		return vk.BorderColorIntOpaqueWhite
	case 3: // TODO: use border color ptr
		return vk.BorderColorIntTransparentBlack
	default:
		return vk.BorderColorIntTransparentBlack
	}
}

func translateClampMode(mode uint8) vk.SamplerAddressMode {
	switch mode {
	case 0: // wrap
		return vk.SamplerAddressModeRepeat
	case 1: // mirror
		return vk.SamplerAddressModeMirroredRepeat
	case 2: // clamp to edge
		return vk.SamplerAddressModeClampToEdge
	case 3: // mirror once
		return vk.SamplerAddressModeMirrorClampToEdge
	case 4: // clamp to border
		return vk.SamplerAddressModeClampToBorder
	default:
		return vk.SamplerAddressModeRepeat
	}
}

func translateFilter(mode uint8) vk.Filter {
	switch mode {
	case 0: // point
		return vk.FilterNearest
	case 1: // linear
		return vk.FilterLinear
	default:
		return vk.FilterNearest
	}
}

func translateMipmapMode(mode uint8) vk.SamplerMipmapMode {
	switch mode {
	case 0: // none
		return vk.SamplerMipmapModeNearest
	case 1: // point
		return vk.SamplerMipmapModeNearest
	case 2: // linear
		return vk.SamplerMipmapModeLinear
	default:
		return vk.SamplerMipmapModeNearest
	}
}

// TranslateGcnFormat maps GCN DataFormat and NumFormat to Vulkan VkFormat and byte size.
func TranslateGcnFormat(dataFormat, numFormat uint8) (vk.Format, uint32) {
	switch dataFormat {
	case 1: // 8-bit (1 byte)
		switch numFormat {
		case 0, 10:
			return vk.FormatR8Unorm, 1
		case 1:
			return vk.FormatR8Snorm, 1
		case 2, 13:
			return vk.FormatR8Uscaled, 1
		case 3:
			return vk.FormatR8Sscaled, 1
		case 4, 12:
			return vk.FormatR8Uint, 1
		case 5:
			return vk.FormatR8Sint, 1
		}
	case 2: // 16-bit (2 bytes)
		switch numFormat {
		case 0:
			return vk.FormatR16Unorm, 2
		case 1:
			return vk.FormatR16Snorm, 2
		case 2:
			return vk.FormatR16Uscaled, 2
		case 3:
			return vk.FormatR16Sscaled, 2
		case 4:
			return vk.FormatR16Uint, 2
		case 5:
			return vk.FormatR16Sint, 2
		case 7:
			return vk.FormatR16Sfloat, 2
		}
	case 3: // 8_8 (2 bytes)
		switch numFormat {
		case 0, 10:
			return vk.FormatR8g8Unorm, 2
		case 1:
			return vk.FormatR8g8Snorm, 2
		case 2, 13:
			return vk.FormatR8g8Uscaled, 2
		case 3:
			return vk.FormatR8g8Sscaled, 2
		case 4, 12:
			return vk.FormatR8g8Uint, 2
		case 5:
			return vk.FormatR8g8Sint, 2
		}
	case 4: // 32-bit (4 bytes)
		switch numFormat {
		case 4:
			return vk.FormatR32Uint, 4
		case 5:
			return vk.FormatR32Sint, 4
		case 7:
			return vk.FormatR32Sfloat, 4
		}
	case 5: // 16_16 (4 bytes)
		switch numFormat {
		case 0:
			return vk.FormatR16g16Unorm, 4
		case 1:
			return vk.FormatR16g16Snorm, 4
		case 2:
			return vk.FormatR16g16Uscaled, 4
		case 3:
			return vk.FormatR16g16Sscaled, 4
		case 4:
			return vk.FormatR16g16Uint, 4
		case 5:
			return vk.FormatR16g16Sint, 4
		case 7:
			return vk.FormatR16g16Sfloat, 4
		}
	case 6: // 10_11_11 (4 bytes)
		if numFormat == 7 {
			return vk.FormatB10g11r11UfloatPack32, 4
		}
	case 8: // 10_10_10_2 (4 bytes)
		switch numFormat {
		case 0:
			return vk.FormatA2b10g10r10UnormPack32, 4
		case 4:
			return vk.FormatA2b10g10r10UintPack32, 4
		}
	case 10: // 8_8_8_8 (4 bytes)
		switch numFormat {
		case 0, 10:
			return vk.FormatR8g8b8a8Unorm, 4
		case 1:
			return vk.FormatR8g8b8a8Snorm, 4
		case 2, 13:
			return vk.FormatR8g8b8a8Uscaled, 4
		case 3:
			return vk.FormatR8g8b8a8Sscaled, 4
		case 4, 12:
			return vk.FormatR8g8b8a8Uint, 4
		case 5:
			return vk.FormatR8g8b8a8Sint, 4
		case 9:
			return vk.FormatR8g8b8a8Srgb, 4
		}
	case 11: // 32_32 (8 bytes)
		switch numFormat {
		case 4:
			return vk.FormatR32g32Uint, 8
		case 5:
			return vk.FormatR32g32Sint, 8
		case 7:
			return vk.FormatR32g32Sfloat, 8
		}
	case 12: // 16_16_16_16 (8 bytes)
		switch numFormat {
		case 0:
			return vk.FormatR16g16b16a16Unorm, 8
		case 1:
			return vk.FormatR16g16b16a16Snorm, 8
		case 2:
			return vk.FormatR16g16b16a16Uscaled, 8
		case 3:
			return vk.FormatR16g16b16a16Sscaled, 8
		case 4:
			return vk.FormatR16g16b16a16Uint, 8
		case 5:
			return vk.FormatR16g16b16a16Sint, 8
		case 7:
			return vk.FormatR16g16b16a16Sfloat, 8
		}
	case 13: // 32_32_32 (12 bytes)
		switch numFormat {
		case 4:
			return vk.FormatR32g32b32Uint, 12
		case 5:
			return vk.FormatR32g32b32Sint, 12
		case 7:
			return vk.FormatR32g32b32Sfloat, 12
		}
	case 14: // 32_32_32_32 (16 bytes)
		switch numFormat {
		case 4:
			return vk.FormatR32g32b32a32Uint, 16
		case 5:
			return vk.FormatR32g32b32a32Sint, 16
		case 7:
			return vk.FormatR32g32b32a32Sfloat, 16
		}
	case 16: // 5_6_5
		switch numFormat {
		case 0, 9:
			return vk.FormatR5g6b5UnormPack16, 2
		}
	case 17: // 1_5_5_5
		switch numFormat {
		case 0:
			return vk.FormatA1r5g5b5UnormPack16, 2
		}
	case 18: // 5_5_5_1
		switch numFormat {
		case 0:
			return vk.FormatR5g5b5a1UnormPack16, 2
		}
	case 19: // 4_4_4_4
		switch numFormat {
		case 0:
			return vk.FormatR4g4b4a4UnormPack16, 2
		}
	case 35: // BC1
		switch numFormat {
		case 0:
			return vk.FormatBc1RgbaUnormBlock, 8
		case 9:
			return vk.FormatBc1RgbaSrgbBlock, 8
		}
	case 36: // BC2
		switch numFormat {
		case 0:
			return vk.FormatBc2UnormBlock, 16
		case 9:
			return vk.FormatBc2SrgbBlock, 16
		}
	case 37: // BC3
		switch numFormat {
		case 0:
			return vk.FormatBc3UnormBlock, 16
		case 9:
			return vk.FormatBc3SrgbBlock, 16
		}
	case 38: // BC4
		switch numFormat {
		case 0:
			return vk.FormatBc4UnormBlock, 8
		case 1:
			return vk.FormatBc4SnormBlock, 8
		}
	case 39: // BC5
		switch numFormat {
		case 0:
			return vk.FormatBc5UnormBlock, 16
		case 1:
			return vk.FormatBc5SnormBlock, 16
		}
	case 40: // BC6
		switch numFormat {
		case 0, 7:
			return vk.FormatBc6hUfloatBlock, 16
		case 1:
			return vk.FormatBc6hSfloatBlock, 16
		}
	case 41: // BC7
		switch numFormat {
		case 0:
			return vk.FormatBc7UnormBlock, 16
		case 9:
			return vk.FormatBc7SrgbBlock, 16
		}
	}

	return vk.FormatUndefined, 0
}
