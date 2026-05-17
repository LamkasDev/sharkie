package vulkan

import (
	vk "github.com/goki/vulkan"
	"go101.org/nstd"
)

func translateBlendControl(blendControl uint32, colorWriteMask uint32, blendBypass bool) vk.PipelineColorBlendAttachmentState {
	srcColor := blendControl & 0x1f
	dstColor := (blendControl >> 8) & 0x1f
	opColor := (blendControl >> 5) & 0x7

	srcAlpha := srcColor
	dstAlpha := dstColor
	opAlpha := opColor

	if (blendControl>>29)&1 == 1 { // Separate alpha blend
		srcAlpha = (blendControl >> 16) & 0x1f
		dstAlpha = (blendControl >> 24) & 0x1f
		opAlpha = (blendControl >> 21) & 0x7
	}

	blendEnable := (blendControl>>30)&1 == 1
	if blendBypass {
		blendEnable = false
	}

	return vk.PipelineColorBlendAttachmentState{
		ColorWriteMask:      vk.ColorComponentFlags(colorWriteMask),
		BlendEnable:         vk.Bool32(nstd.Btoi(blendEnable)),
		SrcColorBlendFactor: translateBlendFactor(srcColor),
		DstColorBlendFactor: translateBlendFactor(dstColor),
		ColorBlendOp:        translateBlendOp(opColor),
		SrcAlphaBlendFactor: translateBlendFactor(srcAlpha),
		DstAlphaBlendFactor: translateBlendFactor(dstAlpha),
		AlphaBlendOp:        translateBlendOp(opAlpha),
	}
}

func translateBlendFactor(factor uint32) vk.BlendFactor {
	switch factor {
	case 0: // BLEND_ZERO
		return vk.BlendFactorZero
	case 1: // BLEND_ONE
		return vk.BlendFactorOne
	case 2: // BLEND_SRC_COLOR
		return vk.BlendFactorSrcColor
	case 3: // BLEND_ONE_MINUS_SRC_COLOR
		return vk.BlendFactorOneMinusSrcColor
	case 4: // BLEND_SRC_ALPHA
		return vk.BlendFactorSrcAlpha
	case 5: // BLEND_ONE_MINUS_SRC_ALPHA
		return vk.BlendFactorOneMinusSrcAlpha
	case 6: // BLEND_DST_ALPHA
		return vk.BlendFactorDstAlpha
	case 7: // BLEND_ONE_MINUS_DST_ALPHA
		return vk.BlendFactorOneMinusDstAlpha
	case 8: // BLEND_DST_COLOR
		return vk.BlendFactorDstColor
	case 9: // BLEND_ONE_MINUS_DST_COLOR
		return vk.BlendFactorOneMinusDstColor
	case 10: // BLEND_SRC_ALPHA_SATURATE
		return vk.BlendFactorSrcAlphaSaturate
	case 13: // BLEND_CONSTANT_COLOR
		return vk.BlendFactorConstantColor
	case 14: // BLEND_ONE_MINUS_CONSTANT_COLOR
		return vk.BlendFactorOneMinusConstantColor
	case 15: // BLEND_SRC1_COLOR
		return vk.BlendFactorSrc1Color
	case 16: // BLEND_INV_SRC1_COLOR
		return vk.BlendFactorOneMinusSrc1Color
	case 17: // BLEND_SRC1_ALPHA
		return vk.BlendFactorSrc1Alpha
	case 18: // BLEND_INV_SRC1_ALPHA
		return vk.BlendFactorOneMinusSrc1Alpha
	case 19: // BLEND_CONSTANT_ALPHA
		return vk.BlendFactorConstantAlpha
	case 20: // BLEND_ONE_MINUS_CONSTANT_ALPHA
		return vk.BlendFactorOneMinusConstantAlpha
	default:
		return vk.BlendFactorOne
	}
}

func translateBlendOp(op uint32) vk.BlendOp {
	switch op {
	case 0: // COMB_DST_PLUS_SRC
		return vk.BlendOpAdd
	case 1: // COMB_SRC_MINUS_DST
		return vk.BlendOpSubtract
	case 2: // COMB_MIN_DST_SRC
		return vk.BlendOpMin
	case 3: // COMB_MAX_DST_SRC
		return vk.BlendOpMax
	case 4: // COMB_DST_MINUS_SRC
		return vk.BlendOpReverseSubtract
	default:
		return vk.BlendOpAdd
	}
}
