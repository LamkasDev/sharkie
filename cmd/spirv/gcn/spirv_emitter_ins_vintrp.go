package gcn

import (
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	gcnSpec "github.com/LamkasDev/sharkie/cmd/structs/gcn/spec"
)

func EmitVINTRP(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*gcnSpec.VintrpDetails)
	typeFloat := ctx.GetId(BlockContextIdTypeFloat)
	typeUint := ctx.GetId(BlockContextIdTypeUint)

	// Load component from attribute vector.
	attribute := ctx.LoadPsInputParameter(b, details.Attr)
	value := b.EmitCompositeExtract(typeFloat, attribute, uint32(details.Chan))

	var res SpirvId
	switch details.Op {
	case gcnSpec.VintrpOpInterpP1F32, gcnSpec.VintrpOpInterpMovF32:
		// For P1 and MOV, we just load the interpolated value.
		// In SPIR-V, inputs are already interpolated at pixel center by default.
		res = value
	case gcnSpec.VintrpOpInterpP2F32:
		// P2 usually accumulates the result of P1.
		// Since we load the full interpolated value in P1, P2 is effectively a no-op / passthrough.
		res = b.EmitBitcast(typeFloat, ctx.GetGcnVgprId(b, details.Vdst))
	}

	// Store result to destination VGPR.
	ctx.StoreRegisterPointer(b, gcnSpec.OpVgpr0+details.Vdst, b.EmitBitcast(typeUint, res))
}
