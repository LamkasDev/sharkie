package translation

import (
	"github.com/LamkasDev/sharkie/cmd/structs"
)

// ReadMemory services a CPU read fault on GPU-tracked guest memory.
func (t *GpuTranslator) ReadMemory(address, size uintptr) bool {
	if !t.IsGpuMapped(address, size) {
		return false
	}
	if !structs.GlobalMemoryManager.IsRegionGpuModified(address, size) {
		return true
	}

	downloaded := map[uintptr]struct{}{}
	structs.GlobalMemoryManager.ForEachDownloadRange(address, size, func(rangeAddr, rangeSize uintptr) {
		for _, image := range t.CollectGpuResourcesInRange(rangeAddr, rangeSize) {
			if _, ok := downloaded[image.Address]; ok {
				continue
			}
			downloaded[image.Address] = struct{}{}
			_ = image.DownloadFromVkImage(&t.handles)
			recordSyncDownload()
		}
	})

	return true
}

// InvalidateMemory services a CPU write fault or guest mprotect.
func (t *GpuTranslator) InvalidateMemory(address, size uintptr) bool {
	if !t.IsGpuMapped(address, size) {
		return false
	}

	structs.GlobalMemoryManager.InvalidateRegion(address, size, func() {
		t.ReadMemory(address, size)
	})

	t.imagesMutex.Lock()
	t.invalidateImageMemory(address, size)
	t.imagesMutex.Unlock()

	return true
}

// MapMemory records a GPU-visible guest mapping.
func (t *GpuTranslator) MapMemory(address, size uintptr) {
	// Handled directly by MemoryManager Guest tracker
}

// UnmapMemory evicts GPU caches for a released guest mapping (shadPS4 Rasterizer::UnmapMemory).
func (t *GpuTranslator) UnmapMemory(address, size uintptr) {
	t.imagesMutex.Lock()
	t.unmapImageMemory(address, size)
	t.imagesMutex.Unlock()
}

// IsGpuMapped reports whether the address range may contain GPU-tracked guest memory.
func (t *GpuTranslator) IsGpuMapped(address, size uintptr) bool {
	if structs.GlobalMemoryManager.IsRegionTracked(address, size) {
		return true
	}
	if structs.GlobalMemoryManager.Guest().IsGpuVisible(address, size) {
		return true
	}

	return false
}
