package structs

import "github.com/cespare/xxhash"

func NewSamplerDescriptor(dwords []uint32) *SamplerDescriptor {
	isSamplerZero := true
	for i := range 4 {
		if dwords[i] != 0 {
			isSamplerZero = false
		}
	}
	if isSamplerZero {
		return nil
	}

	rawLodBias := int32(dwords[2] & 0x1FFF)
	if (rawLodBias & 0x1000) != 0 {
		rawLodBias |= ^0x1FFF
	}
	rawLodBiasSec := int8((dwords[2] >> 14) & 0x1F)
	if (rawLodBiasSec & 0x10) != 0 {
		rawLodBiasSec |= ^0x1F
	}
	return &SamplerDescriptor{
		Dwords: [4]uint32(dwords),

		ClampX:            uint8(dwords[0] & 0x7),
		ClampY:            uint8((dwords[0] >> 3) & 0x7),
		ClampZ:            uint8((dwords[0] >> 6) & 0x7),
		MaxAnisoRatio:     uint8((dwords[0] >> 9) & 0x7),
		DepthCompareFunc:  uint8((dwords[0] >> 12) & 0x7),
		ForceUnnormalized: (dwords[0]>>15)&1 == 1,
		AnisoThreshold:    uint8((dwords[0] >> 16) & 0x7),
		McCoordTrunc:      (dwords[0]>>19)&1 == 1,
		ForceDegamma:      (dwords[0]>>20)&1 == 1,
		AnisoBias:         float32((dwords[0]>>21)&0x3F) / 32.0,
		TruncCoord:        (dwords[0]>>27)&1 == 1,
		DisableCubeWrap:   (dwords[0]>>28)&1 == 1,
		FilterMode:        uint8((dwords[0] >> 29) & 0x3),

		MinLod:  float32(uint16(dwords[1]&0xFFF)) / 256.0,
		MaxLod:  float32(uint16((dwords[1]>>12)&0xFFF)) / 256.0,
		PerfMip: uint8((dwords[1] >> 24) & 0xF),
		PerfZ:   uint8((dwords[1] >> 28) & 0xF),

		LodBias:          float32(rawLodBias) / 256.0,
		LodBiasSec:       float32(rawLodBiasSec) / 16.0,
		XyMagFilter:      uint8((dwords[2] >> 20) & 0x3),
		XyMinFilter:      uint8((dwords[2] >> 22) & 0x3),
		ZFilter:          uint8((dwords[2] >> 24) & 0x3),
		MipFilter:        uint8((dwords[2] >> 26) & 0x3),
		MipPointPreclamp: (dwords[2]>>28)&1 == 1,
		DisableLsbCeil:   (dwords[2]>>29)&1 == 1,

		BorderColorPtr:  uint16(dwords[3] & 0xFFF),
		BorderColorType: uint8((dwords[3] >> 30) & 0x3),
	}
}

func (z *SamplerDescriptor) Hash() uint64 {
	data, _ := z.MarshalHash()
	return xxhash.Sum64(data)
}
