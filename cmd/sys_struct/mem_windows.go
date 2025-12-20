//go:build windows

package sys_struct

import (
	"golang.org/x/sys/windows"
)

// AllocExecututableMemory allocates a chunk of executable memory with the defined size.
func AllocExecututableMemory(size uintptr) (uintptr, error) {
	addr, _, err := VirtualAlloc.Call(
		0,
		size,
		windows.MEM_COMMIT|windows.MEM_RESERVE,
		windows.PAGE_EXECUTE_READWRITE,
	)
	if addr == 0 {
		return 0, err
	}

	return addr, nil
}

// AllocReadWriteMemory allocates a chunk of read-write memory with the defined size.
func AllocReadWriteMemory(size uintptr) (uintptr, error) {
	addr, _, err := VirtualAlloc.Call(
		0,
		size,
		windows.MEM_COMMIT|windows.MEM_RESERVE,
		windows.PAGE_READWRITE,
	)
	if addr == 0 {
		return 0, err
	}

	return addr, nil
}

// FreeReadWriteMemory releases memory allocated by AllocReadWriteMemory.
func FreeReadWriteMemory(addr uintptr) error {
	size, _, err := VirtualFree.Call(
		addr,
		0,
		windows.MEM_RELEASE,
	)
	if size == 0 {
		return err
	}

	return nil
}
