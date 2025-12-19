package lib

import (
	"github.com/LamkasDev/sharkie/cmd/sys_struct"
	"golang.org/x/sys/windows"
)

func libKernel_alloc(addr, length, prot, flags uintptr) (uintptr, error) {
	allocatedAddr, _, err := sys_struct.VirtualAlloc.Call(
		addr,
		length,
		windows.MEM_RESERVE|windows.MEM_COMMIT,
		windows.PAGE_READWRITE,
	)

	return allocatedAddr, err
}
