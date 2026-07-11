package structs

import (
	"slices"
	"sync"
	"syscall"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/lib_structs"
)

type MemoryPage struct {
	Mapped    bool
	Prot      uint32
	Resources []interface{}
}

type MemoryManager struct {
	Pages map[uintptr]*MemoryPage
	Lock  sync.Mutex
}

var GlobalMemoryManager = &MemoryManager{
	Pages: make(map[uintptr]*MemoryPage),
	Lock:  sync.Mutex{},
}

func init() {
	lib_structs.HookMap = func(addr uintptr, length uint64, prot int32) {
		GlobalMemoryManager.Map(addr, uintptr(length))
		GlobalMemoryManager.Protect(addr, uintptr(length), uint32(prot))
	}
	lib_structs.HookUnmap = func(addr uintptr, length uintptr) {
		GlobalMemoryManager.Unmap(addr, length)
	}
	lib_structs.HookProtect = func(addr uintptr, length uintptr, prot int32) {
		GlobalMemoryManager.Protect(addr, length, uint32(prot))
	}
}

func (m *MemoryManager) getPage(addr uintptr) *MemoryPage {
	pageAddr := addr >> lib_structs.SystemPageShift
	if p, ok := m.Pages[pageAddr]; ok {
		return p
	}
	p := &MemoryPage{
		Resources: []interface{}{},
	}
	m.Pages[pageAddr] = p
	return p
}

func (m *MemoryManager) Map(address, size uintptr) {
	end := address + size
	m.Lock.Lock()
	for addr := address >> lib_structs.SystemPageShift; (addr << lib_structs.SystemPageShift) < end; addr++ {
		page := m.getPage(addr << lib_structs.SystemPageShift)
		page.Mapped = true
	}
	m.Lock.Unlock()
}

func (m *MemoryManager) Unmap(address, size uintptr) {
	end := address + size
	m.Lock.Lock()
	for addr := address >> lib_structs.SystemPageShift; (addr << lib_structs.SystemPageShift) < end; addr++ {
		page := m.getPage(addr << lib_structs.SystemPageShift)
		page.Mapped = false
	}
	m.Lock.Unlock()
}

func (m *MemoryManager) Protect(address, size uintptr, prot uint32) {
	end := address + size
	m.Lock.Lock()
	for addr := address >> lib_structs.SystemPageShift; (addr << lib_structs.SystemPageShift) < end; addr++ {
		page := m.getPage(addr << lib_structs.SystemPageShift)
		page.Prot = prot
	}
	m.Lock.Unlock()
}

func (m *MemoryManager) UpdateTraps(address, size uintptr, protState int) {
	end := address + size
	for addr := address >> lib_structs.SystemPageShift; (addr << lib_structs.SystemPageShift) < end; addr++ {
		pageAddr := addr << lib_structs.SystemPageShift
		cSetProtState(pageAddr, protState)
	}

	// Update system page protection
	pageMask := uintptr(lib_structs.SystemPageSize - 1)
	alignedAddress := address &^ pageMask
	alignedSize := (size + (address - alignedAddress) + pageMask) &^ pageMask

	var sysProt int
	switch protState {
	case 0: // PROT_NONE
		sysProt = lib_structs.PROT_NONE // 0
	case 1: // PROT_READ
		sysProt = lib_structs.PROT_READ // 1
	case 2: // PROT_READ | PROT_WRITE
		sysProt = lib_structs.PROT_READ | lib_structs.PROT_WRITE // 3
	}

	mprotectSlice := unsafe.Slice((*byte)(unsafe.Pointer(alignedAddress)), alignedSize)
	syscall.Mprotect(mprotectSlice, sysProt)
}

func (m *MemoryManager) Track(address, size uintptr, resource interface{}) {
	end := address + size
	m.Lock.Lock()
	for addr := address >> lib_structs.SystemPageShift; (addr << lib_structs.SystemPageShift) < end; addr++ {
		pageAddr := addr << lib_structs.SystemPageShift
		page := m.getPage(pageAddr)
		page.Resources = append(page.Resources, resource)
		cTrackPage(pageAddr, 0) // 0 = PROT_NONE, traps reads and writes
	}
	m.Lock.Unlock()
	m.UpdateTraps(address, size, 0)
}

func (m *MemoryManager) Untrack(address, size uintptr, resource interface{}) {
	end := address + size
	m.Lock.Lock()
	for addr := address >> lib_structs.SystemPageShift; (addr << lib_structs.SystemPageShift) < end; addr++ {
		pageAddr := addr << lib_structs.SystemPageShift
		page := m.getPage(pageAddr)
		if i := slices.Index(page.Resources, resource); i >= 0 {
			page.Resources = slices.Delete(page.Resources, i, i+1)
		}
		if len(page.Resources) == 0 {
			cUntrackPage(pageAddr)
		}
	}
	m.Lock.Unlock()
}
