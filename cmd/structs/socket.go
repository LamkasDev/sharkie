package structs

import "unsafe"

const (
	NETC_GET_MEM_INFO = 0x14

	SCE_NET_IOCTL_INIT = 0x802450C9
)

const SocketBufferSize = 1024 * 1024

type Socket struct {
	Name     string
	Domain   int32
	Type     int32
	Protocol int32
}

type NetworkMemoryInfo struct {
	BufferSize uint32
}

const NetworkMemoryInfoSize = unsafe.Sizeof(NetworkMemoryInfo{})
