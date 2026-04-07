//go:build windows

package sys_struct

import (
	"golang.org/x/sys/windows"
)

// AllocExecutableMemory allocates a chunk of executable memory with the defined size.
func AllocExecutableMemory(size uintptr) (uintptr, error) {
	targetAddr := GetNextAlignedAddress(0, size)
	addr, _, err := VirtualAlloc.Call(
		targetAddr,
		size,
		windows.MEM_COMMIT|windows.MEM_RESERVE,
		windows.PAGE_EXECUTE_READWRITE,
	)
	if addr == 0 {
		return 0, err
	}

	return addr, nil
}
