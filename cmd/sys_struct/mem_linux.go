//go:build linux

package sys_struct

import (
	"syscall"
)

// AllocExecutableMemory allocates a chunk of executable memory with the defined size.
func AllocExecutableMemory(size uintptr) (uintptr, error) {
	addr, _, err := syscall.Syscall6(
		syscall.SYS_MMAP,
		0,
		size,
		syscall.PROT_READ|syscall.PROT_WRITE|syscall.PROT_EXEC,
		syscall.MAP_PRIVATE|syscall.MAP_ANONYMOUS,
		^uintptr(0),
		0,
	)
	if err != 0 {
		return 0, err
	}

	return addr, nil
}
