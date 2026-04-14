package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitVOP1(b *SpvBuilder, instr *Instruction, ctx *SpirvBlockContext) {
	details := instr.Details.(*VectorDetails)
	switch details.Op {
	case Vop1OpMovB32:
		val := ctx.GetOperandValue(b, details.Src0, instr.Literal)
		ctx.StoreRegisterPointer(b, details.Dst+OpVgpr0, val)
	case Vop1OpSqrtF32:
		val := ctx.GetOperandFloatValue(b, details.Src0, instr.Literal)
		resF := b.EmitExtInst(ctx.GetId(BlockContextIdTypeFloat), ctx.GetId(BlockContextIdGlsl), SpvGlslOpSqrt, val)
		ctx.StoreRegisterPointer(b, details.Dst+OpVgpr0, b.EmitBitcast(ctx.GetId(BlockContextIdTypeUint), resF))
	case Vop1OpRcpF32:
		typeFloat := ctx.GetId(BlockContextIdTypeFloat)

		val := ctx.GetOperandFloatValue(b, details.Src0, instr.Literal)
		resF := b.EmitFDiv(typeFloat, ctx.GetConstId(ConstIdxFloat1), val)
		ctx.StoreRegisterPointer(b, details.Dst+OpVgpr0, b.EmitBitcast(ctx.GetId(BlockContextIdTypeUint), resF))
	default:
		panic(fmt.Sprintf("unknown vop1 op %s", Mnemotics[EncVOP1][details.Op]))
	}
}
