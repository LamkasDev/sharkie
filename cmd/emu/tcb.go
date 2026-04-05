package emu

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/linker"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs/pthread"
	. "github.com/LamkasDev/sharkie/cmd/structs/tcb"
	"github.com/gookit/color"
)

// NewTcb creates a new instance of Tcb for passed thread.
func NewTcb(thread *Thread) *Tcb {
	GlobalModuleManager.ModulesLock.RLock()
	defer GlobalModuleManager.ModulesLock.RUnlock()

	maxTlsIndex := uintptr(len(GlobalModuleManager.ModulesMap))
	tlsSize := uintptr(linker.GlobalLinker.StaticTlsSize)
	tcbOffset := (tlsSize + TcbAlignment - 1) &^ (TcbAlignment - 1)
	totalSize := tcbOffset + TcbSize

	addr := GlobalGoAllocator.Malloc(totalSize)
	tcbAddr := addr + tcbOffset
	tcb := (*Tcb)(unsafe.Pointer(tcbAddr))
	tcb.Self = tcb

	dtvAddr := GlobalGoAllocator.Malloc(DtvEntrySize * (maxTlsIndex + 2))
	tcb.Dtv = (*DtvEntry)(unsafe.Pointer(dtvAddr))
	threadAddr := GlobalGoAllocator.Malloc(PthreadSize)
	tcb.Thread = (*Pthread)(unsafe.Pointer(threadAddr))
	tcb.Fiber = 0

	dtvSlice := unsafe.Slice(tcb.Dtv, maxTlsIndex+2)
	dtvSlice[0].Counter = linker.GlobalLinker.GenerationCounter
	dtvSlice[1].Counter = maxTlsIndex

	tcb.Thread.Self = threadAddr
	tcb.Thread.TcbSelf = tcbAddr
	tcb.Thread.StartFunc = 0
	tcb.Thread.Arg = 0
	tcb.Thread.Attr = PthreadAttr{}
	tcb.Thread.ReturnValue = 0
	tcb.Thread.NamePtr = Cstring(GlobalGoAllocator.Malloc(33))
	CString(tcb.Thread.NamePtr, thread.Name)
	tcb.Thread.CleanupStack = 0
	tcb.Thread.Magic = PthreadMagic

	for _, module := range GlobalModuleManager.Modules {
		if module == nil || module.TlsSection == nil || module.TlsSection.ImageSize == 0 {
			continue
		}
		dest := tcbAddr - uintptr(module.TlsSection.Offset)
		if module.TlsSection.InitImageSize > 0 {
			src := uintptr(unsafe.Pointer(&module.Memory[0])) + uintptr(module.TlsSection.ImageVirtualAddress)
			copy(
				unsafe.Slice((*byte)(unsafe.Pointer(dest)), module.TlsSection.InitImageSize),
				unsafe.Slice((*byte)(unsafe.Pointer(src)), module.TlsSection.InitImageSize),
			)
		}
		dtvSlice[module.ModuleIndex+1].Pointer = dest

		logger.Printf(
			"[%s] Copied %s bytes of %s's PT_TLS data from %s to %s (image size %s).\n",
			color.Green.Sprint(thread.Name),
			color.Green.Sprintf("%d", module.TlsSection.InitImageSize),
			color.Blue.Sprint(module.Name),
			color.Yellow.Sprintf("0x%X", module.TlsSection.ImageVirtualAddress),
			color.Yellow.Sprintf("0x%X", dest),
			color.Gray.Sprintf("%d", module.TlsSection.ImageSize),
		)
	}

	return tcb
}
