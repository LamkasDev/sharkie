//go:build windows

package structs

import (
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/sys_struct"
	"golang.org/x/sys/windows"
)

// MemoryProtToWindowsProt converts memory protection flags to Windows VirtualAlloc flags.
func MemoryProtToWindowsProt(prot int32) uintptr {
	isRead := (prot&PROT_READ) != 0 || (prot&PROT_GPU_READ) != 0
	isWrite := (prot&PROT_WRITE) != 0 || (prot&PROT_GPU_WRITE) != 0
	isExec := (prot & PROT_EXEC) != 0

	switch {
	case isExec && isWrite:
		return windows.PAGE_EXECUTE_READWRITE
	case isExec:
		return windows.PAGE_EXECUTE_READ
	case isWrite:
		return windows.PAGE_READWRITE
	case isRead:
		return windows.PAGE_READONLY
	default:
		return windows.PAGE_NOACCESS
	}
}

func AllocKernelMemory(addr uintptr, length uint64, prot, flags int32) (uintptr, error) {
	allocationType := uintptr(windows.MEM_RESERVE | windows.MEM_COMMIT)
	addr = sys_struct.GetNextAlignedAddress(addr, length)
	allocatedAddr, _, err := sys_struct.VirtualAlloc.Call(
		addr,
		uintptr(length),
		allocationType,
		MemoryProtToWindowsProt(prot),
	)
	if allocatedAddr == 0 {
		return 0, err
	}

	return allocatedAddr, nil
}

func FreeKernelMemory(addr uintptr, length uint64) (uintptr, error) {
	ret, _, err := sys_struct.VirtualFree.Call(
		addr,
		uintptr(length),
		windows.MEM_DECOMMIT,
	)
	if ret == 0 {
		return 0, err
	}

	return ret, nil
}

func ProtectKernelMemory(addr uintptr, length uint64, prot int32) (uintptr, error) {
	var oldProt uint32
	ret, _, err := sys_struct.VirtualProtect.Call(
		addr,
		uintptr(length),
		MemoryProtToWindowsProt(prot),
		uintptr(unsafe.Pointer(&oldProt)),
	)
	if ret == 0 {
		return 0, err
	}

	return ret, nil
}

func MapVulkanMemory(addr uintptr, length uint64, handle uintptr) error {
	ret, _, err := sys_struct.UnmapViewOfFile.Call(addr)
	if ret == 0 {
		return err
	}
	allocatedAddr, _, err := sys_struct.MapViewOfFileEx.Call(
		handle,
		0xF001F, // FILE_MAP_ALL_ACCESS = 0xF001F (READ | WRITE | ...)
		0,
		0,
		uintptr(length),
		addr,
	)
	if allocatedAddr == 0 {
		return err
	}
	if allocatedAddr != addr {
		return fmt.Errorf("MapVulkanMemory: failed to map at fixed address")
	}

	return nil
}
