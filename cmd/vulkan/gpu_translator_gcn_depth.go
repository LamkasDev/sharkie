package vulkan

import (
	vk "github.com/goki/vulkan"
	"go101.org/nstd"
)

func translateDepthControl(depthControl uint32, stencilControl uint32) vk.PipelineDepthStencilStateCreateInfo {
	return vk.PipelineDepthStencilStateCreateInfo{
		SType:                 vk.StructureTypePipelineDepthStencilStateCreateInfo,
		DepthTestEnable:       vk.Bool32(nstd.Btoi((depthControl>>1)&1 == 1)),
		DepthWriteEnable:      vk.Bool32(nstd.Btoi((depthControl>>2)&1 == 1)),
		DepthCompareOp:        translateCompareOp((depthControl >> 4) & 0x7),
		DepthBoundsTestEnable: vk.Bool32(nstd.Btoi((depthControl>>3)&1 == 1)),
		StencilTestEnable:     vk.Bool32(nstd.Btoi((depthControl>>0)&1 == 1)),
		Front: vk.StencilOpState{
			FailOp:      translateStencilOp(stencilControl & 0xf),
			PassOp:      translateStencilOp((stencilControl >> 4) & 0xf),
			DepthFailOp: translateStencilOp((stencilControl >> 8) & 0xf),
			CompareOp:   translateCompareOp((depthControl >> 8) & 0x7),
		},
		Back: vk.StencilOpState{
			FailOp:      translateStencilOp((stencilControl >> 12) & 0xf),
			PassOp:      translateStencilOp((stencilControl >> 16) & 0xf),
			DepthFailOp: translateStencilOp((stencilControl >> 20) & 0xf),
			CompareOp:   translateCompareOp((depthControl >> 20) & 0x7),
		},
	}
}

func translateCompareOp(op uint32) vk.CompareOp {
	switch op {
	case 0: // FRAG_NEVER / REF_NEVER
		return vk.CompareOpNever
	case 1: // FRAG_LESS / REF_LESS
		return vk.CompareOpLess
	case 2: // FRAG_EQUAL / REF_EQUAL
		return vk.CompareOpEqual
	case 3: // FRAG_LEQUAL / REF_LEQUAL
		return vk.CompareOpLessOrEqual
	case 4: // FRAG_GREATER / REF_GREATER
		return vk.CompareOpGreater
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

func TranslateGcnDepthFormat(format uint32) vk.Format {
	switch format {
	case 1: // Z_16: 16-bit UNORM depth surface.
		return vk.FormatD16Unorm
	case 3: // Z_32_FLOAT: 32-bit FLOAT depth surface.
		return vk.FormatD32Sfloat
	default:
		return vk.FormatUndefined
	}
}
