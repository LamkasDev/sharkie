package spirv

import (
	"fmt"

	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
	gcnSpec "github.com/LamkasDev/sharkie/cmd/structs/gcn/spec"
)

func emitEXP(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext) {
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
		vec := b.EmitCompositeConstruct(typeV4Float, comps[0], comps[1], comps[2], comps[3])
		b.EmitStore(outId, vec)
	}
}
