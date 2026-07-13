package emu

import (
	"encoding/binary"
	"unsafe"
)

const ErrnoTcbOffset = 0x188

// GetErrnoAddress returns address of the errno variable for current thread.
func GetErrnoAddress() uintptr {
	thread := GetCurrentThread()
	return uintptr(unsafe.Pointer(thread.Tcb)) + ErrnoTcbOffset
}

// GetErrno returns error number of current thread.
func GetErrno() uintptr {
	errNoAddr := GetErrnoAddress()
	errNoSlice := unsafe.Slice((*byte)(unsafe.Pointer(errNoAddr)), 8)
	return uintptr(binary.LittleEndian.Uint64(errNoSlice))
}

// SetErrno sets error number for current thread.
func SetErrno(err uintptr) {
	errNoAddr := GetErrnoAddress()
	errNoSlice := unsafe.Slice((*byte)(unsafe.Pointer(errNoAddr)), 8)
	binary.LittleEndian.PutUint64(errNoSlice, uint64(err))
}
