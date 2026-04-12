package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitEXP(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	details := instr.Details.(*ExpDetails)
	var comps [4]uint32
	idFloat := ctx.GetId(BlockContextIdTypeFloat)

	if details.Compr {
		// Compressed: 16-bit pairs in 2 VGPRs.
		// EN[0] enables VSRC0 (R,G), EN[2] enables VSRC1 (B,A).
		if details.En&0b0001 != 0 {
			val0 := ctx.GetOperandValue(b, details.VSrcs[0]+OpVgpr0, 0)
			v01 := b.EmitExtInst(ctx.GetId(BlockContextIdTypeV2Float), ctx.GetId(BlockContextIdGlsl), SpvGlslOpUnpackHalf2x16, val0)
			comps[0] = b.EmitCompositeExtract(idFloat, v01, 0)
			comps[1] = b.EmitCompositeExtract(idFloat, v01, 1)
		}
		if details.En&0b0100 != 0 {
			val1 := ctx.GetOperandValue(b, details.VSrcs[1]+OpVgpr0, 0)
			v23 := b.EmitExtInst(ctx.GetId(BlockContextIdTypeV2Float), ctx.GetId(BlockContextIdGlsl), SpvGlslOpUnpackHalf2x16, val1)
			comps[2] = b.EmitCompositeExtract(idFloat, v23, 0)
			comps[3] = b.EmitCompositeExtract(idFloat, v23, 1)
		}
	} else {
		// Uncompressed: 32-bit components in 4 VGPRs.
		for i := range comps {
			if details.En&(1<<i) != 0 {
				comps[i] = ctx.GetOperandFloatValue(b, details.VSrcs[i]+OpVgpr0, 0)
			}
		}
	}

	// Determine output ID based on target.
	var outId uint32
	switch {
	case details.Target <= 7: // MRT 0..7
		outId = ctx.GetId(BlockContextIdColorOut0 + BlockContextId(details.Target))
	case details.Target == 8: // Z
		outId = ctx.GetId(BlockContextIdFragDepthOut)
	case details.Target == 9: // Null
		return
	case details.Target >= 12 && details.Target <= 15: // Position 0..3
		outId = ctx.GetId(BlockContextIdPosOut)
	case details.Target >= 32 && details.Target <= 63: // Param 0..31
		outId = ctx.GetId(BlockContextIdParamOut0 + BlockContextId(details.Target-32))
	default:
		panic(fmt.Sprintf("unknown export target %d", details.Target))
	}

	if outId == 0 {
		return // Not declared for this stage.
	}

	// Store to target.
	if details.Target == 8 {
		// Depth is a single float.
		if details.En&1 != 0 {
			b.EmitStore(outId, comps[0])
		}
	} else {
		// Other targets are vec4.
		idV4Float := ctx.GetId(BlockContextIdTypeV4Float)
		idZeroF := ctx.GetConstId(ConstIdxFloat0)
		for i := range comps {
			if comps[i] == 0 {
				comps[i] = idZeroF
			}
		}
		vec := b.EmitCompositeConstruct(idV4Float, comps[0], comps[1], comps[2], comps[3])
		b.EmitStore(outId, vec)
	}
}
