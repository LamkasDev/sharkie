package structs

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
	"golang.org/x/sys/unix"
)

type TrackedPage struct {
	Address     uintptr
	Size        uintptr
	IsDirty     bool
	Format      *PageFormat
	BaseAddress uintptr
}

type PageFormat struct {
	DataFormat  uint8
	NumFormat   uint8
	TilingIndex uint8
	Pitch       uint32
	Height      uint32
	IsSurface   bool
}

type MemoryManager struct {
	TrackedPages map[uintptr]*TrackedPage
	PageFormats  map[uintptr]*PageFormat
	Lock         sync.RWMutex
}

var GlobalMemoryManager = &MemoryManager{
	TrackedPages: make(map[uintptr]*TrackedPage),
	PageFormats:  make(map[uintptr]*PageFormat),
}

const SystemPageSize = uintptr(4096)

// RegisterFormat records the format and tiling metadata for a specific base address.
func (m *MemoryManager) RegisterFormat(baseAddress uintptr, format PageFormat) {
	if baseAddress == 0 {
		return
	}

	m.Lock.Lock()
	defer m.Lock.Unlock()

	alignedAddr := baseAddress & ^(SystemPageSize - 1)
	m.PageFormats[alignedAddr] = &format
}

// TrackRegion marks a memory region as tracked (PROT_NONE) to intercept CPU accesses.
func (m *MemoryManager) TrackRegion(address uintptr, size uintptr) error {
	if address == 0 {
		return nil
	}

	alignedAddr := address & ^(SystemPageSize - 1)
	alignedSize := (size + (address - alignedAddr) + SystemPageSize - 1) & ^(SystemPageSize - 1)

	b := unsafe.Slice((*byte)(unsafe.Pointer(alignedAddr)), alignedSize)

	err := unix.Mprotect(b, unix.PROT_NONE)
	if err != nil {
		return fmt.Errorf("failed to mprotect region: %v", err)
	}

	m.Lock.Lock()
	defer m.Lock.Unlock()

	for p := alignedAddr; p < alignedAddr+alignedSize; p += SystemPageSize {
		m.TrackedPages[p] = &TrackedPage{
			Address:     p,
			Size:        SystemPageSize,
			IsDirty:     false,
			BaseAddress: address,
		}
		cTrackPage(p)
	}

	return nil
}

// UntrackRegion unprotects and stops tracking a memory region.
func (m *MemoryManager) UntrackRegion(address uintptr, size uintptr) error {
	if address == 0 {
		return nil
	}

	alignedAddr := address & ^(SystemPageSize - 1)
	alignedSize := (size + (address - alignedAddr) + SystemPageSize - 1) & ^(SystemPageSize - 1)

	b := unsafe.Slice((*byte)(unsafe.Pointer(alignedAddr)), alignedSize)

	err := unix.Mprotect(b, unix.PROT_READ|unix.PROT_WRITE)
	if err != nil {
		return fmt.Errorf("failed to unprotect region: %v", err)
	}

	m.Lock.Lock()
	defer m.Lock.Unlock()

	for p := alignedAddr; p < alignedAddr+alignedSize; p += SystemPageSize {
		delete(m.TrackedPages, p)
		cUntrackPage(p)
	}

	return nil
}

// HandleFault is called by the SIGSEGV handler. Returns true if the fault was successfully handled.
func (m *MemoryManager) HandleFault(faultAddress uintptr) bool {
	alignedAddr := faultAddress & ^(SystemPageSize - 1)

	m.Lock.Lock()
	page, ok := m.TrackedPages[alignedAddr]
	if !ok {
		m.Lock.Unlock()
		return false
	}
	m.Lock.Unlock() // Unlock before potentially slow operations

	// For step 1, just log the fault and unprotect it to allow execution to continue.
	logger.Printf("[MemoryManager] Trapped CPU access to tracked GPU memory at address: %s (Page: %s)\n",
		color.Yellow.Sprintf("0x%X", faultAddress),
		color.Yellow.Sprintf("0x%X", alignedAddr))

	b := unsafe.Slice((*byte)(unsafe.Pointer(alignedAddr)), SystemPageSize)
	err := unix.Mprotect(b, unix.PROT_READ|unix.PROT_WRITE)
	if err != nil {
		logger.Printf("[MemoryManager] FATAL: failed to unprotect page: %v\n", err)
		return false
	}

	m.Lock.Lock()
	page.IsDirty = true
	m.Lock.Unlock()

	return true
}
