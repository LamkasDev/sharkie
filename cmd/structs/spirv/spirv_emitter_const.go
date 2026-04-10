package spirv

import (
	"math"
)

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

// EmitConstantFloat emits OpConstant for a float.
func (b *SpvBuilder) EmitConstantFloat(floatType uint32, value float32) uint32 {
	id := b.AllocId()
	b.instr(&b.types, SpvOpConstant, floatType, id, math.Float32bits(value))
	return id
}

// EmitConstantComposite emits OpConstantComposite for a vector or composite constant.
func (b *SpvBuilder) EmitConstantComposite(resultType uint32, constituents ...uint32) uint32 {
	id := b.AllocId()
	operands := append([]uint32{resultType, id}, constituents...)
	b.instr(&b.types, SpvOpConstantComposite, operands...)
	return id
}
