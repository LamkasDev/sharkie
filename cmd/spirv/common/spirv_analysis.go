package common

import gcnSpec "github.com/LamkasDev/sharkie/cmd/lib_structs/gcn/spec"

type SpirvShaderResourceType uint8

const (
	SpirvShaderResourceTypeImage SpirvShaderResourceType = iota
	SpirvShaderResourceTypeBuffer
	SpirvShaderResourceTypeMemory
)

type ResourceAccessKind interface {
	isResourceAccessKind()
}

type ImageAccessKind uint8

const (
	ImageAccessUnknown ImageAccessKind = iota
	ImageAccessLoad
	ImageAccessStore
	ImageAccessSample
)

func (kind ImageAccessKind) isResourceAccessKind() {}

func (kind ImageAccessKind) String() string {
	switch kind {
	case ImageAccessLoad:
		return "load"
	case ImageAccessStore:
		return "store"
	case ImageAccessSample:
		return "sample"
	}
	return "??"
}

type BufferAccessKind uint8

const (
	BufferAccessUnknown BufferAccessKind = iota
	BufferAccessLoad
	BufferAccessStore
)

func (kind BufferAccessKind) isResourceAccessKind() {}

func (kind BufferAccessKind) String() string {
	switch kind {
	case BufferAccessLoad:
		return "load"
	case BufferAccessStore:
		return "store"
	}
	return "??"
}

type MemoryAccessKind uint8

const (
	MemoryAccessUnknown MemoryAccessKind = iota
	MemoryAccessLoad
)

func (kind MemoryAccessKind) isResourceAccessKind() {}

func (kind MemoryAccessKind) String() string {
	switch kind {
	case MemoryAccessLoad:
		return "load"
	}
	return "??"
}

type SgprSource struct {
	UserDataOffset int32 // -1 if unknown
}

type SpirvShaderResource struct {
	Instruction *gcnSpec.Instruction
	Type        SpirvShaderResourceType
	Kind        ResourceAccessKind
}
