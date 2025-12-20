package emu

import (
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/linker"
	_struct "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/LamkasDev/sharkie/cmd/sys_struct"
	"github.com/gookit/color"
)

// NewTCB creates a new instance of Tcb based on passed Elf.
func NewTCB(l *linker.Linker) *_struct.Tcb {
	tcbSize := unsafe.Sizeof(_struct.Tcb{})
	dtvSize := unsafe.Sizeof(_struct.DtvEntry{}) * (l.MaxTlsIndex + 2)
	threadSize := unsafe.Sizeof(_struct.Pthread{})
	totalSize := l.StaticTlsSize + uint64(tcbSize)

	addr, _ := sys_struct.AllocReadWriteMemory(uintptr(totalSize))
	tcb := (*_struct.Tcb)(unsafe.Pointer(addr + uintptr(l.StaticTlsSize)))
	tcb.Self = tcb

	dtvAddr, _ := sys_struct.AllocReadWriteMemory(dtvSize)
	tcb.Dtv = (*_struct.DtvEntry)(unsafe.Pointer(dtvAddr))
	threadAddr, _ := sys_struct.AllocReadWriteMemory(threadSize)
	tcb.Thread = (*_struct.Pthread)(unsafe.Pointer(threadAddr))
	tcb.Fiber = 0

	dtvSlice := unsafe.Slice(tcb.Dtv, l.MaxTlsIndex+2)
	dtvSlice[0].Counter = l.GenerationCounter
	dtvSlice[1].Counter = l.MaxTlsIndex

	tcb.Thread.ThreadId = 1337
	tcb.Thread.Flags = 0
	tcb.Thread.ReturnValue = 0
	tcb.Thread.Error = 0
	tcb.Thread.CleanupHandlerStack = 0
	copy(tcb.Thread.Name[:], "MainThread")

	for _, module := range GlobalModuleManager.Modules {
		if module.TlsSection == nil || module.TlsSection.InitImageSize == 0 {
			continue
		}
		src := uintptr(unsafe.Pointer(&module.Memory[0])) + uintptr(module.TlsSection.ImageVirtualAddress)
		dest := addr + uintptr(module.TlsSection.Offset)
		copy(
			unsafe.Slice((*byte)(unsafe.Pointer(dest)), module.TlsSection.InitImageSize),
			unsafe.Slice((*byte)(unsafe.Pointer(src)), module.TlsSection.InitImageSize),
		)
		if module.TlsSection.ModuleIndex > 0 && module.TlsSection.ModuleIndex <= uint64(l.MaxTlsIndex) {
			dtvSlice[module.TlsSection.ModuleIndex+1].Pointer = dest
		}

		fmt.Printf(
			"%s's PT_TLS data from %s loaded into TCB at %s (%s bytes).\n",
			color.Blue.Sprint(module.Name),
			color.Yellow.Sprintf("0x%X", module.TlsSection.ImageVirtualAddress),
			color.Yellow.Sprintf("0x%X", dest),
			color.Gray.Sprintf("%d", module.TlsSection.InitImageSize),
		)
	}

	return tcb
}

// GetCurrentThread returns pointer to the Pthread structs of the current thread.
func GetCurrentThread() uintptr {
	tcbAddr, _, _ := sys_struct.TlsGetValue.Call(sys_struct.TlsSlot)
	if tcbAddr == 0 {
		return 0
	}

	tcb := (*_struct.Tcb)(unsafe.Pointer(tcbAddr))
	return (uintptr)(unsafe.Pointer(tcb.Thread))
}
