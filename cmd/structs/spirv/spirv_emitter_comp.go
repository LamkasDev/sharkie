package spirv

// EmitCompositeConstruct emits OpCompositeConstruct and returns the result id.
func (b *SpvBuilder) EmitCompositeConstruct(resultType uint32, constituents ...uint32) uint32 {
	id := b.AllocId()
	ops := append([]uint32{resultType, id}, constituents...)
	b.instr(&b.code, SpvOpCompositeConstruct, ops...)
	return id
}

// EmitCompositeExtract emits OpCompositeExtract and returns the result id.
func (b *SpvBuilder) EmitCompositeExtract(resultType, composite uint32, indices ...uint32) uint32 {
	id := b.AllocId()
	ops := append([]uint32{resultType, id, composite}, indices...)
	b.instr(&b.code, SpvOpCompositeExtract, ops...)
	return id
}
