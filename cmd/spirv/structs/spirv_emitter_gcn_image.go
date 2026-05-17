package structs

type ImageNoSamplerKey [8]uint32
type ImageSamplerKey [12]uint32

type ImageIdentityKey struct {
	BaseAddress uintptr
	DataFormat  uint8
	NumFormat   uint8
	Width       uint32
	Height      uint32
	Type        uint8
	BaseLevel   uint8
	LastLevel   uint8
	DstSelX     uint8
	DstSelY     uint8
	DstSelZ     uint8
	DstSelW     uint8
}

type ImageDescriptor struct {
	BaseAddress uintptr
	MinLod      uint16
	DataFormat  uint8
	NumFormat   uint8
	MType       uint8
	Width       uint32
	Height      uint32
	Type        uint8
	BaseLevel   uint8
	LastLevel   uint8
	DstSelX     uint8
	DstSelY     uint8
	DstSelZ     uint8
	DstSelW     uint8
	TilingIndex uint8
	Pow2Pad     bool
	Atc         bool
	Depth       uint32
	Pitch       uint32
	BaseArray   uint32
	LastArray   uint32
}

func NewImageDescriptor(dwords []uint32) ImageDescriptor {
	d := ImageDescriptor{}

	// DW0 & DW1: Base Address (40 bits) + Min Lod (12 bits)
	d.BaseAddress = (uintptr(dwords[0]) | (uintptr(dwords[1]&0xFF) << 32)) << 8
	d.BaseAddress &= 0xFFFFFFFFFF
	d.MinLod = uint16((dwords[1] >> 8) & 0xFFF)

	// DW1: Data Format (6 bits) + Num Format (4 bits)
	d.DataFormat = uint8((dwords[1] >> 20) & 0x3F)
	d.NumFormat = uint8((dwords[1] >> 26) & 0xF)
	d.MType = uint8((dwords[1] >> 30) & 0x3)

	// DW2: Width (14 bits) + Height (14 bits)
	d.Width = (dwords[2] & 0x3FFF) + 1
	d.Height = ((dwords[2] >> 14) & 0x3FFF) + 1

	// DW3: DstSel (3 bits each), Base/Last level and Type (4 bits)
	d.DstSelX = uint8(dwords[3] & 0x7)
	d.DstSelY = uint8((dwords[3] >> 3) & 0x7)
	d.DstSelZ = uint8((dwords[3] >> 6) & 0x7)
	d.DstSelW = uint8((dwords[3] >> 9) & 0x7)
	d.Type = uint8((dwords[3] >> 28) & 0xF)
	d.BaseLevel = uint8((dwords[3] >> 12) & 0xF)
	d.LastLevel = uint8((dwords[3] >> 16) & 0xF)
	d.TilingIndex = uint8((dwords[3] >> 20) & 0x1F)
	d.Pow2Pad = ((dwords[3] >> 25) & 0x1) != 0
	d.MType |= uint8(((dwords[3] >> 26) & 0x1) << 2)
	d.Atc = ((dwords[3] >> 27) & 0x1) != 0

	// DW4/DW5 are present for 256-bit resources.
	d.Depth = (dwords[4] & 0x1FFF) + 1
	d.Pitch = ((dwords[4] >> 13) & 0x3FFF) + 1
	d.BaseArray = dwords[5] & 0x1FFF
	d.LastArray = (dwords[5] >> 13) & 0x1FFF

	return d
}

func NewImageIdentityKey(d ImageDescriptor) ImageIdentityKey {
	return ImageIdentityKey{
		BaseAddress: d.BaseAddress,
		DataFormat:  d.DataFormat,
		NumFormat:   d.NumFormat,
		Width:       d.Width,
		Height:      d.Height,
		Type:        d.Type,
		BaseLevel:   d.BaseLevel,
		LastLevel:   d.LastLevel,
		DstSelX:     d.DstSelX,
		DstSelY:     d.DstSelY,
		DstSelZ:     d.DstSelZ,
		DstSelW:     d.DstSelW,
	}
}
