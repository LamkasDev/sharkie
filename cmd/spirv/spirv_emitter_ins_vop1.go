package spirv

import (
	"fmt"

	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
	gcnSpec "github.com/LamkasDev/sharkie/cmd/structs/gcn/spec"
)

func emitVOP1(b *SpvBuilder, instr *gcnSpec.Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*gcnSpec.VectorDetails)
	switch details.Op {
	case gcnSpec.Vop1OpMovB32:
		val := ctx.GetOperandValue(b, details.Src0, instr.Literal)
		ctx.StoreRegisterPointerMasked(b, details.Dst+gcnSpec.OpVgpr0, val)
	case gcnSpec.Vop1OpSqrtF32:
		val := ctx.GetOperandFloatValue(b, details.Src0, instr.Literal)
		resF := b.EmitExtInst(ctx.GetId(BlockContextIdTypeFloat), ctx.GetId(BlockContextIdGlsl), spec.SpvGlslOpSqrt, val)
		ctx.StoreRegisterPointerMasked(b, details.Dst+gcnSpec.OpVgpr0, b.EmitBitcast(ctx.GetId(BlockContextIdTypeUint), resF))
	case gcnSpec.Vop1OpRcpF32:
		typeFloat := ctx.GetId(BlockContextIdTypeFloat)

		val := ctx.GetOperandFloatValue(b, details.Src0, instr.Literal)
		resF := b.EmitFDiv(typeFloat, ctx.GetConstId(ConstIdFloat1), val)
		ctx.StoreRegisterPointerMasked(b, details.Dst+gcnSpec.OpVgpr0, b.EmitBitcast(ctx.GetId(BlockContextIdTypeUint), resF))
	default:
		panic(fmt.Sprintf("unknown vop1 op %s", gcnSpec.Mnemotics[gcnSpec.EncVOP1][details.Op]))
	}
}
