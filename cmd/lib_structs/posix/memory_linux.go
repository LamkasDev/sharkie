//go:build linux

package posix

import (
	"fmt"
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

func AllocKernelMemory(addr uintptr, length uint64, prot, flags int32) (uintptr, error) {
	addr = sys_struct.GetNextAlignedAddress(addr, length)
	linuxProt := MemoryProtToLinuxProt(prot)
	linuxFlags := MemoryFlagsToLinuxFlags(flags, addr)
	for {
		allocatedAddr, _, err := syscall.Syscall6(
			syscall.SYS_MMAP,
			addr,
			uintptr(length),
			linuxProt,
			linuxFlags,
			ERR_PTR,
			0,
		)
		if err == 0 {
			if allocatedAddr < 0xFFFFFFFFFF {
				return allocatedAddr, nil
			}
			syscall.Syscall6(syscall.SYS_MUNMAP, allocatedAddr, uintptr(length), 0, 0, 0, 0)
		} else if err != syscall.EEXIST {
			return 0, err
		}
		addr = sys_struct.GetNextAlignedAddress(0, length)
		linuxFlags |= 0x100000
	}
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
