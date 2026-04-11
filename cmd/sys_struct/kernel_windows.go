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
	VirtualProtect              = Kernel32.NewProc("VirtualProtect")
	TlsAlloc                    = Kernel32.NewProc("TlsAlloc")
	TlsSetValue                 = Kernel32.NewProc("TlsSetValue")
	TlsGetValue                 = Kernel32.NewProc("TlsGetValue")
	GetCurrentThread            = Kernel32.NewProc("GetCurrentThread")
	GetThreadContext            = Kernel32.NewProc("GetThreadContext")
	SetThreadContext            = Kernel32.NewProc("SetThreadContext")
	MapViewOfFileEx             = Kernel32.NewProc("MapViewOfFileEx")
	UnmapViewOfFile             = Kernel32.NewProc("UnmapViewOfFile")
)
