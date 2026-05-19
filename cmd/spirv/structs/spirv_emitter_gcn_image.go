package structs

import "github.com/cespare/xxhash"

type ImageNoSamplerKey [8]uint32
type ImageSamplerKey [12]uint32

func NewImageDescriptor(dwords []uint32) ImageDescriptor {
	return ImageDescriptor{
		BaseAddress:    ((uintptr(dwords[0]) | (uintptr(dwords[1]&0xFF) << 32)) << 8) & 0xFFFFFFFFFF,
		MinLod:         uint16((dwords[1] >> 8) & 0xFFF),
		DataFormat:     uint8((dwords[1] >> 20) & 0x3F),
		NumFormat:      uint8((dwords[1] >> 26) & 0xF),
		MType:          uint8((dwords[1]>>30)&0x3) | uint8(((dwords[3]>>26)&0x1)<<2),
		Width:          uint16(dwords[2]&0x3FFF) + 1,
		Height:         uint16((dwords[2]>>14)&0x3FFF) + 1,
		PerfModulation: uint8((dwords[2] >> 28) & 0x7),
		Interlaced:     (dwords[2]>>31)&0b1 == 1,
		DstSelX:        uint8(dwords[3] & 0x7),
		DstSelY:        uint8((dwords[3] >> 3) & 0x7),
		DstSelZ:        uint8((dwords[3] >> 6) & 0x7),
		DstSelW:        uint8((dwords[3] >> 9) & 0x7),
		BaseLevel:      uint8((dwords[3] >> 12) & 0xF),
		LastLevel:      uint8((dwords[3] >> 16) & 0xF),
		TilingIndex:    uint8((dwords[3] >> 20) & 0x1F),
		Pow2Pad:        (dwords[3]>>25)&0b1 == 1,
		Atc:            (dwords[3]>>27)&0b1 == 1,
		Type:           uint8((dwords[3] >> 28) & 0xF),

		Depth:      uint16(dwords[4]&0x1FFF) + 1,
		Pitch:      uint16((dwords[4]>>13)&0x3FFF) + 1,
		BaseArray:  uint16(dwords[5] & 0x1FFF),
		LastArray:  uint16((dwords[5] >> 13) & 0x1FFF),
		MinLodWarn: uint16(dwords[6] & 0xFFF),
	}
}

func (z *ImageDescriptor) Hash() uint64 {
	data, _ := z.MarshalHash()
	return xxhash.Sum64(data)
}
