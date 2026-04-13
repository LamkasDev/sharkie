package spirv

import (
	"go101.org/nstd"
)

// EmitTypeInt declares a void type.
func (b *SpvBuilder) EmitTypeVoid() uint32 {
	id := b.AllocId()
	b.instr(&b.types, SpvOpTypeVoid, id)
	return id
}

// EmitTypeInt declares a boolean type.
func (b *SpvBuilder) EmitTypeBool() uint32 {
	id := b.AllocId()
	b.instr(&b.types, SpvOpTypeBool, id)
	return id
}

// EmitTypeInt declares an integer type.
func (b *SpvBuilder) EmitTypeInt(width uint32, signed bool) uint32 {
	id := b.AllocId()
	b.instr(&b.types, SpvOpTypeInt, id, width, uint32(nstd.Btoi(signed)))
	return id
}

// EmitTypeFloat declares a float type.
func (b *SpvBuilder) EmitTypeFloat(width uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.types, SpvOpTypeFloat, id, width)
	return id
}

// EmitTypeVector declares a vector elementType[count].
func (b *SpvBuilder) EmitTypeVector(elementType, count uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.types, SpvOpTypeVector, id, elementType, count)
	return id
}

// EmitTypeArray declares an array elementType[length] (length is the ID of an integer constant).
func (b *SpvBuilder) EmitTypeArray(elementType, lengthID uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.types, SpvOpTypeArray, id, elementType, lengthID)
	return id
}

// EmitTypeRuntimeArray declares an unsized array elementType[].
func (b *SpvBuilder) EmitTypeRuntimeArray(elementType uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.types, SpvOpTypeRuntimeArray, id, elementType)
	return id
}

// EmitTypeImage declares an image type.
func (b *SpvBuilder) EmitTypeImage(sampledType, dim, depth, arrayed, ms, sampled, format uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.types, SpvOpTypeImage, id, sampledType, dim, depth, arrayed, ms, sampled, format)
	return id
}

// EmitTypeSampler declares a sampler type.
func (b *SpvBuilder) EmitTypeSampler() uint32 {
	id := b.AllocId()
	b.instr(&b.types, SpvOpTypeSampler, id)
	return id
}

// EmitTypeSampledImage declares a combined image/sampler type.
func (b *SpvBuilder) EmitTypeSampledImage(imageType uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.types, SpvOpTypeSampledImage, id, imageType)
	return id
}

// EmitTypeStruct declares a struct with given member types.
func (b *SpvBuilder) EmitTypeStruct(memberTypes ...uint32) uint32 {
	id := b.AllocId()
	operands := append([]uint32{id}, memberTypes...)
	b.instr(&b.types, SpvOpTypeStruct, operands...)
	return id
}

// EmitTypePointer declares a pointer to pointerType in the given storage class.
func (b *SpvBuilder) EmitTypePointer(storageClass, pointerType uint32) uint32 {
	id := b.AllocId()
	b.instr(&b.types, SpvOpTypePointer, id, storageClass, pointerType)
	return id
}

// EmitTypeFunction emits OpTypeFunction.
func (b *SpvBuilder) EmitTypeFunction(returnType uint32, paramTypes ...uint32) uint32 {
	id := b.AllocId()
	operands := append([]uint32{id, returnType}, paramTypes...)
	b.instr(&b.types, SpvOpTypeFunction, operands...)
	return id
}
