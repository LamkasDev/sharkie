package net

import "unsafe"

const (
	NETC_GET_MEM_INFO = 0x14
)

type NetworkMemoryInfo struct {
	BufferSize uint32
}

const NetworkMemoryInfoSize = unsafe.Sizeof(NetworkMemoryInfo{})
