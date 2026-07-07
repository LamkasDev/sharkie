package structs

import (
	"sync"

	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
)

// MemoryRasterizer handles guest CPU faults and map/unmap notifications.
type MemoryRasterizer interface {
	ReadMemory(addr uintptr, size uintptr) bool
	InvalidateMemory(addr uintptr, size uintptr) bool
	IsGpuMapped(addr uintptr, size uintptr) bool
	MapMemory(addr uintptr, size uintptr)
	UnmapMemory(addr uintptr, size uintptr)
}

type MemoryManager struct {
	pageManager *PageManager
	cpuModified map[uintptr]struct{}
	gpuModified map[uintptr]struct{}
	guest       *GuestMemory
	rasterizer  MemoryRasterizer
	Lock        sync.RWMutex
}

var GlobalMemoryManager = &MemoryManager{
	pageManager: newPageManager(),
	cpuModified: make(map[uintptr]struct{}),
	gpuModified: make(map[uintptr]struct{}),
	guest:       newGuestMemory(),
}

func init() {
	HookRegisterDirectAllocation = GlobalMemoryManager.Guest().RegisterDirectAllocation
	HookMapAnonymous = GlobalMemoryManager.Guest().MapAnonymous
	HookMapDirect = GlobalMemoryManager.Guest().MapDirect
	HookOnMapGuest = GlobalMemoryManager.OnMapGuest
	HookOnUnmapGuest = GlobalMemoryManager.OnUnmapGuest
	HookOnProtectGuest = GlobalMemoryManager.OnProtectGuest
}

func alignPageRange(address, size uintptr) (start, end uintptr) {
	start = address & ^(uintptr(0x1000) - 1)
	end = address + size
	end = (end + uintptr(0x1000) - 1) & ^(uintptr(0x1000) - 1)

	return start, end
}

// SetRasterizer wires the Vulkan translator for fault-time sync (called from NewGpuTranslator).
func (m *MemoryManager) SetRasterizer(r MemoryRasterizer) {
	m.rasterizer = r
}

// Guest returns the guest virtual memory region table.
func (m *MemoryManager) Guest() *GuestMemory {
	return m.guest
}

// OnMapGuest notifies the rasterizer that guest memory became GPU-visible.
func (m *MemoryManager) OnMapGuest(address uintptr, size uintptr) {
	if !m.guest.IsGpuVisible(address, size) {
		return
	}
	m.rasterizer.MapMemory(address, size)
}

func (m *MemoryManager) setPageBits(bits map[uintptr]struct{}, address, size uintptr, enabled bool) {
	start, end := alignPageRange(address, size)
	for page := start; page < end; page += SystemPageSize {
		if enabled {
			bits[page] = struct{}{}
		} else {
			delete(bits, page)
		}
	}
}

func (m *MemoryManager) anyPageIn(bits map[uintptr]struct{}, address, size uintptr) bool {
	start, end := alignPageRange(address, size)
	for page := start; page < end; page += SystemPageSize {
		if _, ok := bits[page]; ok {
			return true
		}
	}

	return false
}

// IsRegionCpuModified reports whether any page in the range was CPU-written.
func (m *MemoryManager) IsRegionCpuModified(address, size uintptr) bool {
	m.Lock.Lock()
	defer m.Lock.Unlock()
	return m.anyPageIn(m.cpuModified, address, size)
}

// IsRegionGpuModified reports whether any page in the range was GPU-written.
func (m *MemoryManager) IsRegionGpuModified(address, size uintptr) bool {
	m.Lock.Lock()
	defer m.Lock.Unlock()
	return m.anyPageIn(m.gpuModified, address, size)
}

// MarkRegionCpuModified marks CPU ownership and removes write traps.
func (m *MemoryManager) MarkRegionCpuModified(address, size uintptr) {
	m.Lock.Lock()
	defer m.Lock.Unlock()
	start, end := alignPageRange(address, size)
	for page := start; page < end; page += SystemPageSize {
		m.cpuModified[page] = struct{}{}
		delete(m.gpuModified, page)
	}
	m.pageManager.UpdatePageWatchers(address, size, false, false)
	m.pageManager.UpdatePageWatchers(address, size, false, true)
}

// MarkRegionGpuModified marks GPU ownership and installs read traps.
func (m *MemoryManager) MarkRegionGpuModified(address, size uintptr) {
	m.Lock.Lock()
	defer m.Lock.Unlock()
	start, end := alignPageRange(address, size)
	for page := start; page < end; page += SystemPageSize {
		delete(m.cpuModified, page)
		m.gpuModified[page] = struct{}{}
	}
	m.pageManager.UpdatePageWatchers(address, size, true, true)
}

// UnmarkRegionCpuModified clears CPU-dirty state and restores write traps.
func (m *MemoryManager) UnmarkRegionCpuModified(address, size uintptr) {
	m.Lock.Lock()
	defer m.Lock.Unlock()
	start, end := alignPageRange(address, size)
	for page := start; page < end; page += SystemPageSize {
		delete(m.cpuModified, page)
	}
	m.pageManager.UpdatePageWatchers(address, size, false, false)
}

// UnmarkRegionGpuModified clears GPU-dirty state and removes read traps.
func (m *MemoryManager) UnmarkRegionGpuModified(address, size uintptr) {
	m.Lock.Lock()
	defer m.Lock.Unlock()
	start, end := alignPageRange(address, size)
	for page := start; page < end; page += SystemPageSize {
		delete(m.gpuModified, page)
	}
	m.pageManager.UpdatePageWatchers(address, size, false, true)
}

// TrackRegion installs write watchers for a registered guest region.
func (m *MemoryManager) TrackRegion(address, size uintptr) {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	start, end := alignPageRange(address, size)
	var batchStart uintptr
	var batchSize uintptr
	flushBatch := func() {
		if batchSize == 0 {
			return
		}
		m.pageManager.UpdatePageWatchers(batchStart, batchSize, true, false)
		batchStart = 0
		batchSize = 0
	}
	for page := start; page < end; page += SystemPageSize {
		if _, cpu := m.cpuModified[page]; cpu {
			flushBatch()
			continue
		}
		if batchSize == 0 {
			batchStart = page
			batchSize = SystemPageSize
			continue
		}
		batchSize += SystemPageSize
	}
	flushBatch()
}

// UntrackRegion removes all watchers for a guest region.
func (m *MemoryManager) UntrackRegion(address, size uintptr) {
	m.pageManager.UpdatePageWatchers(address, size, false, false)
	m.pageManager.UpdatePageWatchers(address, size, false, true)
}

// InvalidateRegion handles CPU write faults: mark CPU-dirty and optionally flush GPU data.
func (m *MemoryManager) InvalidateRegion(address, size uintptr, onFlush func()) {
	m.Lock.Lock()
	shouldFlush := m.anyPageIn(m.gpuModified, address, size)
	m.Lock.Unlock()

	if shouldFlush && onFlush != nil {
		onFlush()
	}

	m.MarkRegionCpuModified(address, size)
}

// ForEachDownloadRange invokes fn for each contiguous GPU-modified subrange and clears GPU bits.
func (m *MemoryManager) ForEachDownloadRange(address, size uintptr, fn func(rangeAddress, rangeSize uintptr)) {
	m.Lock.Lock()
	defer m.Lock.Unlock()

	start, end := alignPageRange(address, size)
	var rangeStart uintptr
	var inRange bool

	flushRange := func(rangeEnd uintptr) {
		if !inRange {
			return
		}
		rangeSize := rangeEnd - rangeStart
		fn(rangeStart, rangeSize)
		m.setPageBits(m.gpuModified, rangeStart, rangeSize, false)
		m.pageManager.UpdatePageWatchers(rangeStart, rangeSize, false, true)
		inRange = false
	}

	for page := start; page < end; page += SystemPageSize {
		if _, ok := m.gpuModified[page]; ok {
			if !inRange {
				rangeStart = page
				inRange = true
			}
			continue
		}
		if inRange {
			flushRange(page)
		}
	}
	if inRange {
		flushRange(end)
	}
}

func (m *MemoryManager) clearRegion(address, size uintptr) {
	m.Lock.Lock()
	start, end := alignPageRange(address, size)
	for page := start; page < end; page += SystemPageSize {
		delete(m.cpuModified, page)
		delete(m.gpuModified, page)
	}
	m.Lock.Unlock()
	m.pageManager.UpdatePageWatchers(address, size, false, false)
	m.pageManager.UpdatePageWatchers(address, size, false, true)
}

// OnUnmapGuest notifies the rasterizer before guest memory is released.
func (m *MemoryManager) OnUnmapGuest(address uintptr, size uintptr) {
	m.rasterizer.UnmapMemory(address, size)
	m.clearRegion(address, size)
	m.guest.Unmap(address, uint64(size))
}

// IsRegionTracked reports whether any page in the range has active watchers.
func (m *MemoryManager) IsRegionTracked(addr, size uintptr) bool {
	m.pageManager.mu.Lock()
	defer m.pageManager.mu.Unlock()
	start, end := alignPageRange(addr, size)
	for page := start; page < end; page += SystemPageSize {
		if st, ok := m.pageManager.pages[page]; ok && (st.readWatchers > 0 || st.writeWatchers > 0) {
			return true
		}
	}

	return false
}

// OnProtectGuest is called after sceKernelMprotect succeeds.
func (m *MemoryManager) OnProtectGuest(address uintptr, size uintptr, prot int32) {
	m.guest.Protect(address, uint64(size), prot)
	if m.rasterizer.IsGpuMapped(address, size) && (prot&(PROT_WRITE|PROT_GPU_WRITE)) != 0 {
		m.rasterizer.InvalidateMemory(address, size)
	}
}

// HandleReadFault services a CPU read on a GPU-protected page.
func (m *MemoryManager) HandleReadFault(faultAddress uintptr) bool {
	return m.rasterizer.ReadMemory(faultAddress, SystemPageSize)
}

// HandleWriteFault services a CPU write on a write-protected page.
func (m *MemoryManager) HandleWriteFault(faultAddress uintptr) bool {
	return m.rasterizer.InvalidateMemory(faultAddress, SystemPageSize)
}
