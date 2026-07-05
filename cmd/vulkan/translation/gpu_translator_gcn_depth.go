package translation

import (
	vk "github.com/goki/vulkan"
	"go101.org/nstd"
)

func translateDepthControl(depthControl uint32, stencilControl uint32, stencilRefMask uint32, stencilRefMaskBf uint32) vk.PipelineDepthStencilStateCreateInfo {
	backfaceEnable := (depthControl>>7)&1 == 1

	frontState := vk.StencilOpState{
		FailOp:      translateStencilOp(stencilControl & 0xf),
		PassOp:      translateStencilOp((stencilControl >> 4) & 0xf),
		DepthFailOp: translateStencilOp((stencilControl >> 8) & 0xf),
		CompareOp:   translateCompareOp((depthControl >> 8) & 0x7),
		CompareMask: (stencilRefMask >> 8) & 0xff,
		WriteMask:   (stencilRefMask >> 16) & 0xff,
		Reference:   stencilRefMask & 0xff,
	}

	var backState vk.StencilOpState
	if backfaceEnable {
		backState = vk.StencilOpState{
			FailOp:      translateStencilOp((stencilControl >> 12) & 0xf),
			PassOp:      translateStencilOp((stencilControl >> 16) & 0xf),
			DepthFailOp: translateStencilOp((stencilControl >> 20) & 0xf),
			CompareOp:   translateCompareOp((depthControl >> 20) & 0x7),
			CompareMask: (stencilRefMaskBf >> 8) & 0xff,
			WriteMask:   (stencilRefMaskBf >> 16) & 0xff,
			Reference:   stencilRefMaskBf & 0xff,
		}
	} else {
		backState = frontState
	}

	// When ZFUNC is ALWAYS (7), disable depth writes unconditionally.
	// Writing depth from ZFUNC=ALWAYS draws provides no occlusion benefit
	// and, with depth-clearing to 1.0, any geometry that maps to depth < 1
	// will corrupt the depth buffer for subsequent LESS comparisons.
	zfunc := (depthControl >> 4) & 0x7
	depthWriteEnable := (depthControl>>2)&1 == 1
	if zfunc == 7 {
		depthWriteEnable = false
	}

	return vk.PipelineDepthStencilStateCreateInfo{
		SType:                 vk.StructureTypePipelineDepthStencilStateCreateInfo,
		DepthTestEnable:       vk.Bool32(nstd.Btoi((depthControl>>1)&1 == 1)),
		DepthWriteEnable:      vk.Bool32(nstd.Btoi(depthWriteEnable)),
		DepthCompareOp:        translateCompareOp(zfunc),
		DepthBoundsTestEnable: vk.Bool32(nstd.Btoi((depthControl>>3)&1 == 1)),
		StencilTestEnable:     vk.Bool32(nstd.Btoi((depthControl>>0)&1 == 1)),
		Front:                 frontState,
		Back:                  backState,
	}
}

func translateCompareOp(op uint32) vk.CompareOp {
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

func translateStencilOp(op uint32) vk.StencilOp {
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

func (t *GpuTranslator) TranslateGcnDepthFormat(format uint32) vk.Format {
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
	var props vk.FormatProperties
	vk.GetPhysicalDeviceFormatProperties(t.handles.PhysicalDevice, requested, &props)
	props.Deref()
	required := vk.FormatFeatureFlags(vk.FormatFeatureDepthStencilAttachmentBit | vk.FormatFeatureSampledImageBit)
	if (props.OptimalTilingFeatures & required) == required {
		return requested
	}

	// Fallback to D32_SFLOAT_S8_UINT which is universally supported by all Vulkan implementations supporting stencil.
	return vk.FormatD32SfloatS8Uint
}
