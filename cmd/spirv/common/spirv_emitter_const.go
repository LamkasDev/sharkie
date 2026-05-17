package common

import (
	"math"

	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
)

// EmitConstantTrue emits OpConstantTrue.
func (b *SpvBuilder) EmitConstantTrue(boolType SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.types, spec.SpvOpConstantTrue, uint32(boolType), uint32(id))
	return id
}

// EmitConstantFalse emits OpConstantFalse.
func (b *SpvBuilder) EmitConstantFalse(boolType SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.types, spec.SpvOpConstantFalse, uint32(boolType), uint32(id))
	return id
}

// EmitConstantUint emits OpConstant for a 32-bit unsigned integer.
func (b *SpvBuilder) EmitConstantUint(uintType SpirvId, value uint32) SpirvId {
	id := b.AllocId()
	b.instr(&b.types, spec.SpvOpConstant, uint32(uintType), uint32(id), value)
	return id
}

// EmitConstantUint64 emits OpConstant for a uint64.
func (b *SpvBuilder) EmitConstantUint64(resultType SpirvId, value uint64) SpirvId {
	id := b.AllocId()
	b.instr(&b.types, spec.SpvOpConstant, uint32(resultType), uint32(id), uint32(value), uint32(value>>32))
	return id
}

// EmitConstantFloat emits OpConstant for a float32.
func (b *SpvBuilder) EmitConstantFloat(resultType SpirvId, value float32) SpirvId {
	id := b.AllocId()
	b.instr(&b.types, spec.SpvOpConstant, uint32(resultType), uint32(id), math.Float32bits(value))
	return id
}

// EmitDeferredConstantUint emits OpConstant for a 32-bit unsigned integer into the deferredConstants section.
func (b *SpvBuilder) EmitDeferredConstantUint(uintType, id SpirvId, value uint32) {
	b.instr(&b.deferredConstants, spec.SpvOpConstant, uint32(uintType), uint32(id), value)
}

// EmitDeferredConstantUint64 emits OpConstant for a 64-bit unsigned integer into the deferredConstants section.
func (b *SpvBuilder) EmitDeferredConstantUint64(uint64Type, id SpirvId, value uint64) {
	b.instr(&b.deferredConstants, spec.SpvOpConstant, uint32(uint64Type), uint32(id), uint32(value), uint32(value>>32))
}

// EmitDeferredConstantFloat emits OpConstant for a float32 into the deferredConstants section.
func (b *SpvBuilder) EmitDeferredConstantFloat(resultType, id SpirvId, value float32) {
	b.instr(&b.deferredConstants, spec.SpvOpConstant, uint32(resultType), uint32(id), math.Float32bits(value))
}

// EmitConstantComposite emits OpConstantComposite for a vector or composite constant.
func (b *SpvBuilder) EmitConstantComposite(resultType SpirvId, constituents ...SpirvId) SpirvId {
	id := b.AllocId()
	operands := []uint32{uint32(resultType), uint32(id)}
	for _, c := range constituents {
		operands = append(operands, uint32(c))
	}
	b.instr(&b.types, spec.SpvOpConstantComposite, operands...)
	return id
}
