package common

import (
	"github.com/cespare/xxhash"
)

func (z *SpirvVertexShaderContext) Hash() uint64 {
	data, _ := z.MarshalHash()
	return xxhash.Sum64(data)
}

func (z *SpirvFragmentShaderContext) Hash() uint64 {
	data, _ := z.MarshalHash()
	return xxhash.Sum64(data)
}

func (z *SpirvComputeShaderContext) Hash() uint64 {
	data, _ := z.MarshalHash()
	return xxhash.Sum64(data)
}
