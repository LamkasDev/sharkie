package gcn

import (
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
	gcnSpec "github.com/LamkasDev/sharkie/cmd/structs/gcn/spec"
)

func EmitVINTRP(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*gcnSpec.VintrpDetails)
	typeFloat := ctx.GetId(BlockContextIdTypeFloat)
	typeUint := ctx.GetId(BlockContextIdTypeUint)

	// Get attribute input variable.
	attrId := ctx.GetId(BlockContextIdParamIn0 + SpirvId(details.Attr))
	var res SpirvId

	if attrId != 0 {
		// Load component from attribute vector.
		ptr := b.EmitAccessChain(b.EmitTypePointer(spec.SpvStorageInput, typeFloat), attrId, ctx.GetConstId(SpirvId(details.Chan)))
		val := b.EmitLoad(typeFloat, ptr)

		switch details.Op {
		case gcnSpec.VintrpOpInterpP1F32, gcnSpec.VintrpOpInterpMovF32:
			// For P1 and MOV, we just load the interpolated value.
			// In SPIR-V, inputs are already interpolated at pixel center by default.
			res = val
		case gcnSpec.VintrpOpInterpP2F32:
			// P2 usually accumulates the result of P1.
			// Since we load the full interpolated value in P1, P2 is effectively a no-op / passthrough.
			res = b.EmitBitcast(typeFloat, ctx.GetGcnVgprId(b, details.Vdst))
		}
	} else {
		res = b.EmitConstantFloat(typeFloat, 0.0)
	}

	// Store result to destination VGPR.
	ctx.StoreRegisterPointer(b, gcnSpec.OpVgpr0+details.Vdst, b.EmitBitcast(typeUint, res))
}
