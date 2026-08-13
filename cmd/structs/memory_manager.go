package structs

import (
	"fmt"
	"slices"
	"sync"
	"syscall"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
)

type MemoryPage struct {
	Mapped    bool
	Reserved  bool
	Prot      uint32
	Resources []interface{}
}

type MemoryManager struct {
	Pages      map[uintptr]*MemoryPage
	VMAs       []VMA
	DirectVMAs []VMA
	Lock       sync.Mutex
}

var GlobalMemoryManager = &MemoryManager{
	Pages: make(map[uintptr]*MemoryPage),
	VMAs: []VMA{
		{Start: 0, End: ^uintptr(0), Mapped: false, Reserved: false, Prot: 0},
	},
	DirectVMAs: []VMA{
		{Start: 0, End: ^uintptr(0), Mapped: false, Reserved: false, Prot: 0},
	},
	Lock: sync.Mutex{},
}

func init() {
	posix.HookMap = func(addr uintptr, length uint64, prot int32) {
		GlobalMemoryManager.Map(addr, uintptr(length))
		GlobalMemoryManager.Protect(addr, uintptr(length), uint32(prot))
	}
	posix.HookMapDirect = func(addr uintptr, length uint64, offset uint64, memType int32, prot int32) {
		GlobalMemoryManager.MapDirect(addr, uintptr(length), offset, memType, uint32(prot))
	}
	posix.HookReserve = func(addr uintptr, length uint64) {
		GlobalMemoryManager.Reserve(addr, uintptr(length))
	}
	posix.HookUnmap = func(addr uintptr, length uintptr) {
		GlobalMemoryManager.Unmap(addr, length)
	}
	posix.HookProtect = func(addr uintptr, length uintptr, prot int32) {
		GlobalMemoryManager.Protect(addr, length, uint32(prot))
	}
	posix.HookAllocateDirect = func(offset uintptr, length uint64, memType int32) {
		GlobalMemoryManager.AllocateDirect(offset, uintptr(length), memType)
	}
	posix.HookName = func(addr uintptr, length uint64, name string) {
		GlobalMemoryManager.Name(addr, uintptr(length), name)
	}
}

func (m *MemoryManager) getPage(addr uintptr) *MemoryPage {
	pageAddr := addr >> posix.SystemPageShift
	if p, ok := m.Pages[pageAddr]; ok {
		return p
	}
	p := &MemoryPage{
		Resources: []interface{}{},
	}
	m.Pages[pageAddr] = p
	return p
}

func (m *MemoryManager) IsAddressRangeFree(address, size uintptr) bool {
	end := address + size
	m.Lock.Lock()
	defer m.Lock.Unlock()
	for addr := address >> posix.SystemPageShift; (addr << posix.SystemPageShift) < end; addr++ {
		if page, ok := m.Pages[addr]; ok && (page.Mapped || page.Reserved) {
			return false
		}
	}

	return true
}

func (m *MemoryManager) IsAddressRangeUnmapped(address, size uintptr) bool {
	end := address + size
	m.Lock.Lock()
	defer m.Lock.Unlock()
	for addr := address >> posix.SystemPageShift; (addr << posix.SystemPageShift) < end; addr++ {
		if page, ok := m.Pages[addr]; ok && page.Mapped {
			return false
		}
	}

	return true
}

func (m *MemoryManager) Reserve(address, size uintptr) {
	end := address + size
	m.Lock.Lock()
	for addr := address >> posix.SystemPageShift; (addr << posix.SystemPageShift) < end; addr++ {
		page := m.getPage(addr << posix.SystemPageShift)
		page.Reserved = true
	}
	m.updateVMA(address, end, func(v *VMA) {
		v.Reserved = true
	})
	m.Lock.Unlock()
}

func (m *MemoryManager) Name(address, size uintptr, name string) {
	end := address + size
	m.Lock.Lock()
	m.updateVMA(address, end, func(v *VMA) {
		v.Name = name
	})
	m.Lock.Unlock()
}

func (m *MemoryManager) Map(address, size uintptr) {
	end := address + size
	m.Lock.Lock()
	for addr := address >> posix.SystemPageShift; (addr << posix.SystemPageShift) < end; addr++ {
		page := m.getPage(addr << posix.SystemPageShift)
		page.Mapped = true
		page.Reserved = false
	}
	m.updateVMA(address, end, func(v *VMA) {
		v.Mapped = true
		v.Reserved = false
	})
	m.Lock.Unlock()
}

func (m *MemoryManager) MapDirect(address, size uintptr, offset uint64, memType int32, prot uint32) {
	end := address + size
	m.Lock.Lock()
	for addr := address >> posix.SystemPageShift; (addr << posix.SystemPageShift) < end; addr++ {
		page := m.getPage(addr << posix.SystemPageShift)
		page.Mapped = true
		page.Reserved = false
		page.Prot = prot
	}
	m.updateVMA(address, end, func(v *VMA) {
		v.Mapped = true
		v.Reserved = false
		v.IsDirect = true
		v.Offset = offset
		v.MemoryType = memType
		v.Prot = prot
	})
	m.Lock.Unlock()
}

func (m *MemoryManager) Unmap(address, size uintptr) {
	end := address + size
	m.Lock.Lock()
	for addr := address >> posix.SystemPageShift; (addr << posix.SystemPageShift) < end; addr++ {
		page := m.getPage(addr << posix.SystemPageShift)
		page.Mapped = false
		page.Reserved = false
	}
	m.updateVMA(address, end, func(v *VMA) {
		v.Mapped = false
		v.Reserved = false
		v.Prot = 0
		v.Name = ""
		v.MemoryType = 0
		v.IsDirect = false
	})
	m.Lock.Unlock()
}

func (m *MemoryManager) Protect(address, size uintptr, prot uint32) {
	end := address + size
	m.Lock.Lock()
	for addr := address >> posix.SystemPageShift; (addr << posix.SystemPageShift) < end; addr++ {
		page := m.getPage(addr << posix.SystemPageShift)
		page.Prot = prot
	}
	m.updateVMA(address, end, func(v *VMA) {
		v.Prot = prot
	})
	m.Lock.Unlock()
}

func (m *MemoryManager) UpdateTraps(address, size uintptr, protState int) {
	end := address + size
	for addr := address >> posix.SystemPageShift; (addr << posix.SystemPageShift) < end; addr++ {
		pageAddr := addr << posix.SystemPageShift
		cSetProtState(pageAddr, protState)
	}

	// Update system page protection
	alignedAddress := address & ^(uintptr(posix.SystemPageSize) - 1)
	alignedSize := (size + posix.SystemPageSize - 1) & ^(posix.SystemPageSize - 1)

	var sysProt int
	switch protState {
	case 0: // PROT_NONE
		sysProt = posix.PROT_NONE // 0
	case 1: // PROT_READ
		sysProt = posix.PROT_READ // 1
	case 2: // PROT_READ | PROT_WRITE
		sysProt = posix.PROT_READ | posix.PROT_WRITE // 3
	}

	mprotectSlice := unsafe.Slice((*byte)(unsafe.Pointer(alignedAddress)), alignedSize)
	err := syscall.Mprotect(mprotectSlice, sysProt)
	if err != nil {
		fmt.Printf("failed mprotect on %X (size %X): %v\n", alignedAddress, alignedSize, err)
	}
}

func (m *MemoryManager) Track(address, size uintptr, resource interface{}) {
	end := address + size
	m.Lock.Lock()
	for addr := address >> posix.SystemPageShift; (addr << posix.SystemPageShift) < end; addr++ {
		pageAddr := addr << posix.SystemPageShift
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
	for addr := address >> posix.SystemPageShift; (addr << posix.SystemPageShift) < end; addr++ {
		pageAddr := addr << posix.SystemPageShift
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
