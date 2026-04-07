//go:build linux

package structs

import (
	"syscall"

	"github.com/LamkasDev/sharkie/cmd/sys_struct"
)

// MemoryProtToLinuxProt converts memory protection flags to Linux mmap/mprotect flags.
func MemoryProtToLinuxProt(prot int32) uintptr {
	return uintptr(prot & int32(syscall.PROT_READ|syscall.PROT_WRITE|syscall.PROT_EXEC))
}

// MemoryProtToLinuxProt converts memory flags to Linux mmap/mprotect flags.
func MemoryFlagsToLinuxFlags(flags int32, addr uintptr) uintptr {
	flags = flags&int32(syscall.MAP_SHARED|syscall.MAP_PRIVATE|syscall.MAP_FIXED) | syscall.MAP_ANONYMOUS
	if addr != 0 {
		flags |= syscall.MAP_FIXED
	}
	if flags&(syscall.MAP_SHARED|syscall.MAP_PRIVATE) == 0 {
		flags |= syscall.MAP_PRIVATE
	}

	return uintptr(flags)
}

func ReserveKernelMemory(addr uintptr, length uint64) (uintptr, error) {
	allocatedAddr, _, err := syscall.Syscall6(
		syscall.SYS_MMAP,
		addr,
		uintptr(length),
		uintptr(syscall.PROT_NONE),
		MemoryFlagsToLinuxFlags(0, addr),
		ERR_PTR,
		0,
	)
	if err != 0 {
		return 0, err
	}

	return allocatedAddr, nil
}

func AllocKernelMemory(addr uintptr, length uint64, prot, flags int32) (uintptr, error) {
	isDirectMemory, isGpuMemory := MemoryIsDirectOrGpu(addr)
	if isDirectMemory || isGpuMemory {
		if _, err := ProtectKernelMemory(addr, length, prot); err != nil {
			return 0, err
		}
		return addr, nil
	}
	addr = sys_struct.GetNextAlignedAddress(addr, length)
	allocatedAddr, _, err := syscall.Syscall6(
		syscall.SYS_MMAP,
		addr,
		uintptr(length),
		MemoryProtToLinuxProt(prot),
		MemoryFlagsToLinuxFlags(flags, addr),
		ERR_PTR,
		0,
	)
	if err != 0 {
		return 0, err
	}

	return allocatedAddr, nil
}

func FreeKernelMemory(addr uintptr, length uint64) (uintptr, error) {
	_, _, err := syscall.Syscall6(
		syscall.SYS_MMAP,
		addr,
		uintptr(length),
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

func ProtectKernelMemory(addr uintptr, length uint64, prot int32) (uintptr, error) {
	_, _, err := syscall.Syscall(
		syscall.SYS_MPROTECT,
		addr,
		uintptr(length),
		MemoryProtToLinuxProt(prot),
	)
	if err != 0 {
		return 0, err
	}

	return 1, nil
}
