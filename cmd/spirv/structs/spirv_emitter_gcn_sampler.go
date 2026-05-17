package structs

type SamplerDescriptor struct {
	ClampX    uint8
	ClampY    uint8
	ClampZ    uint8
	MagFilter uint8
	MinFilter uint8
	MipFilter uint8
}

func NewSamplerDescriptor(dwords []uint32) SamplerDescriptor {
	s := SamplerDescriptor{}

	// DW0: Clamp X (3 bits), Y (3 bits), Z (3 bits)
	s.ClampX = uint8(dwords[0] & 0x7)
	s.ClampY = uint8((dwords[0] >> 3) & 0x7)
	s.ClampZ = uint8((dwords[0] >> 6) & 0x7)

	// DW2: Filter Mag (2 bits), Min (2 bits), Mip (2 bits)
	s.MagFilter = uint8((dwords[2] >> 20) & 0x3)
	s.MinFilter = uint8((dwords[2] >> 22) & 0x3)
	s.MipFilter = uint8((dwords[2] >> 26) & 0x3)

	return s
}
