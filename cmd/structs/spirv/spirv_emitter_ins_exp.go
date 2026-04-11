package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitEXP(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	details := instr.Details.(*ExpDetails)
	switch details.Target {
	case 0: // target=0 (color)
		val0 := ctx.GetOperandValue(b, details.VSrcs[0]+OpVgpr0, 0)
		val1 := ctx.GetOperandValue(b, details.VSrcs[1]+OpVgpr0, 0)
		v01 := b.EmitExtInst(ctx.GetId(SpirvBlockContextIdTypeV2Float), ctx.GetId(SpirvBlockContextIdGlsl), 62, val0) // UnpackHalf2x16
		v23 := b.EmitExtInst(ctx.GetId(SpirvBlockContextIdTypeV2Float), ctx.GetId(SpirvBlockContextIdGlsl), 62, val1) // UnpackHalf2x16
		f0 := b.EmitCompositeExtract(ctx.GetId(SpirvBlockContextIdTypeFloat), v01, 0)
		f1 := b.EmitCompositeExtract(ctx.GetId(SpirvBlockContextIdTypeFloat), v01, 1)
		f2 := b.EmitCompositeExtract(ctx.GetId(SpirvBlockContextIdTypeFloat), v23, 0)
		f3 := b.EmitCompositeExtract(ctx.GetId(SpirvBlockContextIdTypeFloat), v23, 1)
		vec := b.EmitCompositeConstruct(ctx.GetId(SpirvBlockContextIdTypeV4Float), f0, f1, f2, f3)
		b.EmitStore(ctx.GetId(SpirvBlockContextIdColorOut), vec)
	default:
		panic(fmt.Sprintf("unknown export target %d", details.Target))
	}
}
