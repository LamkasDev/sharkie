package common

type ImageAccessKind uint8

const (
	ImageAccessUnknown ImageAccessKind = iota
	ImageAccessLoad
	ImageAccessStore
	ImageAccessSample
)

func (kind ImageAccessKind) Access() (BindingAccess, bool) {
	switch kind {
	case ImageAccessLoad, ImageAccessSample:
		return BindingAccessSampledRead, true
	case ImageAccessStore:
		return BindingAccessStorageWrite, true
	default:
		return 0, false
	}
}

type BufferAccessKind uint8

const (
	BufferAccessUnknown BufferAccessKind = iota
	BufferAccessLoad
	BufferAccessStore
)

type SgprSource struct {
	UserDataOffset int32 // -1 if unknown
}

type SpirvShaderResource struct {
	InstructionOffset uintptr
	Kind              ImageAccessKind
}
