package spirv

type ImageDescriptor struct {
	BaseAddress uintptr
	MinLod      uint16
	DataFormat  uint8
	NumFormat   uint8
	Width       uint32
	Height      uint32
	Type        uint8
}

func NewImageDescriptor(dwords []uint32) ImageDescriptor {
	d := ImageDescriptor{}

	// DW0 & DW1: Base Address (40 bits) + Min Lod (12 bits)
	d.BaseAddress = uintptr(dwords[0]) | (uintptr(dwords[1]&0xFF) << 32)
	d.MinLod = uint16((dwords[1] >> 8) & 0xFFF)

	// DW1: Data Format (6 bits) + Num Format (4 bits)
	d.DataFormat = uint8((dwords[1] >> 20) & 0x3F)
	d.NumFormat = uint8((dwords[1] >> 26) & 0xF)

	// DW2: Width (14 bits) + Height (14 bits)
	d.Width = (dwords[2] & 0x3FFF) + 1
	d.Height = ((dwords[2] >> 14) & 0x3FFF) + 1

	// DW3: Type (4 bits)
	d.Type = uint8((dwords[3] >> 28) & 0xF)

	return d
}
