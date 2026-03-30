package gpu

const (
	PM4_IT_NOP                  = 0x10
	PM4_IT_INDIRECT_BUFFER_CNST = 0x33
	PM4_IT_WRITE_DATA           = 0x37
	PM4_IT_INDIRECT_BUFFER      = 0x3F
	PM4_IT_EVENT_WRITE_EOP      = 0x47
)

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
