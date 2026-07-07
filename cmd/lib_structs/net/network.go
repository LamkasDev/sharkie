package net

import "unsafe"

const (
	SCE_NET_CTL_ERROR_INVALID_ADDR = 0x80412107
)

const (
	NETC_GET_MEM_INFO = 0x14
)

type NetworkMemoryInfo struct {
	BufferSize uint32
}

const NetworkMemoryInfoSize = unsafe.Sizeof(NetworkMemoryInfo{})
