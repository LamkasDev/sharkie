package gcn

import (
	"fmt"

	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
)

func EmitEXP(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*gcnSpec.ExpDetails)
	var comps [4]SpirvId
	typeFloat := ctx.GetId(BlockContextIdTypeFloat)
	typeV4Float := ctx.GetId(BlockContextIdTypeV4Float)
	typeBool := ctx.GetId(BlockContextIdTypeBool)

	if details.Compr {
		// Compressed: 16-bit pairs in 2 VGPRs.
		// EN[0] enables VSRC0 (R,G), EN[2] enables VSRC1 (B,A).
		if details.En&0b0001 != 0 {
			val0 := ctx.GetOperandValue(b, details.VSrcs[0]+gcnSpec.OpVgpr0, 0)
			v01 := b.EmitExtInst(ctx.GetId(BlockContextIdTypeV2Float), ctx.GetId(BlockContextIdGlsl), spec.SpvGlslOpUnpackHalf2x16, val0)
			comps[0] = b.EmitCompositeExtract(typeFloat, v01, 0)
			comps[1] = b.EmitCompositeExtract(typeFloat, v01, 1)
		}
		if details.En&0b0100 != 0 {
			val1 := ctx.GetOperandValue(b, details.VSrcs[1]+gcnSpec.OpVgpr0, 0)
			v23 := b.EmitExtInst(ctx.GetId(BlockContextIdTypeV2Float), ctx.GetId(BlockContextIdGlsl), spec.SpvGlslOpUnpackHalf2x16, val1)
			comps[2] = b.EmitCompositeExtract(typeFloat, v23, 0)
			comps[3] = b.EmitCompositeExtract(typeFloat, v23, 1)
		}
	} else {
		// Uncompressed: 32-bit components in 4 VGPRs.
		for i := range comps {
			if details.En&(1<<i) != 0 {
				comps[i] = ctx.GetOperandFloatValue(b, details.VSrcs[i]+gcnSpec.OpVgpr0, 0)
			}
		}
	}

	// Determine output ID based on target.
	var outId SpirvId
	switch {
	case details.Target <= 7: // MRT 0..7
		outId = ctx.GetId(BlockContextIdColorOut0 + SpirvId(details.Target))
	case details.Target == 8: // Z
		outId = ctx.GetId(BlockContextIdFragDepthOut)
	case details.Target == 9: // Null
		return
	case details.Target >= 12 && details.Target <= 15: // Position 0..3
		outId = ctx.GetId(BlockContextIdPosOut)
	case details.Target >= 32 && details.Target <= 63: // Param 0..31
		outId = ctx.GetId(BlockContextIdParamOut0 + SpirvId(details.Target-32))
	default:
		panic(fmt.Sprintf("unknown export target %d", details.Target))
	}
	if outId == 0 {
		return // Not declared for this stage.
	}

	// Store to target.
	idZeroF := ctx.GetConstId(ConstIdFloat0)
	idOneF := ctx.GetConstId(ConstIdFloat1)
	switch {
	case details.Target == 8:
		// Depth is a single float.
		if details.En&1 != 0 {
			b.EmitStore(outId, comps[0])
		}
	default:
		// Other targets are vec4.
		for i := range comps {
			if comps[i] == 0 {
				if details.Target <= 7 && i == 3 { // MRT alpha
					comps[i] = idOneF
				} else {
					comps[i] = idZeroF
				}
			} else if details.Target >= 12 && details.Target <= 15 { // Position 0..3
				isNaN := b.EmitFUnordNotEqual(typeBool, comps[i], comps[i])
				defaultValue := idZeroF
				if i == 3 {
					defaultValue = idOneF
				}
				comps[i] = b.EmitSelect(typeFloat, isNaN, defaultValue, comps[i])
			}
		}
		if details.Target >= 12 && details.Target <= 15 {
			vteControl := ctx.LoadPushConstantValue(b, PushConstantVteControl)
			vtxXyFmt := ctx.TestMask(b, vteControl, 1<<8)
			vtxZFmt := ctx.TestMask(b, vteControl, 1<<9)
			vtxW0Fmt := ctx.TestMask(b, vteControl, 1<<10)

			// Prevent W=0.0 to avoid +Inf generation.
			isZero := b.EmitFUnordEqual(typeBool, comps[3], idZeroF)
			safeW := b.EmitSelect(typeFloat, isZero, idOneF, comps[3])

			// For VTX_W0_FMT:
			// 1 = Shader exported 1/W 	-> Vulkan needs W = 1.0 / shader_W.
			// 0 = Shader exported W 	-> Vulkan needs W = shader_W.
			wVulkan := b.EmitSelect(typeFloat, vtxW0Fmt, b.EmitFDiv(typeFloat, idOneF, safeW), safeW)

			// If FMT is 1, X/Y/Z are already multiplied by 1/W, so multiply by W to undo Vulkan's divide.
			comps[0] = b.EmitSelect(typeFloat, vtxXyFmt, b.EmitFMul(typeFloat, comps[0], wVulkan), comps[0])
			comps[1] = b.EmitSelect(typeFloat, vtxXyFmt, b.EmitFMul(typeFloat, comps[1], wVulkan), comps[1])
			comps[2] = b.EmitSelect(typeFloat, vtxZFmt, b.EmitFMul(typeFloat, comps[2], wVulkan), comps[2])
			comps[3] = wVulkan

			// Adjust Z based on DX_CLIP_SPACE_DEF.
			clipControl := ctx.LoadPushConstantValue(b, PushConstantClipControl)
			dxClipSpaceDef := ctx.TestMask(b, clipControl, 1<<19)

			// OpenGL: Z_vul = (Z_gl + W) / 2.0
			zGl := comps[2]
			zGlPlusW := b.EmitFAdd(typeFloat, zGl, wVulkan)
			halfW := b.EmitFDiv(typeFloat, zGlPlusW, ctx.GetConstId(ConstIdFloat2))
			comps[2] = b.EmitSelect(typeFloat, dxClipSpaceDef, zGl, halfW)
		}

		vec := b.EmitCompositeConstruct(typeV4Float, comps[0], comps[1], comps[2], comps[3])
		b.EmitStore(outId, vec)
	}
}
