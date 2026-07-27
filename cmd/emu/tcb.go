package emu

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/elf"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/libc"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/pthread"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/tcb"
	"github.com/LamkasDev/sharkie/cmd/linker"
	"github.com/LamkasDev/sharkie/cmd/logger"
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

	tcb.Dtv, _ = allocateDtv(maxTlsIndex+2, nil)
	dtvSlice := unsafe.Slice(tcb.Dtv, maxTlsIndex+2)

	threadAddr := GlobalGoAllocator.Malloc(PthreadSize)
	tcb.Thread = (*Pthread)(unsafe.Pointer(threadAddr))
	tcb.Fiber = 0

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
		copyTlsData(module, dest)
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

// ExpandThreadTLS dynamically allocates TLS memory for a newly loaded module across all existing threads.
func ExpandThreadTLS(module *elf.Elf) {
	if module == nil || module.TlsSection == nil || module.TlsSection.ImageSize == 0 {
		return
	}

	GlobalModuleManager.ModulesLock.RLock()
	maxTlsIndex := uintptr(len(GlobalModuleManager.ModulesMap))
	GlobalModuleManager.ModulesLock.RUnlock()

	ThreadLock.Lock()
	defer ThreadLock.Unlock()

	for _, thread := range ThreadRepo {
		if thread.Tcb == nil || thread.Tcb.Dtv == nil {
			continue
		}
		oldDtvHeader := unsafe.Slice(thread.Tcb.Dtv, 2)
		oldCounter := oldDtvHeader[1].Counter
		oldDtvSlice := unsafe.Slice(thread.Tcb.Dtv, oldCounter+2)
		if maxTlsIndex <= oldCounter {
			// DTV is already large enough, just allocate the block if it's not already allocated.
			if uintptr(module.ModuleIndex)+1 < uintptr(len(oldDtvSlice)) && oldDtvSlice[module.ModuleIndex+1].Pointer == 0 {
				oldDtvSlice[module.ModuleIndex+1].Pointer = allocateTlsBlock(module)
			}
			continue
		}

		// Allocate new DTV.
		newDtv, newDtvSlice := allocateDtv(maxTlsIndex+2, oldDtvSlice)
		thread.Tcb.Dtv = newDtv

		// Allocate TLS block for the new module.
		newDtvSlice[module.ModuleIndex+1].Pointer = allocateTlsBlock(module)
	}
}

// allocateDtv allocates a new DTV array and copies old entries if provided.
func allocateDtv(size uintptr, oldDtv []DtvEntry) (*DtvEntry, []DtvEntry) {
	newDtvAddr := GlobalGoAllocator.Malloc(DtvEntrySize * size)
	newDtv := (*DtvEntry)(unsafe.Pointer(newDtvAddr))
	newDtvSlice := unsafe.Slice(newDtv, size)
	if oldDtv != nil {
		copy(newDtvSlice[2:], oldDtv[2:])
	}
	newDtvSlice[0].Counter = linker.GlobalLinker.GenerationCounter
	newDtvSlice[1].Counter = size - 2
	return newDtv, newDtvSlice
}

// allocateTlsBlock allocates and initializes a new TLS block for a module.
func allocateTlsBlock(module *elf.Elf) uintptr {
	dest := GlobalGoAllocator.Malloc(uintptr(module.TlsSection.ImageSize))
	copyTlsData(module, dest)
	return dest
}

// copyTlsData copies initialized TLS data and zero-fills the BSS section.
func copyTlsData(module *elf.Elf, dest uintptr) {
	if module.TlsSection.InitImageSize > 0 {
		src := uintptr(unsafe.Pointer(&module.Memory[0])) + uintptr(module.TlsSection.ImageVirtualAddress)
		copy(
			unsafe.Slice((*byte)(unsafe.Pointer(dest)), module.TlsSection.InitImageSize),
			unsafe.Slice((*byte)(unsafe.Pointer(src)), module.TlsSection.InitImageSize),
		)
	}
	if module.TlsSection.ImageSize > module.TlsSection.InitImageSize {
		bssSize := module.TlsSection.ImageSize - module.TlsSection.InitImageSize
		bssDest := dest + uintptr(module.TlsSection.InitImageSize)
		for i := uint64(0); i < bssSize; i++ {
			*(*byte)(unsafe.Pointer(bssDest + uintptr(i))) = 0
		}
	}
}
