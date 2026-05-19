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
		case 0:
			return vk.FormatR8Unorm, 1
		case 1:
			return vk.FormatR8Snorm, 1
		case 2:
			return vk.FormatR8Uscaled, 1
		case 3:
			return vk.FormatR8Sscaled, 1
		case 4:
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
		case 0:
			return vk.FormatR8g8Unorm, 2
		case 1:
			return vk.FormatR8g8Snorm, 2
		case 2:
			return vk.FormatR8g8Uscaled, 2
		case 3:
			return vk.FormatR8g8Sscaled, 2
		case 4:
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
		case 0:
			return vk.FormatR8g8b8a8Unorm, 4
		case 1:
			return vk.FormatR8g8b8a8Snorm, 4
		case 2:
			return vk.FormatR8g8b8a8Uscaled, 4
		case 3:
			return vk.FormatR8g8b8a8Sscaled, 4
		case 4:
			return vk.FormatR8g8b8a8Uint, 4
		case 5:
			return vk.FormatR8g8b8a8Sint, 4
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
	}

	return vk.FormatUndefined, 0
}
