package structs

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/sys_struct"
	"golang.org/x/sys/windows"
)

// MemoryProtToWindowsProt converts memory protection flags to Windows VirtualAlloc flags.
func MemoryProtToWindowsProt(prot uintptr) uintptr {
	isRead := (prot & PROT_READ) != 0
	isWrite := (prot & PROT_WRITE) != 0
	isExec := (prot & PROT_EXEC) != 0

	switch {
	case isExec && isRead && isWrite:
		return windows.PAGE_EXECUTE_READWRITE
	case isExec && isRead:
		return windows.PAGE_EXECUTE_READ
	case isExec:
		return windows.PAGE_EXECUTE
	case isRead && isWrite:
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
	allocationType := uintptr(windows.MEM_RESERVE | windows.MEM_COMMIT)
	if addr != 0 &&
		addr >= GlobalAllocator.DirectMemoryBase &&
		addr < GlobalAllocator.DirectMemoryBase+GlobalAllocator.DirectMemorySize {
		allocationType = windows.MEM_COMMIT
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
