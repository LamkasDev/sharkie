package emu

import (
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/linker"
	"github.com/LamkasDev/sharkie/cmd/mem"
	"github.com/LamkasDev/sharkie/cmd/sys_struct"
	"github.com/gookit/color"
)

// DtvEntry represent an entry in a dynamic thread vector.
// https://github.com/shadps4-emu/shadPS4/blob/9e287564ced1c7d84a5a165ce4ad6ba85d561ee1/src/core/tls.h#L22
type DtvEntry struct {
	Counter uintptr
	Pointer uintptr
}

// Tcb represent the thread control block used by a thread.
// https://github.com/shadps4-emu/shadPS4/blob/9e287564ced1c7d84a5a165ce4ad6ba85d561ee1/src/core/tls.h#L27
type Tcb struct {
	Self   *Tcb
	Dtv    *DtvEntry
	Thread uintptr
	Fiber  uintptr
}

// NewTCB creates a new instance of Tcb based on passed Elf.
func NewTCB(l *linker.Linker) *Tcb {
	tcbSize := unsafe.Sizeof(Tcb{})
	dtvSize := unsafe.Sizeof(DtvEntry{}) * (l.MaxTlsIndex + 2)
	pthreadSize := unsafe.Sizeof(Pthread{})
	totalSize := l.StaticTlsSize + uint64(tcbSize)

	addr := mem.AllocReadWriteMemory(uintptr(totalSize))
	tcb := (*Tcb)(unsafe.Pointer(addr + uintptr(l.StaticTlsSize)))
	dtv := (*DtvEntry)(unsafe.Pointer(mem.AllocReadWriteMemory(dtvSize)))
	pthreadAddr := mem.AllocReadWriteMemory(pthreadSize)
	pthread := (*Pthread)(unsafe.Pointer(pthreadAddr))

	tcb.Self = tcb
	tcb.Dtv = dtv
	tcb.Thread = pthreadAddr
	tcb.Fiber = 0

	dtvSlice := unsafe.Slice(dtv, l.MaxTlsIndex+2)
	dtvSlice[0].Counter = l.GenerationCounter
	dtvSlice[1].Counter = l.MaxTlsIndex

	pthread.Magic = PthreadMagic
	pthread.ThreadId = 1337
	pthread.Flags = 0
	pthread.ReturnValue = 0
	pthread.Error = 0
	pthread.CleanupHandlerStack = 0
	copy(pthread.Name[:], "MainThread")

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

// GetCurrentThread returns pointer to the Pthread struct of the current thread.
func GetCurrentThread() uintptr {
	tcbAddr, _, _ := sys_struct.TlsGetValue.Call(sys_struct.TlsSlot)
	if tcbAddr == 0 {
		return 0
	}

	tcb := (*Tcb)(unsafe.Pointer(tcbAddr))
	return tcb.Thread
}
