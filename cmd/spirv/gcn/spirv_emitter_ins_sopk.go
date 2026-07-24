package gcn

import (
	"fmt"

	gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"
	. "github.com/LamkasDev/sharkie/cmd/spirv/common"
)

func EmitSOPK(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*gcnSpec.ScalarDetails)
	switch details.Op {
	case gcnSpec.SopkOpMovkI32:
		// D.i = signext(SIMM16)
		typeUint := ctx.GetId(BlockContextIdTypeUint)

		simm16 := uint32(int32(int16(instr.Literal)))
		val := b.EmitConstantUint(typeUint, simm16)
		ctx.StoreRegisterPointer(b, details.Dst, val)
	case gcnSpec.SopkOpAddkI32:
		// D.i = D.i + signext(SIMM16). SCC = signed overflow.
		typeUint := ctx.GetId(BlockContextIdTypeUint)
		idC31 := ctx.GetConstId(ConstIdUint31)

		simm16 := uint32(int32(int16(instr.Literal)))
		val := b.EmitConstantUint(typeUint, simm16)
		dst := ctx.GetOperandUintValue(b, details.Dst, 0)
		res := b.EmitIAdd(typeUint, dst, val)
		ctx.StoreRegisterPointer(b, details.Dst, res)

		// Calculate signed overflow - MSB of ((A ^ Result) & (B ^ Result)).
		dstXorRes := b.EmitBitwiseXor(typeUint, dst, res)
		valXorRes := b.EmitBitwiseXor(typeUint, val, res)
		overflowMask := b.EmitBitwiseAnd(typeUint, dstXorRes, valXorRes)
		overflowBit := b.EmitShiftRightLogical(typeUint, overflowMask, idC31)

		emitSccUpdateNonZero(b, ctx, overflowBit)
	default:
		panic(fmt.Sprintf("unknown sopk op %s", gcnSpec.Mnemotics[gcnSpec.EncSOPK][details.Op]))
	}
}
