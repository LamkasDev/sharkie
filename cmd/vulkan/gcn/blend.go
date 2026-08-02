package gcn

import (
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/reg"
	vk "github.com/goki/vulkan"
	"go101.org/nstd"
)

func TranslateBlendControl(blendControl reg.CbBlendControl, colorWriteMask reg.CbTargetMask, blendBypass bool) vk.PipelineColorBlendAttachmentState {
	srcColor := blendControl.ColorSrcblend()
	dstColor := blendControl.ColorDestblend()
	opColor := blendControl.ColorCombFcn()

	srcAlpha := srcColor
	dstAlpha := dstColor
	opAlpha := opColor

	if blendControl.SeparateAlphaBlend() {
		srcAlpha = blendControl.AlphaSrcblend()
		dstAlpha = blendControl.AlphaDestblend()
		opAlpha = blendControl.AlphaCombFcn()
	}

	blendEnable := blendControl.Enable()
	if blendBypass {
		blendEnable = false
	}

	return vk.PipelineColorBlendAttachmentState{
		ColorWriteMask:      0, // Set later in CreateBlendAttachments
		BlendEnable:         vk.Bool32(nstd.Btoi(blendEnable)),
		SrcColorBlendFactor: TranslateBlendFactor(srcColor),
		DstColorBlendFactor: TranslateBlendFactor(dstColor),
		ColorBlendOp:        TranslateBlendOp(opColor),
		SrcAlphaBlendFactor: TranslateBlendFactor(srcAlpha),
		DstAlphaBlendFactor: TranslateBlendFactor(dstAlpha),
		AlphaBlendOp:        TranslateBlendOp(opAlpha),
	}
}

func CreateBlendAttachments(baseAttachment vk.PipelineColorBlendAttachmentState, targetMask reg.CbTargetMask, shaderMask reg.CbShaderMask, colFormat reg.SpiShaderColFormat, colorControl reg.CbColorControl) []vk.PipelineColorBlendAttachmentState {
	attachments := make([]vk.PipelineColorBlendAttachmentState, 8)
	for i := 0; i < 8; i++ {
		attachments[i] = baseAttachment

		var enabled bool
		var format uint32
		var writeMask uint32
		switch i {
		case 0:
			enabled = targetMask.Target0Enable() != 0
			format = colFormat.Col0ExportFormat()
			writeMask = shaderMask.Output0Enable()
		case 1:
			enabled = targetMask.Target1Enable() != 0
			format = colFormat.Col1ExportFormat()
			writeMask = shaderMask.Output1Enable()
		case 2:
			enabled = targetMask.Target2Enable() != 0
			format = colFormat.Col2ExportFormat()
			writeMask = shaderMask.Output2Enable()
		case 3:
			enabled = targetMask.Target3Enable() != 0
			format = colFormat.Col3ExportFormat()
			writeMask = shaderMask.Output3Enable()
		case 4:
			enabled = targetMask.Target4Enable() != 0
			format = colFormat.Col4ExportFormat()
			writeMask = shaderMask.Output4Enable()
		case 5:
			enabled = targetMask.Target5Enable() != 0
			format = colFormat.Col5ExportFormat()
			writeMask = shaderMask.Output5Enable()
		case 6:
			enabled = targetMask.Target6Enable() != 0
			format = colFormat.Col6ExportFormat()
			writeMask = shaderMask.Output6Enable()
		case 7:
			enabled = targetMask.Target7Enable() != 0
			format = colFormat.Col7ExportFormat()
			writeMask = shaderMask.Output7Enable()
		}

		if enabled && format != 0 && colorControl.Mode() != 0 /* CB_DISABLE */ {
			attachments[i].ColorWriteMask = vk.ColorComponentFlags(writeMask)
		} else {
			attachments[i].ColorWriteMask = 0
		}
	}
	return attachments
}

func TranslateBlendFactor(factor uint32) vk.BlendFactor {
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

func TranslateBlendOp(op uint32) vk.BlendOp {
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
