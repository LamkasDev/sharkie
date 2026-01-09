//go:build linux

package structs

import (
	"syscall"
)

func ReserveKernelMemory(addr, length uintptr) (uintptr, error) {
	flags := syscall.MAP_PRIVATE | syscall.MAP_ANONYMOUS
	if addr != 0 {
		flags |= syscall.MAP_FIXED
	}
	allocatedAddr, _, err := syscall.Syscall6(
		syscall.SYS_MMAP,
		addr,
		length,
		uintptr(syscall.PROT_NONE),
		uintptr(flags),
		ERR_PTR,
		0,
	)
	if err != 0 {
		return 0, err
	}

	return allocatedAddr, nil
}

func AllocKernelMemory(addr, length, prot, flags uintptr) (uintptr, error) {
	isDirectMemory, isGpuMemory := MemoryIsDirectOrGpu(addr)
	if isDirectMemory || isGpuMemory {
		if _, err := ProtectKernelMemory(addr, length, flags); err != nil {
			return 0, err
		}
		return addr, nil
	}
	flags |= syscall.MAP_ANONYMOUS
	if addr != 0 {
		flags |= syscall.MAP_FIXED
	}
	allocatedAddr, _, err := syscall.Syscall6(
		syscall.SYS_MMAP,
		addr,
		length,
		prot,
		flags,
		ERR_PTR,
		0,
	)
	if err != 0 {
		return 0, err
	}

	return allocatedAddr, nil
}

func FreeKernelMemory(addr, length uintptr) (uintptr, error) {
	_, _, err := syscall.Syscall6(
		syscall.SYS_MMAP,
		addr,
		length,
		uintptr(syscall.PROT_NONE),
		uintptr(syscall.MAP_PRIVATE|syscall.MAP_ANONYMOUS|syscall.MAP_FIXED),
		ERR_PTR,
		0,
	)
	if err != 0 {
		return 0, err
	}

	return addr, nil
}

func ProtectKernelMemory(addr, length, prot uintptr) (uintptr, error) {
	_, _, err := syscall.Syscall(
		syscall.SYS_MPROTECT,
		addr,
		length,
		prot,
	)
	if err != 0 {
		return 0, err
	}

	return 1, nil
}
