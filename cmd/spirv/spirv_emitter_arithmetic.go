package spirv

import "github.com/LamkasDev/sharkie/cmd/spirv/spec"

// EmitIAdd emits OpIAdd and returns the result ID.
func (b *SpvBuilder) EmitIAdd(resultType, op1, op2 SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpIAdd, uint32(resultType), uint32(id), uint32(op1), uint32(op2))
	return id
}

// EmitISub emits OpISub and returns the result ID.
func (b *SpvBuilder) EmitISub(resultType, op1, op2 SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpISub, uint32(resultType), uint32(id), uint32(op1), uint32(op2))
	return id
}

// EmitFAdd emits OpFAdd and returns the result ID.
func (b *SpvBuilder) EmitFAdd(resultType, op1, op2 SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpFAdd, uint32(resultType), uint32(id), uint32(op1), uint32(op2))
	return id
}

// EmitFSub emits OpFSub and returns the result ID.
func (b *SpvBuilder) EmitFSub(resultType, op1, op2 SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpFSub, uint32(resultType), uint32(id), uint32(op1), uint32(op2))
	return id
}

// EmitIMul emits OpIMul and returns the result ID.
func (b *SpvBuilder) EmitIMul(resultType, op1, op2 SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpIMul, uint32(resultType), uint32(id), uint32(op1), uint32(op2))
	return id
}

// EmitFMul emits OpFMul and returns the result ID.
func (b *SpvBuilder) EmitFMul(resultType, op1, op2 SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpFMul, uint32(resultType), uint32(id), uint32(op1), uint32(op2))
	return id
}

// EmitUDiv emits OpUDiv and returns the result ID.
func (b *SpvBuilder) EmitUDiv(resultType, op1, op2 SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpUDiv, uint32(resultType), uint32(id), uint32(op1), uint32(op2))
	return id
}

// EmitUMod emits OpUMod and returns the result ID.
func (b *SpvBuilder) EmitUMod(resultType, op1, op2 SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpUMod, uint32(resultType), uint32(id), uint32(op1), uint32(op2))
	return id
}

// EmitFDiv emits OpFDiv and returns the result ID.
func (b *SpvBuilder) EmitFDiv(resultType, op1, op2 SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpFDiv, uint32(resultType), uint32(id), uint32(op1), uint32(op2))
	return id
}

// EmitLogicalOr emits OpLogicalOr and returns the result ID.
func (b *SpvBuilder) EmitLogicalOr(resultType, op1, op2 SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpLogicalOr, uint32(resultType), uint32(id), uint32(op1), uint32(op2))
	return id
}

// EmitLogicalAnd emits OpLogicalAnd and returns the result ID.
func (b *SpvBuilder) EmitLogicalAnd(resultType, op1, op2 SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpLogicalAnd, uint32(resultType), uint32(id), uint32(op1), uint32(op2))
	return id
}

// EmitLogicalNot emits OpLogicalNot and returns the result ID.
func (b *SpvBuilder) EmitLogicalNot(resultType, operand SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpLogicalNot, uint32(resultType), uint32(id), uint32(operand))
	return id
}

// EmitIEqual emits OpIEqual and returns the result ID.
func (b *SpvBuilder) EmitIEqual(resultType, op1, op2 SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpIEqual, uint32(resultType), uint32(id), uint32(op1), uint32(op2))
	return id
}

// EmitBitwiseAnd emits OpBitwiseAnd and returns the result ID.
func (b *SpvBuilder) EmitBitwiseAnd(resultType, op1, op2 SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpBitwiseAnd, uint32(resultType), uint32(id), uint32(op1), uint32(op2))
	return id
}

// EmitNot emits OpNot and returns the result ID.
func (b *SpvBuilder) EmitNot(resultType, operand SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpNot, uint32(resultType), uint32(id), uint32(operand))
	return id
}

// EmitBitwiseOr emits OpBitwiseOr and returns the result ID.
func (b *SpvBuilder) EmitBitwiseOr(resultType, op1, op2 SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpBitwiseOr, uint32(resultType), uint32(id), uint32(op1), uint32(op2))
	return id
}

// EmitBitwiseXor emits OpBitwiseOr and returns the result ID.
func (b *SpvBuilder) EmitBitwiseXor(resultType, op1, op2 SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpBitwiseXor, uint32(resultType), uint32(id), uint32(op1), uint32(op2))
	return id
}

// EmitShiftLeftLogical emits OpShiftLeftLogical and returns the result ID.
func (b *SpvBuilder) EmitShiftLeftLogical(resultType, base, shift SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpShiftLeftLogical, uint32(resultType), uint32(id), uint32(base), uint32(shift))
	return id
}

// EmitShiftRightLogical emits OpShiftRightLogical and returns the result ID.
func (b *SpvBuilder) EmitShiftRightLogical(resultType, base, shift SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpShiftRightLogical, uint32(resultType), uint32(id), uint32(base), uint32(shift))
	return id
}

// EmitShiftRightArithmetic emits OpShiftRightArithmetic and returns the result ID.
func (b *SpvBuilder) EmitShiftRightArithmetic(resultType, base, shift SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpShiftRightArithmetic, uint32(resultType), uint32(id), uint32(base), uint32(shift))
	return id
}

// EmitINotEqual emits OpINotEqual and returns the result ID.
func (b *SpvBuilder) EmitINotEqual(resultType, op1, op2 SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpINotEqual, uint32(resultType), uint32(id), uint32(op1), uint32(op2))
	return id
}

// EmitUGreaterThan emits OpUGreaterThan and returns the result ID.
func (b *SpvBuilder) EmitUGreaterThan(resultType, op1, op2 SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpUGreaterThan, uint32(resultType), uint32(id), uint32(op1), uint32(op2))
	return id
}

// EmitUGreaterThanEqual emits OpUGreaterThanEqual and returns the result ID.
func (b *SpvBuilder) EmitUGreaterThanEqual(resultType, op1, op2 SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpUGreaterThanEqual, uint32(resultType), uint32(id), uint32(op1), uint32(op2))
	return id
}

// EmitULessThan emits OpULessThan and returns the result ID.
func (b *SpvBuilder) EmitULessThan(resultType, op1, op2 SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpULessThan, uint32(resultType), uint32(id), uint32(op1), uint32(op2))
	return id
}

// EmitBitFieldUExtract emits OpBitFieldUExtract and returns the result ID.
func (b *SpvBuilder) EmitBitFieldUExtract(resultType, base, offset, count SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpBitFieldUExtract, uint32(resultType), uint32(id), uint32(base), uint32(offset), uint32(count))
	return id
}

// EmitSampledImage emits OpSampledImage and returns the result ID.
func (b *SpvBuilder) EmitSampledImage(resultType, image, sampler SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpSampledImage, uint32(resultType), uint32(id), uint32(image), uint32(sampler))
	return id
}

// EmitImageSampleImplicitLod emits OpImageSampleImplicitLod and returns the result ID.
func (b *SpvBuilder) EmitImageSampleImplicitLod(resultType, sampledImage, coordinate SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpImageSampleImplicitLod, uint32(resultType), uint32(id), uint32(sampledImage), uint32(coordinate))
	return id
}

// EmitFOrdEqual emits OpFOrdEqual and returns the result ID.
func (b *SpvBuilder) EmitFOrdEqual(resultType, op1, op2 SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpFOrdEqual, uint32(resultType), uint32(id), uint32(op1), uint32(op2))
	return id
}

// EmitFUnordNotEqual emits OpFUnordNotEqual and returns the result ID.
func (b *SpvBuilder) EmitFUnordNotEqual(resultType, op1, op2 SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpFUnordNotEqual, uint32(resultType), uint32(id), uint32(op1), uint32(op2))
	return id
}

// EmitFOrdGreaterThan emits OpFOrdGreaterThan and returns the result ID.
func (b *SpvBuilder) EmitFOrdGreaterThan(resultType, op1, op2 SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.code, spec.SpvOpFOrdGreaterThan, uint32(resultType), uint32(id), uint32(op1), uint32(op2))
	return id
}
