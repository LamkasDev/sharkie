package common

import (
	"github.com/LamkasDev/sharkie/cmd/spirv/spec"
	"go101.org/nstd"
)

// EmitTypeInt declares a void type.
func (b *SpvBuilder) EmitTypeVoid() SpirvId {
	id := b.AllocId()
	b.instr(&b.types, spec.SpvOpTypeVoid, uint32(id))
	return id
}

// EmitTypeInt declares a boolean type.
func (b *SpvBuilder) EmitTypeBool() SpirvId {
	id := b.AllocId()
	b.instr(&b.types, spec.SpvOpTypeBool, uint32(id))
	return id
}

// EmitTypeInt declares an integer type.
func (b *SpvBuilder) EmitTypeInt(width uint32, signed bool) SpirvId {
	id := b.AllocId()
	b.instr(&b.types, spec.SpvOpTypeInt, uint32(id), width, uint32(nstd.Btoi(signed)))
	return id
}

// EmitTypeFloat declares a float type.
func (b *SpvBuilder) EmitTypeFloat(width uint32) SpirvId {
	id := b.AllocId()
	b.instr(&b.types, spec.SpvOpTypeFloat, uint32(id), width)
	return id
}

// EmitTypeVector declares a vector elementType[count].
func (b *SpvBuilder) EmitTypeVector(elementType SpirvId, count uint32) SpirvId {
	id := b.AllocId()
	b.instr(&b.types, spec.SpvOpTypeVector, uint32(id), uint32(elementType), count)
	return id
}

// EmitTypeArray declares an array elementType[length] (length is the ID of an integer constant).
func (b *SpvBuilder) EmitTypeArray(elementType, lengthID SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.types, spec.SpvOpTypeArray, uint32(id), uint32(elementType), uint32(lengthID))
	return id
}

// EmitTypeRuntimeArray declares an unsized array elementType[].
func (b *SpvBuilder) EmitTypeRuntimeArray(elementType SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.types, spec.SpvOpTypeRuntimeArray, uint32(id), uint32(elementType))
	return id
}

// EmitTypeImage declares an image type.
func (b *SpvBuilder) EmitTypeImage(sampledType SpirvId, dim, depth, arrayed, ms, sampled, format uint32) SpirvId {
	id := b.AllocId()
	b.instr(&b.types, spec.SpvOpTypeImage, uint32(id), uint32(sampledType), dim, depth, arrayed, ms, sampled, format)
	return id
}

// EmitTypeSampler declares a sampler type.
func (b *SpvBuilder) EmitTypeSampler() SpirvId {
	id := b.AllocId()
	b.instr(&b.types, spec.SpvOpTypeSampler, uint32(id))
	return id
}

// EmitTypeSampledImage declares a combined image/sampler type.
func (b *SpvBuilder) EmitTypeSampledImage(imageType SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.types, spec.SpvOpTypeSampledImage, uint32(id), uint32(imageType))
	return id
}

// EmitTypeStruct declares a struct with given member types.
func (b *SpvBuilder) EmitTypeStruct(memberTypes ...SpirvId) SpirvId {
	id := b.AllocId()
	operands := []uint32{uint32(id)}
	for _, t := range memberTypes {
		operands = append(operands, uint32(t))
	}
	b.instr(&b.types, spec.SpvOpTypeStruct, operands...)
	return id
}

// EmitTypePointer declares a pointer to pointerType in the given storage class.
func (b *SpvBuilder) EmitTypePointer(storageClass uint32, pointerType SpirvId) SpirvId {
	id := b.AllocId()
	b.instr(&b.types, spec.SpvOpTypePointer, uint32(id), storageClass, uint32(pointerType))
	return id
}

// EmitTypeFunction emits OpTypeFunction.
func (b *SpvBuilder) EmitTypeFunction(returnType SpirvId, paramTypes ...SpirvId) SpirvId {
	id := b.AllocId()
	operands := []uint32{uint32(id), uint32(returnType)}
	for _, p := range paramTypes {
		operands = append(operands, uint32(p))
	}
	b.instr(&b.types, spec.SpvOpTypeFunction, operands...)
	return id
}
