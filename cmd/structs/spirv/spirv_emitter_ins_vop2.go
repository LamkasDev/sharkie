package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitVOP2(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	switch instr.Details.(*VectorDetails).Op {
	case Vop2OpCvtPkrtzF16F32:
		val0 := ctx.GetOperandValue(b, instr.Details.(*VectorDetails).Src0, instr.Literal)
		val1 := ctx.GetOperandValue(b, instr.Details.(*VectorDetails).Src1+OpVgpr0, 0)
		fval0 := b.EmitBitcast(ctx.GetId(SpirvBlockContextIdFloat), val0)
		fval1 := b.EmitBitcast(ctx.GetId(SpirvBlockContextIdFloat), val1)
		vec := b.EmitCompositeConstruct(ctx.GetId(SpirvBlockContextIdV2Float), fval0, fval1)
		resU := b.EmitExtInst(ctx.GetId(SpirvBlockContextIdUint), ctx.GetId(SpirvBlockContextIdGlsl), 58, vec) // PackHalf2x16
		ptr := ctx.GetRegisterPointer(instr.Details.(*VectorDetails).Dst + OpVgpr0)
		b.EmitStore(ptr, resU)
	default:
		panic(fmt.Sprintf("unknown vop2 op %d", instr.Details.(*VectorDetails).Op))
	}
}
