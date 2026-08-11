package gcn

import (
	"fmt"

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

// TranslateGcnFormat maps GCN formats to Vulkan VkFormat and byte size.
func TranslateGcnFormat(dataFormat, numFormat uint8, compSwap uint32) (vk.Format, uint32) {
	if compSwap != 0 && dataFormat != gcn.GcnDataFormat8_8_8_8 &&
		dataFormat != gcn.GcnDataFormat10_10_10_2 && dataFormat != gcn.GcnDataFormat2_10_10_10 {
		panic(fmt.Sprintf("unhandled comp swap %d for format %d", compSwap, dataFormat))
	}

	switch dataFormat {
	case gcn.GcnDataFormatInvalid:
		switch numFormat {
		case gcn.GcnNumFormatUnorm:
			return vk.FormatUndefined, 0
		default:
			panic("invalid format")
		}
	case gcn.GcnDataFormat8:
		switch numFormat {
		case gcn.GcnNumFormatUnorm, gcn.GcnNumFormatUbnorm:
			return vk.FormatR8Unorm, 1
		case gcn.GcnNumFormatSnorm, gcn.GcnNumFormatSnormOgl:
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
		case gcn.GcnNumFormatSnorm, gcn.GcnNumFormatSnormOgl:
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
		case gcn.GcnNumFormatSnorm, gcn.GcnNumFormatSnormOgl:
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
		case gcn.GcnNumFormatSnorm, gcn.GcnNumFormatSnormOgl:
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
	case gcn.GcnDataFormat10_11_11, gcn.GcnDataFormat11_11_10:
		switch numFormat {
		case gcn.GcnNumFormatUnorm:
			// TODO: not sure about this one.
			return vk.FormatB10g11r11UfloatPack32, 4
		case gcn.GcnNumFormatSfloat:
			return vk.FormatB10g11r11UfloatPack32, 4
		}
	case gcn.GcnDataFormat10_10_10_2, gcn.GcnDataFormat2_10_10_10:
		if compSwap == 1 { // SWAP_ALT (BGRA)
			switch numFormat {
			case gcn.GcnNumFormatUnorm:
				return vk.FormatA2r10g10b10UnormPack32, 4
			case gcn.GcnNumFormatSnorm, gcn.GcnNumFormatSnormOgl:
				return vk.FormatA2r10g10b10SnormPack32, 4
			case gcn.GcnNumFormatUint:
				return vk.FormatA2r10g10b10UintPack32, 4
			case gcn.GcnNumFormatSint:
				return vk.FormatA2r10g10b10SintPack32, 4
			}
			break
		}
		switch numFormat {
		case gcn.GcnNumFormatUnorm:
			return vk.FormatA2b10g10r10UnormPack32, 4
		case gcn.GcnNumFormatUint:
			return vk.FormatA2b10g10r10UintPack32, 4
		case gcn.GcnNumFormatSnorm, gcn.GcnNumFormatSnormOgl:
			return vk.FormatA2b10g10r10SnormPack32, 4
		case gcn.GcnNumFormatSint:
			return vk.FormatA2b10g10r10SintPack32, 4
		}
	case gcn.GcnDataFormat8_8_8_8:
		if compSwap == 1 { // SWAP_ALT (BGRA)
			switch numFormat {
			case gcn.GcnNumFormatUnorm:
				return vk.FormatB8g8r8a8Unorm, 4
			// TODO: not sure about this one.
			case gcn.GcnNumFormatSnorm, gcn.GcnNumFormatSnormOgl:
				return vk.FormatR8g8b8a8Snorm, 4
			case gcn.GcnNumFormatUint:
				return vk.FormatB8g8r8a8Uint, 4
			case gcn.GcnNumFormatSint:
				return vk.FormatB8g8r8a8Sint, 4
			case gcn.GcnNumFormatSrgb:
				return vk.FormatB8g8r8a8Srgb, 4
			}
			break
		}
		switch numFormat {
		case gcn.GcnNumFormatUnorm, gcn.GcnNumFormatUbnorm:
			return vk.FormatR8g8b8a8Unorm, 4
		case gcn.GcnNumFormatSnorm, gcn.GcnNumFormatSnormOgl:
			return vk.FormatR8g8b8a8Snorm, 4
		case gcn.GcnNumFormatUscaled, gcn.GcnNumFormatUbscaled:
			return vk.FormatR8g8b8a8Uscaled, 4
		case gcn.GcnNumFormatSscaled:
			return vk.FormatR8g8b8a8Sscaled, 4
		case gcn.GcnNumFormatUint, gcn.GcnNumFormatUbint:
			return vk.FormatR8g8b8a8Uint, 4
		case gcn.GcnNumFormatSint:
			return vk.FormatR8g8b8a8Sint, 4
		case gcn.GcnNumFormatSfloat:
			// TODO: not sure about this one.
			return vk.FormatR16g16b16a16Sfloat, 8
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
		case gcn.GcnNumFormatSnorm, gcn.GcnNumFormatSnormOgl:
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
		case gcn.GcnNumFormatSnorm, gcn.GcnNumFormatSnormOgl:
			return vk.FormatBc4SnormBlock, 8
		}
	case gcn.GcnDataFormatBC5:
		switch numFormat {
		case gcn.GcnNumFormatUnorm:
			return vk.FormatBc5UnormBlock, 16
		case gcn.GcnNumFormatSnorm, gcn.GcnNumFormatSnormOgl:
			return vk.FormatBc5SnormBlock, 16
		}
	case gcn.GcnDataFormatBC6:
		switch numFormat {
		case gcn.GcnNumFormatUnorm, gcn.GcnNumFormatSfloat:
			return vk.FormatBc6hUfloatBlock, 16
		case gcn.GcnNumFormatSnorm, gcn.GcnNumFormatSnormOgl:
			return vk.FormatBc6hSfloatBlock, 16
		}
	case gcn.GcnDataFormatBC7:
		switch numFormat {
		case gcn.GcnNumFormatUnorm:
			return vk.FormatBc7UnormBlock, 16
		case gcn.GcnNumFormatSrgb:
			return vk.FormatBc7SrgbBlock, 16
		}
	case gcn.GcnDataFormatFmask8_1:
		switch numFormat {
		case gcn.GcnNumFormatUint:
			return vk.FormatR8Uint, 1
		}
	}

	panic(fmt.Sprintf("unhandled data format %d number format %d", dataFormat, numFormat))
}

// TranslateComponentMapping maps GCN DataFormat and NumFormat to Vulkan VkFormat and byte size.
func TranslateComponentMapping(dataFormat, numFormat, dstSelX, dstSelY, dstSelZ, dstSelW uint8) vk.ComponentMapping {
	components := vk.ComponentMapping{
		R: TranslateDstSelToVkSwizzle(dstSelX),
		G: TranslateDstSelToVkSwizzle(dstSelY),
		B: TranslateDstSelToVkSwizzle(dstSelZ),
		A: TranslateDstSelToVkSwizzle(dstSelW),
	}

	// Address LSB/MSB mismatch between Vulkan/GCN formats.
	switch dataFormat {
	case gcn.GcnDataFormat5_6_5, gcn.GcnDataFormat1_5_5_5, gcn.GcnDataFormat11_11_10:
		components = vk.ComponentMapping{
			R: components.B,
			G: components.G,
			B: components.R,
			A: components.A,
		}
	case gcn.GcnDataFormat10_10_10_2:
		components = vk.ComponentMapping{
			R: components.A,
			G: components.B,
			B: components.G,
			A: components.R,
		}
	case gcn.GcnDataFormat4_4_4_4:
		components = vk.ComponentMapping{
			R: components.G,
			G: components.B,
			B: components.A,
			A: components.R,
		}
	case gcn.GcnDataFormat8, gcn.GcnDataFormat8_8, gcn.GcnDataFormat16_16:
		if components.R == vk.ComponentSwizzleA {
			components.R = vk.ComponentSwizzleR
		}
		if components.G == vk.ComponentSwizzleA {
			components.G = vk.ComponentSwizzleR
		}
		if components.B == vk.ComponentSwizzleA {
			components.B = vk.ComponentSwizzleR
		}
		if components.A == vk.ComponentSwizzleA {
			components.A = vk.ComponentSwizzleR
		}
	}

	// Intentional swap because we're creating AMD GCN's 2_10_10_10 through Vulkan's 10_10_10_2.
	switch dataFormat {
	case gcn.GcnDataFormat2_10_10_10:
		components = vk.ComponentMapping{
			R: components.A,
			G: components.B,
			B: components.G,
			A: components.R,
		}
	// TODO: not sure about this one.
	case gcn.GcnDataFormat11_11_10:
		components = vk.ComponentMapping{
			R: components.A,
			G: components.G,
			B: components.B,
		}
	}

	return components
}
