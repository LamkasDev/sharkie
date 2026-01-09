//go:build windows

package structs

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/sys_struct"
	"golang.org/x/sys/windows"
)

// MemoryProtToWindowsProt converts memory protection flags to Windows VirtualAlloc flags.
func MemoryProtToWindowsProt(prot uintptr) uintptr {
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

func ReserveKernelMemory(addr, length uintptr) (uintptr, error) {
	allocatedAddr, _, err := sys_struct.VirtualAlloc.Call(
		addr,
		length,
		windows.MEM_RESERVE,
		windows.PAGE_NOACCESS,
	)
	if allocatedAddr == 0 {
		return 0, err
	}

	return allocatedAddr, nil
}

func AllocKernelMemory(addr, length, prot, flags uintptr) (uintptr, error) {
	allocationType := uintptr(windows.MEM_COMMIT)
	isDirectMemory, isGpuMemory := MemoryIsDirectOrGpu(addr)
	if !isDirectMemory && !isGpuMemory {
		allocationType |= windows.MEM_RESERVE
	}
	allocatedAddr, _, err := sys_struct.VirtualAlloc.Call(
		addr,
		length,
		allocationType,
		MemoryProtToWindowsProt(prot),
	)
	if allocatedAddr == 0 {
		return 0, err
	}

	return allocatedAddr, nil
}

func FreeKernelMemory(addr, length uintptr) (uintptr, error) {
	ret, _, err := sys_struct.VirtualFree.Call(
		addr,
		length,
		windows.MEM_DECOMMIT,
	)
	if ret == 0 {
		return 0, err
	}

	return ret, nil
}

func ProtectKernelMemory(addr, length, prot uintptr) (uintptr, error) {
	var oldProt uint32
	ret, _, err := sys_struct.VirtualProtect.Call(
		addr,
		length,
		MemoryProtToWindowsProt(prot),
		uintptr(unsafe.Pointer(&oldProt)),
	)
	if ret == 0 {
		return 0, err
	}

	return ret, nil
}
