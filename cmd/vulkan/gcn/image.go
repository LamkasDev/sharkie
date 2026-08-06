package gcn

import (
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gcn"
	vk "github.com/goki/vulkan"
)

func TranslateBorderColorType(colorType uint8) vk.BorderColor {
	switch colorType {
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

func TranslateClampMode(mode uint8) vk.SamplerAddressMode {
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

func TranslateFilter(mode uint8) vk.Filter {
	switch mode {
	case 0: // point
		return vk.FilterNearest
	case 1: // linear
		return vk.FilterLinear
	default:
		return vk.FilterNearest
	}
}

func TranslateMipmapMode(mode uint8) vk.SamplerMipmapMode {
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

func TranslateDstSelToVkSwizzle(sel uint8) vk.ComponentSwizzle {
	switch sel {
	case 0:
		return vk.ComponentSwizzleZero
	case 1:
		return vk.ComponentSwizzleOne
	case 4:
		return vk.ComponentSwizzleR
	case 5:
		return vk.ComponentSwizzleG
	case 6:
		return vk.ComponentSwizzleB
	case 7:
		return vk.ComponentSwizzleA
	default:
		return vk.ComponentSwizzleIdentity
	}
}

// TranslateGcnFormat maps GCN DataFormat and NumFormat to Vulkan VkFormat and byte size.
func TranslateGcnFormat(dataFormat, numFormat uint8) (vk.Format, uint32) {
	switch dataFormat {
	case gcn.GcnDataFormat8:
		switch numFormat {
		case gcn.GcnNumFormatUnorm, gcn.GcnNumFormatUbnorm:
			return vk.FormatR8Unorm, 1
		case gcn.GcnNumFormatSnorm:
			return vk.FormatR8Snorm, 1
		case gcn.GcnNumFormatUscaled, gcn.GcnNumFormatUbscaled:
			return vk.FormatR8Uscaled, 1
		case gcn.GcnNumFormatSscaled:
			return vk.FormatR8Sscaled, 1
		case gcn.GcnNumFormatUint, gcn.GcnNumFormatUbint:
			return vk.FormatR8Uint, 1
		case gcn.GcnNumFormatSint:
			return vk.FormatR8Sint, 1
		}
	case gcn.GcnDataFormat16:
		switch numFormat {
		case gcn.GcnNumFormatUnorm:
			return vk.FormatR16Unorm, 2
		case gcn.GcnNumFormatSnorm:
			return vk.FormatR16Snorm, 2
		case gcn.GcnNumFormatUscaled:
			return vk.FormatR16Uscaled, 2
		case gcn.GcnNumFormatSscaled:
			return vk.FormatR16Sscaled, 2
		case gcn.GcnNumFormatUint:
			return vk.FormatR16Uint, 2
		case gcn.GcnNumFormatSint:
			return vk.FormatR16Sint, 2
		case gcn.GcnNumFormatSfloat:
			return vk.FormatR16Sfloat, 2
		}
	case gcn.GcnDataFormat8_8:
		switch numFormat {
		case gcn.GcnNumFormatUnorm, gcn.GcnNumFormatUbnorm:
			return vk.FormatR8g8Unorm, 2
		case gcn.GcnNumFormatSnorm:
			return vk.FormatR8g8Snorm, 2
		case gcn.GcnNumFormatUscaled, gcn.GcnNumFormatUbscaled:
			return vk.FormatR8g8Uscaled, 2
		case gcn.GcnNumFormatSscaled:
			return vk.FormatR8g8Sscaled, 2
		case gcn.GcnNumFormatUint, gcn.GcnNumFormatUbint:
			return vk.FormatR8g8Uint, 2
		case gcn.GcnNumFormatSint:
			return vk.FormatR8g8Sint, 2
		}
	case gcn.GcnDataFormat32:
		switch numFormat {
		case gcn.GcnNumFormatUint:
			return vk.FormatR32Uint, 4
		case gcn.GcnNumFormatSint:
			return vk.FormatR32Sint, 4
		case gcn.GcnNumFormatSfloat:
			return vk.FormatR32Sfloat, 4
		}
	case gcn.GcnDataFormat16_16:
		switch numFormat {
		case gcn.GcnNumFormatUnorm:
			return vk.FormatR16g16Unorm, 4
		case gcn.GcnNumFormatSnorm:
			return vk.FormatR16g16Snorm, 4
		case gcn.GcnNumFormatUscaled:
			return vk.FormatR16g16Uscaled, 4
		case gcn.GcnNumFormatSscaled:
			return vk.FormatR16g16Sscaled, 4
		case gcn.GcnNumFormatUint:
			return vk.FormatR16g16Uint, 4
		case gcn.GcnNumFormatSint:
			return vk.FormatR16g16Sint, 4
		case gcn.GcnNumFormatSfloat:
			return vk.FormatR16g16Sfloat, 4
		}
	case gcn.GcnDataFormat10_11_11:
		if numFormat == gcn.GcnNumFormatSfloat {
			return vk.FormatB10g11r11UfloatPack32, 4
		}
	case gcn.GcnDataFormat11_11_10:
		if numFormat == gcn.GcnNumFormatSfloat {
			// TODO: fix this.
			return vk.FormatB10g11r11UfloatPack32, 4
		}
	case gcn.GcnDataFormat10_10_10_2:
		switch numFormat {
		case gcn.GcnNumFormatUnorm:
			return vk.FormatA2b10g10r10UnormPack32, 4
		case gcn.GcnNumFormatUint:
			return vk.FormatA2b10g10r10UintPack32, 4
		}
	case gcn.GcnDataFormat8_8_8_8:
		switch numFormat {
		case gcn.GcnNumFormatUnorm, gcn.GcnNumFormatUbnorm:
			return vk.FormatR8g8b8a8Unorm, 4
		case gcn.GcnNumFormatSnorm:
			return vk.FormatR8g8b8a8Snorm, 4
		case gcn.GcnNumFormatUscaled, gcn.GcnNumFormatUbscaled:
			return vk.FormatR8g8b8a8Uscaled, 4
		case gcn.GcnNumFormatSscaled:
			return vk.FormatR8g8b8a8Sscaled, 4
		case gcn.GcnNumFormatUint, gcn.GcnNumFormatUbint:
			return vk.FormatR8g8b8a8Uint, 4
		case gcn.GcnNumFormatSint:
			return vk.FormatR8g8b8a8Sint, 4
		case gcn.GcnNumFormatSrgb:
			return vk.FormatR8g8b8a8Srgb, 4
		}
	case gcn.GcnDataFormat32_32:
		switch numFormat {
		case gcn.GcnNumFormatUint:
			return vk.FormatR32g32Uint, 8
		case gcn.GcnNumFormatSint:
			return vk.FormatR32g32Sint, 8
		case gcn.GcnNumFormatSfloat:
			return vk.FormatR32g32Sfloat, 8
		}
	case gcn.GcnDataFormat16_16_16_16:
		switch numFormat {
		case gcn.GcnNumFormatUnorm:
			return vk.FormatR16g16b16a16Unorm, 8
		case gcn.GcnNumFormatSnorm:
			return vk.FormatR16g16b16a16Snorm, 8
		case gcn.GcnNumFormatUscaled:
			return vk.FormatR16g16b16a16Uscaled, 8
		case gcn.GcnNumFormatSscaled:
			return vk.FormatR16g16b16a16Sscaled, 8
		case gcn.GcnNumFormatUint:
			return vk.FormatR16g16b16a16Uint, 8
		case gcn.GcnNumFormatSint:
			return vk.FormatR16g16b16a16Sint, 8
		case gcn.GcnNumFormatSfloat:
			return vk.FormatR16g16b16a16Sfloat, 8
		}
	case gcn.GcnDataFormat32_32_32:
		switch numFormat {
		case gcn.GcnNumFormatUint:
			return vk.FormatR32g32b32Uint, 12
		case gcn.GcnNumFormatSint:
			return vk.FormatR32g32b32Sint, 12
		case gcn.GcnNumFormatSfloat:
			return vk.FormatR32g32b32Sfloat, 12
		}
	case gcn.GcnDataFormat32_32_32_32:
		switch numFormat {
		case gcn.GcnNumFormatUint:
			return vk.FormatR32g32b32a32Uint, 16
		case gcn.GcnNumFormatSint:
			return vk.FormatR32g32b32a32Sint, 16
		case gcn.GcnNumFormatSfloat:
			return vk.FormatR32g32b32a32Sfloat, 16
		}
	case gcn.GcnDataFormat5_6_5:
		switch numFormat {
		case gcn.GcnNumFormatUnorm, gcn.GcnNumFormatSrgb:
			return vk.FormatR5g6b5UnormPack16, 2
		}
	case gcn.GcnDataFormat1_5_5_5:
		switch numFormat {
		case gcn.GcnNumFormatUnorm:
			return vk.FormatA1r5g5b5UnormPack16, 2
		}
	case gcn.GcnDataFormat5_5_5_1:
		switch numFormat {
		case gcn.GcnNumFormatUnorm:
			return vk.FormatR5g5b5a1UnormPack16, 2
		}
	case gcn.GcnDataFormat4_4_4_4:
		switch numFormat {
		case gcn.GcnNumFormatUnorm:
			return vk.FormatR4g4b4a4UnormPack16, 2
		}
	case gcn.GcnDataFormatBC1:
		switch numFormat {
		case gcn.GcnNumFormatUnorm:
			return vk.FormatBc1RgbaUnormBlock, 8
		case gcn.GcnNumFormatSrgb:
			return vk.FormatBc1RgbaSrgbBlock, 8
		}
	case gcn.GcnDataFormatBC2:
		switch numFormat {
		case gcn.GcnNumFormatUnorm:
			return vk.FormatBc2UnormBlock, 16
		case gcn.GcnNumFormatSrgb:
			return vk.FormatBc2SrgbBlock, 16
		}
	case gcn.GcnDataFormatBC3:
		switch numFormat {
		case gcn.GcnNumFormatUnorm:
			return vk.FormatBc3UnormBlock, 16
		case gcn.GcnNumFormatSrgb:
			return vk.FormatBc3SrgbBlock, 16
		}
	case gcn.GcnDataFormatBC4:
		switch numFormat {
		case gcn.GcnNumFormatUnorm:
			return vk.FormatBc4UnormBlock, 8
		case gcn.GcnNumFormatSnorm:
			return vk.FormatBc4SnormBlock, 8
		}
	case gcn.GcnDataFormatBC5:
		switch numFormat {
		case gcn.GcnNumFormatUnorm:
			return vk.FormatBc5UnormBlock, 16
		case gcn.GcnNumFormatSnorm:
			return vk.FormatBc5SnormBlock, 16
		}
	case gcn.GcnDataFormatBC6:
		switch numFormat {
		case 0, 7:
			return vk.FormatBc6hUfloatBlock, 16
		case gcn.GcnNumFormatSnorm:
			return vk.FormatBc6hSfloatBlock, 16
		}
	case gcn.GcnDataFormatBC7:
		switch numFormat {
		case gcn.GcnNumFormatUnorm:
			return vk.FormatBc7UnormBlock, 16
		case gcn.GcnNumFormatSrgb:
			return vk.FormatBc7SrgbBlock, 16
		}
	}

	return vk.FormatUndefined, 0
}

// GetBytesPerPixel returns the size of a single pixel based on the GCN DataFormat.
func GetBytesPerPixel(format uint8) uint32 {
	switch format {
	case gcn.GcnDataFormat8:
		return 1
	case gcn.GcnDataFormat16, gcn.GcnDataFormat8_8, gcn.GcnDataFormat32, gcn.GcnDataFormat16_16, gcn.GcnDataFormat10_11_11, 25:
		return 2
	case gcn.GcnDataFormat10_10_10_2, gcn.GcnDataFormat8_8_8_8, gcn.GcnDataFormat32_32, 26, 34:
		return 4
	case gcn.GcnDataFormat16_16_16_16, gcn.GcnDataFormat32_32_32, gcn.GcnDataFormat32_32_32_32, gcn.GcnDataFormatBC1, gcn.GcnDataFormatBC4:
		return 8
	case gcn.GcnDataFormatReserved_15, gcn.GcnDataFormatBC2, gcn.GcnDataFormatBC3, gcn.GcnDataFormatBC5, gcn.GcnDataFormatBC6, gcn.GcnDataFormatBC7:
		return 16
	default:
		return 4 // Fallback
	}
}
