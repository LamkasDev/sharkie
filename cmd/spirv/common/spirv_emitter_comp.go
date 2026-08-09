package common

import "github.com/LamkasDev/sharkie/cmd/spirv/spec"

// EmitVectorShuffle emits OpVectorShuffle and returns the result id.
func (b *SpvBuilder) EmitVectorShuffle(resultType SpirvId, vector1 SpirvId, vector2 SpirvId, components ...uint32) SpirvId {
	id := b.AllocId()
	ops := append([]uint32{uint32(resultType), uint32(id), uint32(vector1), uint32(vector2)}, components...)
	b.instr(&b.code, spec.SpvOpVectorShuffle, ops...)
	return id
}

// EmitCompositeConstruct emits OpCompositeConstruct and returns the result id.
func (b *SpvBuilder) EmitCompositeConstruct(resultType SpirvId, constituents ...SpirvId) SpirvId {
	id := b.AllocId()
	ops := []uint32{uint32(resultType), uint32(id)}
	for _, c := range constituents {
		ops = append(ops, uint32(c))
	}
	b.instr(&b.code, spec.SpvOpCompositeConstruct, ops...)
	return id
}

// EmitCompositeExtract emits OpCompositeExtract and returns the result id.
func (b *SpvBuilder) EmitCompositeExtract(resultType, composite SpirvId, indices ...uint32) SpirvId {
	id := b.AllocId()
	ops := append([]uint32{uint32(resultType), uint32(id), uint32(composite)}, indices...)
	b.instr(&b.code, spec.SpvOpCompositeExtract, ops...)
	return id
}
