package gcn

import (
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/reg"
	vk "github.com/goki/vulkan"
	"go101.org/nstd"
)

func TranslateDepthControl(depthControl reg.DbDepthControl, stencilControl reg.DbStencilControl, stencilRefMask reg.DbStencilrefmask, stencilRefMaskBf reg.DbStencilrefmaskBf) vk.PipelineDepthStencilStateCreateInfo {
	backfaceEnable := depthControl.BackfaceEnable()

	frontState := vk.StencilOpState{
		FailOp:      TranslateStencilOp(stencilControl.Stencilfail()),
		PassOp:      TranslateStencilOp(stencilControl.Stencilzpass()),
		DepthFailOp: TranslateStencilOp(stencilControl.Stencilzfail()),
		CompareOp:   TranslateCompareOp(depthControl.Stencilfunc()),
		CompareMask: stencilRefMask.Stencilmask(),
		WriteMask:   stencilRefMask.Stencilwritemask(),
		Reference:   stencilRefMask.Stenciltestval(),
	}

	var backState vk.StencilOpState
	if backfaceEnable {
		backState = vk.StencilOpState{
			FailOp:      TranslateStencilOp(stencilControl.StencilfailBf()),
			PassOp:      TranslateStencilOp(stencilControl.StencilzpassBf()),
			DepthFailOp: TranslateStencilOp(stencilControl.StencilzfailBf()),
			CompareOp:   TranslateCompareOp(depthControl.StencilfuncBf()),
			CompareMask: stencilRefMaskBf.StencilmaskBf(),
			WriteMask:   stencilRefMaskBf.StencilwritemaskBf(),
			Reference:   stencilRefMaskBf.StenciltestvalBf(),
		}
	} else {
		backState = frontState
	}

	zfunc := depthControl.Zfunc()
	depthWriteEnable := depthControl.ZWriteEnable()
	if zfunc == 7 {
		depthWriteEnable = false
	}

	return vk.PipelineDepthStencilStateCreateInfo{
		SType:                 vk.StructureTypePipelineDepthStencilStateCreateInfo,
		DepthTestEnable:       vk.Bool32(nstd.Btoi(depthControl.ZEnable())),
		DepthWriteEnable:      vk.Bool32(nstd.Btoi(depthWriteEnable)),
		DepthCompareOp:        TranslateCompareOp(zfunc),
		DepthBoundsTestEnable: vk.Bool32(nstd.Btoi(depthControl.DepthBoundsEnable())),
		StencilTestEnable:     vk.Bool32(nstd.Btoi(depthControl.StencilEnable())),
		Front:                 frontState,
		Back:                  backState,
	}
}

func TranslateCompareOp(op uint32) vk.CompareOp {
	switch op {
	case 0: // FRAG_NEVER / REF_NEVER
		return vk.CompareOpNever
	case 1: // FRAG_LESS / REF_LESS
		return vk.CompareOpLessOrEqual
	case 2: // FRAG_EQUAL / REF_EQUAL
		return vk.CompareOpEqual
	case 3: // FRAG_LEQUAL / REF_LEQUAL
		return vk.CompareOpLessOrEqual
	case 4: // FRAG_GREATER / REF_GREATER
		return vk.CompareOpGreaterOrEqual
	case 5: // FRAG_NOTEQUAL / REF_NOTEQUAL
		return vk.CompareOpNotEqual
	case 6: // FRAG_GEQUAL / REF_GEQUAL
		return vk.CompareOpGreaterOrEqual
	case 7: // FRAG_ALWAYS / REF_ALWAYS
		return vk.CompareOpAlways
	default:
		return vk.CompareOpAlways
	}
}

func TranslateStencilOp(op uint32) vk.StencilOp {
	switch op {
	case 0: // STENCIL_KEEP
		return vk.StencilOpKeep
	case 1: // STENCIL_ZERO
		return vk.StencilOpZero
	case 2: // STENCIL_ONES
		return vk.StencilOpReplace // No Oned, Replace with 0xFF? Vulkan StencilOpReplace uses reference.
	case 3: // STENCIL_REPLACE_TEST
		return vk.StencilOpReplace
	case 4: // STENCIL_REPLACE_OP
		return vk.StencilOpReplace
	case 5: // STENCIL_ADD_CLAMP
		return vk.StencilOpIncrementAndClamp
	case 6: // STENCIL_SUB_CLAMP
		return vk.StencilOpDecrementAndClamp
	case 7: // STENCIL_INVERT
		return vk.StencilOpInvert
	case 8: // STENCIL_ADD_WRAP
		return vk.StencilOpIncrementAndWrap
	case 9: // STENCIL_SUB_WRAP
		return vk.StencilOpDecrementAndWrap
	default:
		return vk.StencilOpKeep
	}
}

func TranslateGcnDepthFormat(format uint32, formatProperties map[vk.Format]vk.FormatProperties) vk.Format {
	var requested vk.Format
	switch format {
	case 1: // Z_16: 16-bit UNORM depth surface.
		requested = vk.FormatD16UnormS8Uint
	case 2: // Z_24: 24-bit UNORM depth surface.
		requested = vk.FormatD24UnormS8Uint
	case 3: // Z_32_FLOAT: 32-bit FLOAT depth surface.
		requested = vk.FormatD32SfloatS8Uint
	default:
		return vk.FormatUndefined
	}

	// Check if the physical device supports optimal tiling for this format as both a depth-stencil attachment and a sampled image.
	required := vk.FormatFeatureFlags(vk.FormatFeatureDepthStencilAttachmentBit | vk.FormatFeatureSampledImageBit)
	if (formatProperties[requested].OptimalTilingFeatures & required) == required {
		return requested
	}

	// Fallback to D32_SFLOAT_S8_UINT which is universally supported by all Vulkan implementations supporting stencil.
	return vk.FormatD32SfloatS8Uint
}
