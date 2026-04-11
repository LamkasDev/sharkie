package spirv

import (
	"fmt"

	. "github.com/LamkasDev/sharkie/cmd/structs/gcn"
)

type InstructionEmitFunc func(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext)

var InstructionEmitMap = map[Encoding]InstructionEmitFunc{
	EncSOP2:  func(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {},
	EncSOP1:  emitSOP1,
	EncSOPC:  func(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {},
	EncSOPP:  emitSOPP,
	EncVOP2:  emitVOP2,
	EncVOP1:  emitVOP1,
	EncVOPC:  func(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {},
	EncVOP3:  func(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {},
	EncSMRD:  emitSMRD,
	EncMUBUF: func(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {},
	EncMIMG:  func(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {},
	EncEXP:   emitEXP,
}

// GetRegisterPointer returns the result ID of the pointer to the given register.
func (ctx *SpirvBlockContext) GetRegisterPointer(op uint32) uint32 {
	switch {
	case op <= 103:
		return ctx.GetSgprId(op)
	case op >= 106 && op <= 127:
		return ctx.GetSpecialId(op - 106)
	case op >= 251 && op <= 253:
		return ctx.GetSpecialId((op - 251) + 22)
	case op >= 256 && op <= 511:
		return ctx.GetVgprId(op - 256)
	}

	panic(fmt.Sprintf("unknown op %d", op))
}

// GetOperandValue returns the result ID of the value of the given operand.
func (ctx *SpirvBlockContext) GetOperandValue(b *SpvBuilder, op uint32, literal uint32) uint32 {
	switch {
	case op <= 103:
		return b.EmitLoad(ctx.GetId(SpirvBlockContextIdUint), ctx.GetSgprId(op))
	case op >= 106 && op <= 127:
		return b.EmitLoad(ctx.GetId(SpirvBlockContextIdUint), ctx.GetSpecialId(op-106))
	case op >= 128 && op <= 247:
		return ctx.GetConstId(op - 128)
	case op >= 251 && op <= 253:
		return b.EmitLoad(ctx.GetId(SpirvBlockContextIdUint), ctx.GetSpecialId((op-251)+22))
	case op == 255:
		return b.EmitConstantUint(ctx.GetId(SpirvBlockContextIdUint), literal)
	case op >= 256 && op <= 511:
		return b.EmitLoad(ctx.GetId(SpirvBlockContextIdUint), ctx.GetVgprId(op-256))
	}

	panic(fmt.Sprintf("unknown op %d", op))
}

// emitInstruction emits the SPIR-V for a single instruction.
func emitInstruction(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	emitFunc, ok := InstructionEmitMap[instr.Encoding]
	if !ok {
		panic(fmt.Errorf("unknown encoding %s", instr.Encoding))
	}
	emitFunc(b, instr, ctx)
}

func emitSOP1(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	switch instr.SOp {
	case 0x03: // s_mov_b32
		val := ctx.GetOperandValue(b, instr.SSrc0, instr.Literal)
		ptr := ctx.GetRegisterPointer(instr.SDst)
		b.EmitStore(ptr, val)
	}
}

func emitSOPP(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	switch instr.SOp {
	case SoppOpWaitCnt:
		// No-op in SPIR-V for now.
	}
}

func emitVOP2(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	switch instr.VOp {
	case 0x2F: // v_cvt_pkrtz_f16_f32
		val0 := ctx.GetOperandValue(b, instr.VSrc0, instr.Literal)
		val1 := ctx.GetOperandValue(b, instr.VSrc1+256, 0)
		fval0 := b.EmitBitcast(ctx.GetId(SpirvBlockContextIdFloat), val0)
		fval1 := b.EmitBitcast(ctx.GetId(SpirvBlockContextIdFloat), val1)
		vec := b.EmitCompositeConstruct(ctx.GetId(SpirvBlockContextIdV2Float), fval0, fval1)
		resU := b.EmitExtInst(ctx.GetId(SpirvBlockContextIdUint), ctx.GetId(SpirvBlockContextIdGlsl), 58, vec) // PackHalf2x16
		ptr := ctx.GetRegisterPointer(instr.VDst + 256)
		b.EmitStore(ptr, resU)
	}
}

func emitVOP1(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	switch instr.VOp {
	case 0x01: // v_mov_b32
		val := ctx.GetOperandValue(b, instr.VSrc0, instr.Literal)
		ptr := ctx.GetRegisterPointer(instr.VDst + 256)
		b.EmitStore(ptr, val)
	}
}

func emitSMRD(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	switch instr.SmOp {
	case 0x0A: // s_buffer_load_dwordx4
		// Load constant RAM base address (pointer) from push constant.
		idPtrPsbUint := ctx.GetId(SpirvBlockContextIdPtrPsbUint)
		ptrPcPsbUint := b.EmitAccessChain(ctx.GetId(SpirvBlockContextIdPtrPcPsbUint), ctx.GetId(SpirvBlockContextIdPcVar), b.EmitConstantUint(ctx.GetId(SpirvBlockContextIdUint), 2))
		ptrBase := b.EmitLoad(idPtrPsbUint, ptrPcPsbUint)

		// Calculate offset in dwords.
		var offset uint32
		if !instr.SmImmOff {
			panic("s_buffer_load_dwordx4 with non-immediate offset not implemented")
		}
		if instr.HasLiteral {
			// 64-bit SMRD: offset is a 32-bit byte offset.
			offset = b.EmitConstantUint(ctx.GetId(SpirvBlockContextIdUint), instr.Literal/4)
		} else {
			// 32-bit SMRD: offset is an 8-bit dword offset.
			offset = b.EmitConstantUint(ctx.GetId(SpirvBlockContextIdUint), instr.SmOffset)
		}

		for i := uint32(0); i < 4; i++ {
			var idx uint32
			if i == 0 {
				idx = offset
			} else {
				idx = b.EmitIAdd(ctx.GetId(SpirvBlockContextIdUint), offset, ctx.GetConstId(i))
			}
			ptr := b.EmitPtrAccessChain(idPtrPsbUint, ptrBase, idx)
			val := b.EmitLoad(ctx.GetId(SpirvBlockContextIdUint), ptr, SpvMemoryAccessAligned, 4)
			ptrReg := ctx.GetRegisterPointer(instr.SmDst + i)
			b.EmitStore(ptrReg, val)
		}
	}
}

func emitEXP(b *SpvBuilder, instr *Instruction, ctx SpirvBlockContext) {
	if instr.ExpTarget == 0 { // target=0 (color)
		val0 := ctx.GetOperandValue(b, instr.ExpVSrcs[0]+256, 0)
		val1 := ctx.GetOperandValue(b, instr.ExpVSrcs[1]+256, 0)
		v01 := b.EmitExtInst(ctx.GetId(SpirvBlockContextIdV2Float), ctx.GetId(SpirvBlockContextIdGlsl), 62, val0) // UnpackHalf2x16
		v23 := b.EmitExtInst(ctx.GetId(SpirvBlockContextIdV2Float), ctx.GetId(SpirvBlockContextIdGlsl), 62, val1) // UnpackHalf2x16
		f0 := b.EmitCompositeExtract(ctx.GetId(SpirvBlockContextIdFloat), v01, 0)
		f1 := b.EmitCompositeExtract(ctx.GetId(SpirvBlockContextIdFloat), v01, 1)
		f2 := b.EmitCompositeExtract(ctx.GetId(SpirvBlockContextIdFloat), v23, 0)
		f3 := b.EmitCompositeExtract(ctx.GetId(SpirvBlockContextIdFloat), v23, 1)
		vec := b.EmitCompositeConstruct(ctx.GetId(SpirvBlockContextIdV4Float), f0, f1, f2, f3)
		if idColorOut, ok := ctx.TryGetId(SpirvBlockContextIdColorOut); ok {
			b.EmitStore(idColorOut, vec)
		}
	}
}
