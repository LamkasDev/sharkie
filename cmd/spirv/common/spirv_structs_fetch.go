package common

type FetchAttribute struct {
	DestVgpr    uint32
	BufferIndex uint32
	Offset      uint32
	NumElements uint32
}

type FetchShaderLayout []FetchAttribute

func (l FetchShaderLayout) Hash() uint32 {
	hash := uint32(0)
	for _, attr := range l {
		hash ^= attr.DestVgpr
		hash = (hash << 5) | (hash >> 27)
		hash ^= attr.BufferIndex
		hash = (hash << 5) | (hash >> 27)
		hash ^= attr.Offset
		hash = (hash << 5) | (hash >> 27)
		hash ^= attr.NumElements
		hash = (hash << 5) | (hash >> 27)
	}
	return hash
}
