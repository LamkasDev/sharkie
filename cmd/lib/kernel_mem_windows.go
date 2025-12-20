package lib

import (
	"unsafe"

	. "github.com/LamkasDev/sharkie/cmd/structs"
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

func libKernel_alloc(addr, length, prot, flags uintptr) (uintptr, error) {
	allocatedAddr, _, err := sys_struct.VirtualAlloc.Call(
		addr,
		length,
		windows.MEM_RESERVE|windows.MEM_COMMIT,
		MemoryProtToWindowsProt(prot),
	)

	return allocatedAddr, err
}

func libKernel_protect(addr, length, prot uintptr) (uintptr, error) {
	var oldProt uint32
	ret, _, err := sys_struct.VirtualProtect.Call(
		addr,
		length,
		prot,
		uintptr(unsafe.Pointer(&oldProt)),
	)

	return ret, err
}
