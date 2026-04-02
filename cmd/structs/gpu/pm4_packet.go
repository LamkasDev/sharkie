package gpu

const (
	GNM_PREPARE_FLIP_MAGIC         = uint32(0xC03E1000)
	GNM_PREPARE_FLIP_VARIANT_BASE  = uint32(0x68750777)
	GNM_PREPARE_FLIP_VARIANT_ADDR  = uint32(0x68750778)
	GNM_PREPARE_FLIP_VARIANT_MAX   = uint32(0x68750781)
	GNM_PREPARE_FLIP_OFFSET_DWORDS = uint32(64)

	PM4_WRITE_DATA_HEADER  = uint32(0xC0033700)
	PM4_WRITE_DATA_CONTROL = uint32(0x00000500)
)

// NewPM4TypedHeader builds a type-3 packet header.
func NewPM4TypedHeader(opcode, numDWords uint32) uint32 {
	return 0xC0000000 | ((numDWords - 1) << 16) | (opcode << 8)
}
