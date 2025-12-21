package structs

import "unsafe"

// https://docs.particle.io/reference/device-os/api/debugging/posix-errors
const EPERM = 1
const ENOENT = 2
const EAGAIN = 11
const ENOMEM = 12
const EFAULT = 14
const EINVAL = 22
const EDEADLK = 45
const ENAMETOOLONG = 63

const ERR_PTR = ^uintptr(0)

// ResolveHandle converts a guest handle (double pointer) into a host structs pointer.
// Returns the structs pointer and 0 on success, or nil and an error code (EINVAL).
func ResolveHandle[T any](handlePtr uintptr) (*T, int32) {
	if handlePtr == 0 {
		return nil, EINVAL
	}

	ptr := *(*uint64)(unsafe.Pointer(handlePtr))
	if ptr == 0 {
		return nil, EINVAL
	}

	return (*T)(unsafe.Pointer(uintptr(ptr))), 0
}

// ReadCString reads a C-style string with a NULL terminator from stringPtr.
func ReadCString(stringPtr uintptr) string {
	stringSlice := unsafe.Slice((*byte)(unsafe.Pointer(stringPtr)), 256)
	stringLength := 0
	for i, b := range stringSlice {
		if b == 0 {
			stringLength = i
			break
		}
	}

	return string(stringSlice[:stringLength])
}

// WriteCString writes a C-style string with NULL terminator to stringPtr.
func WriteCString(stringPtr uintptr, name string) {
	stringSlice := unsafe.Slice((*byte)(unsafe.Pointer(stringPtr)), 256)
	for i, b := range []byte(name) {
		stringSlice[i] = b
	}
	stringSlice[len(name)] = 0
}
