package structs

//go:generate hsp
//go:generate go run ../../hsp_gen/hsp_gen.go -- spirv_emitter_gcn_buffer_internal_gen.go

import "github.com/cespare/xxhash"

type BufferDescriptor struct {
	BaseAddress   uintptr
	Stride        uint16
	SwizzleCache  bool
	SwizzleEnable bool

	NumRecords   uint32
	DstSelX      uint8
	DstSelY      uint8
	DstSelZ      uint8
	DstSelW      uint8
	NumFormat    uint8
	DataFormat   uint8
	ElementSize  uint8
	IndexStride  uint8
	AddTidEnable bool
	Atc          bool
	HashEnable   bool
	Heap         bool
	MType        uint8
	Type         uint8
}

func (z *BufferDescriptor) Hash() uint64 {
	data, _ := z.MarshalHash()
	return xxhash.Sum64(data)
}
