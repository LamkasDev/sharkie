package common

type ImageAccessKind uint8

const (
	ImageAccessLoad ImageAccessKind = iota
	ImageAccessLoadMip
	ImageAccessStore
	ImageAccessStoreMip
	ImageAccessSample
)

func (kind ImageAccessKind) IsImage() bool {
	switch kind {
	case ImageAccessLoad, ImageAccessLoadMip, ImageAccessStore, ImageAccessStoreMip:
		return true
	default:
		return false
	}
}

func (kind ImageAccessKind) IsRead() bool {
	switch kind {
	case ImageAccessLoad, ImageAccessLoadMip, ImageAccessSample:
		return true
	default:
		return false
	}
}

func (kind ImageAccessKind) Access() (BindingAccess, bool) {
	switch kind {
	case ImageAccessLoad, ImageAccessLoadMip, ImageAccessSample:
		return BindingAccessSampledRead, true
	case ImageAccessStore, ImageAccessStoreMip:
		return BindingAccessStorageWrite, true
	default:
		return 0, false
	}
}

type SgprSource struct {
	UserDataOffset int32 // -1 if unknown
}

type SpirvShaderResource struct {
	InstructionOffset uintptr
	Kind              ImageAccessKind
	RsrcUserData      int32 // UserData offset for T# base
	SampUserData      int32 // UserData offset for S# base
}
