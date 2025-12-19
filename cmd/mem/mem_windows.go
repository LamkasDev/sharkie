//go:build windows

package mem

import (
	"github.com/LamkasDev/sharkie/cmd/sys_struct"
	"golang.org/x/sys/windows"
)

// AllocExecututableMemory allocates a chunk of executable memory with the defined size.
func AllocExecututableMemory(size uintptr) uintptr {
	addr, _, err := sys_struct.VirtualAlloc.Call(
		0,
		size,
		windows.MEM_COMMIT|windows.MEM_RESERVE,
		windows.PAGE_EXECUTE_READWRITE,
	)
	if addr == 0 {
		panic(err)
	}

	return addr
}

// AllocReadWriteMemory allocates a chunk of read-write memory with the defined size.
func AllocReadWriteMemory(size uintptr) uintptr {
	addr, _, err := sys_struct.VirtualAlloc.Call(
		0,
		size,
		windows.MEM_COMMIT|windows.MEM_RESERVE,
		windows.PAGE_READWRITE,
	)
	if addr == 0 {
		panic(err)
	}

	return addr
}

// FreeReadWriteMemory releases memory allocated by AllocReadWriteMemory.
func FreeReadWriteMemory(addr uintptr) {
	if addr == 0 {
		return
	}

	size, _, err := sys_struct.VirtualFree.Call(
		addr,
		0,
		windows.MEM_RELEASE,
	)
	if size == 0 {
		panic(err)
	}
}
