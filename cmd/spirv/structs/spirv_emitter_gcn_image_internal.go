package structs

//go:generate hsp
//go:generate go run ../../hsp_gen/hsp_gen.go -- spirv_emitter_gcn_image_internal_gen.go

type ImageDescriptor struct {
	Dwords [8]uint32

	// 128-bit resource.
	BaseAddress    uintptr
	MinLod         uint16
	DataFormat     uint8
	NumFormat      uint8
	MType          uint8
	Width          uint16
	Height         uint16
	PerfModulation uint8
	Interlaced     bool
	DstSelX        uint8
	DstSelY        uint8
	DstSelZ        uint8
	DstSelW        uint8
	BaseLevel      uint8
	LastLevel      uint8
	TilingIndex    uint8
	Pow2Pad        bool
	Atc            bool
	Type           uint8

	// 256-bit resource.
	Depth      uint16
	Pitch      uint16
	BaseArray  uint16
	LastArray  uint16
	MinLodWarn uint16
}
