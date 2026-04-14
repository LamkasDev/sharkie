package spirv

import (
	"math"
)

// EmitConstantFalse emits OpConstantTrue.
func (b *SpvBuilder) EmitConstantTrue(boolType uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.types, SpvOpConstantTrue, boolType, id)
	return id
}

// EmitConstantFalse emits OpConstantFalse.
func (b *SpvBuilder) EmitConstantFalse(boolType uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.types, SpvOpConstantFalse, boolType, id)
	return id
}

// EmitConstantUint emits OpConstant for a 32-bit unsigned integer.
func (b *SpvBuilder) EmitConstantUint(uintType, value uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.types, SpvOpConstant, uintType, id, value)
	return id
}

// EmitConstantUint64 emits OpConstant for a uint64.
func (b *SpvBuilder) EmitConstantUint64(resultType uint32, value uint64) uint32 {
	id := b.AllocId()
	b.instr(&b.types, SpvOpConstant, resultType, id, uint32(value), uint32(value>>32))
	return id
}

// EmitConstantFloat emits OpConstant for a float32.
func (b *SpvBuilder) EmitConstantFloat(resultType uint32, value float32) uint32 {
	id := b.AllocId()
	b.instr(&b.types, SpvOpConstant, resultType, id, math.Float32bits(value))
	return id
}

// EmitDeferredConstantUint emits OpConstant for a 32-bit unsigned integer into the deferredConstants section.
func (b *SpvBuilder) EmitDeferredConstantUint(uintType, id, value uint32) {
	b.instr(&b.deferredConstants, SpvOpConstant, uintType, id, value)
}

// EmitDeferredConstantUint emits OpConstant for a 64-bit unsigned integer into the deferredConstants section.
func (b *SpvBuilder) EmitDeferredConstantUint64(uint64Type, id uint32, value uint64) {
	b.instr(&b.deferredConstants, SpvOpConstant, uint64Type, id, uint32(value), uint32(value>>32))
}

// EmitDeferredConstantFloat emits OpConstant for a float32 into the deferredConstants section.
func (b *SpvBuilder) EmitDeferredConstantFloat(resultType, id uint32, value float32) {
	b.instr(&b.deferredConstants, SpvOpConstant, resultType, id, math.Float32bits(value))
}

// EmitConstantComposite emits OpConstantComposite for a vector or composite constant.
func (b *SpvBuilder) EmitConstantComposite(resultType uint32, constituents ...uint32) uint32 {
	id := b.AllocId()
	operands := append([]uint32{resultType, id}, constituents...)
	b.instr(&b.types, SpvOpConstantComposite, operands...)
	return id
}
