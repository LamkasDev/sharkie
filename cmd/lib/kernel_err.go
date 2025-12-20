package lib

import (
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
)

const ErrnoTcbOffset = 0x188

func GetErrnoAddress() uintptr {
	tcbAddr := uintptr(unsafe.Pointer(emu.GlobalModuleManager.Tcb))
	return tcbAddr + ErrnoTcbOffset
}

func GetErrno() uintptr {
	errNoAddr := GetErrnoAddress()
	errNoSlice := unsafe.Slice((*byte)(unsafe.Pointer(errNoAddr)), 8)
	return uintptr(binary.LittleEndian.Uint64(errNoSlice))
}

func SetErrno(err uintptr) {
	errNoAddr := GetErrnoAddress()
	errNoSlice := unsafe.Slice((*byte)(unsafe.Pointer(errNoAddr)), 8)
	binary.LittleEndian.PutUint64(errNoSlice, uint64(err))
}

// 0x0000000000002C70
// void *_error()
func libKernel___error() uintptr {
	return GetErrno()
}
