package structs

import "unsafe"

type DceCommand struct {
	CommandId uint32
	_         [4]byte
	Handle    uintptr
	Param1    uintptr
	Param2    uintptr
	Param3    uintptr
	_         [8]byte
}

const DceCommandSize = unsafe.Sizeof(DceCommand{})

type DceRegisterBuffers struct {
	CommandId uint32
	_         [4]byte
	Handle    uint32
	Index     uint32
	Address   uint64
	Size      uint64
	Flags     uint64
	_         [8]byte
}

const DceRegisterBuffersSize = unsafe.Sizeof(DceRegisterBuffers{})
