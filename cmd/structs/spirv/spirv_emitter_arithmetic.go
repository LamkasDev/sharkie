package spirv

// EmitIAdd emits OpIAdd and returns the result ID.
func (b *SpvBuilder) EmitIAdd(resultType, op1, op2 uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpIAdd, resultType, id, op1, op2)
	return id
}

// EmitISub emits OpISub and returns the result ID.
func (b *SpvBuilder) EmitISub(resultType, op1, op2 uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpISub, resultType, id, op1, op2)
	return id
}

// EmitFAdd emits OpFAdd and returns the result ID.
func (b *SpvBuilder) EmitFAdd(resultType, op1, op2 uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpFAdd, resultType, id, op1, op2)
	return id
}

// EmitFSub emits OpFSub and returns the result ID.
func (b *SpvBuilder) EmitFSub(resultType, op1, op2 uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpFSub, resultType, id, op1, op2)
	return id
}

// EmitIMul emits OpIMul and returns the result ID.
func (b *SpvBuilder) EmitIMul(resultType, op1, op2 uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpIMul, resultType, id, op1, op2)
	return id
}

// EmitFMul emits OpFMul and returns the result ID.
func (b *SpvBuilder) EmitFMul(resultType, op1, op2 uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpFMul, resultType, id, op1, op2)
	return id
}

// EmitUDiv emits OpUDiv and returns the result ID.
func (b *SpvBuilder) EmitUDiv(resultType, op1, op2 uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpUDiv, resultType, id, op1, op2)
	return id
}

// EmitUMod emits OpUMod and returns the result ID.
func (b *SpvBuilder) EmitUMod(resultType, op1, op2 uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpUMod, resultType, id, op1, op2)
	return id
}

// EmitFDiv emits OpFDiv and returns the result ID.
func (b *SpvBuilder) EmitFDiv(resultType, op1, op2 uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpFDiv, resultType, id, op1, op2)
	return id
}

// EmitLogicalOr emits OpLogicalOr and returns the result ID.
func (b *SpvBuilder) EmitLogicalOr(resultType, op1, op2 uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpLogicalOr, resultType, id, op1, op2)
	return id
}

// EmitLogicalAnd emits OpLogicalAnd and returns the result ID.
func (b *SpvBuilder) EmitLogicalAnd(resultType, op1, op2 uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpLogicalAnd, resultType, id, op1, op2)
	return id
}

// EmitLogicalNot emits OpLogicalNot and returns the result ID.
func (b *SpvBuilder) EmitLogicalNot(resultType, operand uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpLogicalNot, resultType, id, operand)
	return id
}

// EmitIEqual emits OpIEqual and returns the result ID.
func (b *SpvBuilder) EmitIEqual(resultType, op1, op2 uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpIEqual, resultType, id, op1, op2)
	return id
}

// EmitBitwiseAnd emits OpBitwiseAnd and returns the result ID.
func (b *SpvBuilder) EmitBitwiseAnd(resultType, op1, op2 uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpBitwiseAnd, resultType, id, op1, op2)
	return id
}

// EmitNot emits OpNot and returns the result ID.
func (b *SpvBuilder) EmitNot(resultType, operand uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpNot, resultType, id, operand)
	return id
}

// EmitBitwiseOr emits OpBitwiseOr and returns the result ID.
func (b *SpvBuilder) EmitBitwiseOr(resultType, op1, op2 uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpBitwiseOr, resultType, id, op1, op2)
	return id
}

// EmitBitwiseXor emits OpBitwiseOr and returns the result ID.
func (b *SpvBuilder) EmitBitwiseXor(resultType, op1, op2 uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpBitwiseXor, resultType, id, op1, op2)
	return id
}

// EmitShiftLeftLogical emits OpShiftLeftLogical and returns the result ID.
func (b *SpvBuilder) EmitShiftLeftLogical(resultType, base, shift uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpShiftLeftLogical, resultType, id, base, shift)
	return id
}

// EmitShiftRightLogical emits OpShiftRightLogical and returns the result ID.
func (b *SpvBuilder) EmitShiftRightLogical(resultType, base, shift uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpShiftRightLogical, resultType, id, base, shift)
	return id
}

// EmitShiftRightArithmetic emits OpShiftRightArithmetic and returns the result ID.
func (b *SpvBuilder) EmitShiftRightArithmetic(resultType, base, shift uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpShiftRightArithmetic, resultType, id, base, shift)
	return id
}

// EmitINotEqual emits OpINotEqual and returns the result ID.
func (b *SpvBuilder) EmitINotEqual(resultType, op1, op2 uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpINotEqual, resultType, id, op1, op2)
	return id
}

// EmitUGreaterThan emits OpUGreaterThan and returns the result ID.
func (b *SpvBuilder) EmitUGreaterThan(resultType, op1, op2 uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpUGreaterThan, resultType, id, op1, op2)
	return id
}

// EmitULessThan emits OpULessThan and returns the result ID.
func (b *SpvBuilder) EmitULessThan(resultType, op1, op2 uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpULessThan, resultType, id, op1, op2)
	return id
}

// EmitBitFieldUExtract emits OpBitFieldUExtract and returns the result ID.
func (b *SpvBuilder) EmitBitFieldUExtract(resultType, base, offset, count uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpBitFieldUExtract, resultType, id, base, offset, count)
	return id
}

// EmitSampledImage emits OpSampledImage and returns the result ID.
func (b *SpvBuilder) EmitSampledImage(resultType, image, sampler uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpSampledImage, resultType, id, image, sampler)
	return id
}

// EmitImageSampleImplicitLod emits OpImageSampleImplicitLod and returns the result ID.
func (b *SpvBuilder) EmitImageSampleImplicitLod(resultType, sampledImage, coordinate uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpImageSampleImplicitLod, resultType, id, sampledImage, coordinate)
	return id
}

// EmitFOrdEqual emits OpFOrdEqual and returns the result ID.
func (b *SpvBuilder) EmitFOrdEqual(resultType, op1, op2 uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpFOrdEqual, resultType, id, op1, op2)
	return id
}

// EmitFUnordNotEqual emits OpFUnordNotEqual and returns the result ID.
func (b *SpvBuilder) EmitFUnordNotEqual(resultType, op1, op2 uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpFUnordNotEqual, resultType, id, op1, op2)
	return id
}

// EmitFOrdGreaterThan emits OpFOrdGreaterThan and returns the result ID.
func (b *SpvBuilder) EmitFOrdGreaterThan(resultType, op1, op2 uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.code, SpvOpFOrdGreaterThan, resultType, id, op1, op2)
	return id
}
