//go:build windows

package sys_struct

import (
	"golang.org/x/sys/windows"
)

// AllocExecutableMemory allocates a chunk of executable memory with the defined size.
func AllocExecutableMemory(size uintptr) (uintptr, error) {
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
