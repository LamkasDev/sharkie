package structs

import (
	"unsafe"
)

const (
	IMPI_METHOD_SERVICE_INIT = 0x0
	IMPI_METHOD_SERVICE_TERM = 0x1

	IMPI_METHOD_PING = 0x20000
)

type IpmiSyncMethod struct {
	MethodId   uint32
	InputSize  uint32
	OutputSize uint32
	_          [4]byte
	InputPtr   uintptr
	OutputPtr  uintptr
	ResultPtr  uintptr
	_          [8]byte
}

const IpmiSyncMethodSize = unsafe.Sizeof(IpmiSyncMethod{})
