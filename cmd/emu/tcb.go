package emu

import (
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/linker"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/LamkasDev/sharkie/cmd/sys_struct"
	"github.com/gookit/color"
)

// NewTCB creates a new instance of Tcb based on passed Elf.
func NewTCB(l *linker.Linker) *Tcb {
	maxTlsIndex := uintptr(len(GlobalModuleManager.ModulesMap))
	tcbSize := unsafe.Sizeof(Tcb{})
	dtvSize := unsafe.Sizeof(DtvEntry{}) * (maxTlsIndex + 2)
	threadSize := unsafe.Sizeof(Pthread{})

	tlsSize := uintptr(l.StaticTlsSize)
	padding := (TcbAlignment - (tlsSize % TcbAlignment)) % TcbAlignment
	tcbOffset := tlsSize + padding
	totalSize := tcbOffset + tcbSize

	addr, _ := sys_struct.AllocReadWriteMemory(totalSize)
	tcb := (*Tcb)(unsafe.Pointer(addr + tcbOffset))
	tcb.Self = tcb

	dtvAddr, _ := sys_struct.AllocReadWriteMemory(dtvSize)
	tcb.Dtv = (*DtvEntry)(unsafe.Pointer(dtvAddr))
	threadAddr, _ := sys_struct.AllocReadWriteMemory(threadSize)
	tcb.Thread = (*Pthread)(unsafe.Pointer(threadAddr))
	tcb.Fiber = 0

	dtvSlice := unsafe.Slice(tcb.Dtv, maxTlsIndex+2)
	dtvSlice[0].Counter = l.GenerationCounter
	dtvSlice[1].Counter = maxTlsIndex

	tcb.Thread.ThreadId = 1337
	tcb.Thread.Flags = 0
	tcb.Thread.ReturnValue = 0
	tcb.Thread.Error = 0
	tcb.Thread.CleanupHandlerStack = 0
	copy(tcb.Thread.Name[:], "MainThread")

	for _, module := range GlobalModuleManager.ModulesMap {
		if module.TlsSection == nil || module.TlsSection.InitImageSize == 0 {
			continue
		}
		src := uintptr(unsafe.Pointer(&module.Memory[0])) + uintptr(module.TlsSection.ImageVirtualAddress)
		dest := addr + uintptr(module.TlsSection.Offset)
		copy(
			unsafe.Slice((*byte)(unsafe.Pointer(dest)), module.TlsSection.InitImageSize),
			unsafe.Slice((*byte)(unsafe.Pointer(src)), module.TlsSection.InitImageSize),
		)
		dtvSlice[module.ModuleIndex+1].Pointer = dest
		TlsBaseRepo[module.ModuleIndex] = dest

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

	tcb := (*Tcb)(unsafe.Pointer(tcbAddr))
	return (uintptr)(unsafe.Pointer(tcb.Thread))
}
