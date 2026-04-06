package structs

import (
	"encoding/binary"
	"unsafe"
)

// https://docs.particle.io/reference/device-os/api/debugging/posix-errors
const (
	EPERM        = 1
	ENOENT       = 2
	EAGAIN       = 11
	ENOMEM       = 12
	EFAULT       = 14
	EBUSY        = 16
	EINVAL       = 22
	ESPIPE       = 29
	EDEADLK      = 45
	ETIMEDOUT    = 60
	ENAMETOOLONG = 63
)

const ERR_PTR = ^uintptr(0)
const ERR_PTRI = -1
const ERR_HANDLE = ^uint32(0)

// ResolveHandle converts a guest handle (double pointer) into a host structs pointer.
// Returns the structs pointer and 0 on success, or nil and an error code (EINVAL).
func ResolveHandle[T any](handlePtr uintptr) (*T, uintptr) {
	if handlePtr == 0 {
		return nil, EINVAL
	}

	ptr := *(*uint64)(unsafe.Pointer(handlePtr))
	if ptr == 0 {
		return nil, EINVAL
	}

	return (*T)(unsafe.Pointer(uintptr(ptr))), 0
}

func IsPowerOfTwo(v uintptr) bool {
	return v != 0 && (v&(v-1)) == 0
}

func WriteAddress(addressPtr uintptr, address uintptr) {
	handleSlice := unsafe.Slice((*byte)(unsafe.Pointer(addressPtr)), 8)
	binary.LittleEndian.PutUint64(handleSlice, uint64(address))
}

func WriteResult(resultPtr uintptr, result uint32) {
	resultSlice := unsafe.Slice((*byte)(unsafe.Pointer(resultPtr)), 4)
	binary.LittleEndian.PutUint32(resultSlice, result)
}
