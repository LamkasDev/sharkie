package gcn

import (
	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
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
	case gcnSpec.VintrpOpInterpP1F32:
		// P1 only computes intermediate interpolation (P10 * I + P0) which P2 finishes.
		// Since Vulkan inputs are already interpolated by hardware, P1 is a no-op to avoid
		// clobbering registers (e.g. barycentrics) needed by subsequent instructions.
		return
	case gcnSpec.VintrpOpInterpP2F32, gcnSpec.VintrpOpInterpMovF32:
		res = value
	}

	// Store result to destination VGPR.
	ctx.StoreRegisterPointer(b, gcnSpec.OpVgpr0+details.Vdst, b.EmitBitcast(typeUint, res))
}
