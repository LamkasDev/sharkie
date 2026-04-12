package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

func emitVOP2(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	details := instr.Details.(*VectorDetails)
	switch details.Op {
	case Vop2OpAddF32:
		val0 := ctx.GetOperandFloatValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandFloatValue(b, details.Src1+OpVgpr0, 0)
		resF := b.EmitFAdd(ctx.GetId(SpirvBlockContextIdTypeFloat), val0, val1)
		ctx.StoreRegisterPointer(b, details.Dst+OpVgpr0, b.EmitBitcast(ctx.GetId(SpirvBlockContextIdTypeUint), resF))
	case Vop2OpSubF32:
		val0 := ctx.GetOperandFloatValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandFloatValue(b, details.Src1+OpVgpr0, 0)
		resF := b.EmitFSub(ctx.GetId(SpirvBlockContextIdTypeFloat), val0, val1)
		ctx.StoreRegisterPointer(b, details.Dst+OpVgpr0, b.EmitBitcast(ctx.GetId(SpirvBlockContextIdTypeUint), resF))
	case Vop2OpSubrevF32:
		val0 := ctx.GetOperandFloatValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandFloatValue(b, details.Src1+OpVgpr0, 0)
		resF := b.EmitFSub(ctx.GetId(SpirvBlockContextIdTypeFloat), val1, val0)
		ctx.StoreRegisterPointer(b, details.Dst+OpVgpr0, b.EmitBitcast(ctx.GetId(SpirvBlockContextIdTypeUint), resF))
	case Vop2OpMulF32:
		val0 := ctx.GetOperandFloatValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandFloatValue(b, details.Src1+OpVgpr0, 0)
		resF := b.EmitFMul(ctx.GetId(SpirvBlockContextIdTypeFloat), val0, val1)
		ctx.StoreRegisterPointer(b, details.Dst+OpVgpr0, b.EmitBitcast(ctx.GetId(SpirvBlockContextIdTypeUint), resF))
	case Vop2OpMinF32:
		val0 := ctx.GetOperandFloatValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandFloatValue(b, details.Src1+OpVgpr0, 0)
		resF := b.EmitExtInst(ctx.GetId(SpirvBlockContextIdTypeFloat), ctx.GetId(SpirvBlockContextIdGlsl), SpvGlslOpFMin, val0, val1)
		ctx.StoreRegisterPointer(b, details.Dst+OpVgpr0, b.EmitBitcast(ctx.GetId(SpirvBlockContextIdTypeUint), resF))
	case Vop2OpMaxF32:
		val0 := ctx.GetOperandFloatValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandFloatValue(b, details.Src1+OpVgpr0, 0)
		resF := b.EmitExtInst(ctx.GetId(SpirvBlockContextIdTypeFloat), ctx.GetId(SpirvBlockContextIdGlsl), SpvGlslOpFMax, val0, val1)
		ctx.StoreRegisterPointer(b, details.Dst+OpVgpr0, b.EmitBitcast(ctx.GetId(SpirvBlockContextIdTypeUint), resF))
	case Vop2OpMacF32:
		val0 := ctx.GetOperandFloatValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandFloatValue(b, details.Src1+OpVgpr0, 0)
		valD := b.EmitBitcast(ctx.GetId(SpirvBlockContextIdTypeFloat), ctx.LoadRegisterPointer(b, details.Dst+OpVgpr0))
		resF := b.EmitExtInst(ctx.GetId(SpirvBlockContextIdTypeFloat), ctx.GetId(SpirvBlockContextIdGlsl), SpvGlslOpFma, val0, val1, valD)
		ctx.StoreRegisterPointer(b, details.Dst+OpVgpr0, b.EmitBitcast(ctx.GetId(SpirvBlockContextIdTypeUint), resF))
	case Vop2OpCvtPkrtzF16F32:
		val0 := ctx.GetOperandFloatValue(b, details.Src0, instr.Literal)
		val1 := ctx.GetOperandFloatValue(b, details.Src1+OpVgpr0, 0)
		vec := b.EmitCompositeConstruct(ctx.GetId(SpirvBlockContextIdTypeV2Float), val0, val1)
		resU := b.EmitExtInst(ctx.GetId(SpirvBlockContextIdTypeUint), ctx.GetId(SpirvBlockContextIdGlsl), SpvGlslOpPackHalf2x16, vec)
		ctx.StoreRegisterPointer(b, details.Dst+OpVgpr0, resU)
	default:
		panic(fmt.Sprintf("unknown vop2 op %s", Mnemotics[EncVOP2][details.Op]))
	}
}
