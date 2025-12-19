//go:build windows

package sys_struct

import (
	"syscall"
)

var (
	Kernel32                    = syscall.NewLazyDLL("kernel32.dll")
	AddVectoredExceptionHandler = Kernel32.NewProc("AddVectoredExceptionHandler")
	VirtualAlloc                = Kernel32.NewProc("VirtualAlloc")
	VirtualFree                 = Kernel32.NewProc("VirtualFree")
	TlsAlloc                    = Kernel32.NewProc("TlsAlloc")
	TlsSetValue                 = Kernel32.NewProc("TlsSetValue")
	TlsGetValue                 = Kernel32.NewProc("TlsGetValue")
)
