//go:build linux

package posix

import (
	"fmt"
	"syscall"
)

// MemoryProtToLinuxProt converts memory protection flags to Linux mmap/mprotect flags.
func MemoryProtToLinuxProt(prot int32) uintptr {
	linuxProt := uintptr(0)
	if (prot&PROT_READ) != 0 || (prot&PROT_GPU_READ) != 0 {
		linuxProt |= syscall.PROT_READ
	}
	if (prot&PROT_WRITE) != 0 || (prot&PROT_GPU_WRITE) != 0 {
		linuxProt |= syscall.PROT_WRITE
	}
	if (prot & PROT_EXEC) != 0 {
		linuxProt |= syscall.PROT_EXEC
	}

	return linuxProt
}

// MemoryFlagsToLinuxFlags converts memory flags to Linux mmap/mprotect flags.
func MemoryFlagsToLinuxFlags(flags int32, addr uintptr) uintptr {
	linuxFlags := flags&int32(syscall.MAP_SHARED|syscall.MAP_PRIVATE|syscall.MAP_FIXED) | syscall.MAP_ANONYMOUS
	if linuxFlags&(syscall.MAP_SHARED|syscall.MAP_PRIVATE) == 0 {
		linuxFlags |= syscall.MAP_PRIVATE
	}

	return uintptr(linuxFlags)
}

func AllocKernelMemory(addr uintptr, length uint64, prot, flags int32, alignment uintptr) (uintptr, error) {
	addr = GetNextAlignedAddress(addr, length, alignment)
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

func MapVulkanMemory(addr uintptr, length uint64, fd uintptr, backingOffset uint64) error {
	if _, _, err := syscall.Syscall(syscall.SYS_MUNMAP, addr, uintptr(length), 0); err != 0 && err != syscall.EINVAL {
		return err
	}
	allocatedAddr, _, err := syscall.Syscall6(
		syscall.SYS_MMAP,
		addr,
		uintptr(length),
		uintptr(syscall.PROT_READ|syscall.PROT_WRITE),
		uintptr(syscall.MAP_SHARED|syscall.MAP_FIXED),
		fd,
		uintptr(backingOffset),
	)
	if err != 0 {
		return err
	}
	if allocatedAddr != addr {
		return fmt.Errorf("MapVulkanMemory: failed to map at fixed address")
	}

	return nil
}
