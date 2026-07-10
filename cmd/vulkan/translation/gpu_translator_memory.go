package translation

import (
	"github.com/LamkasDev/sharkie/cmd/lib_structs"
	"github.com/LamkasDev/sharkie/cmd/structs"
)

// ReadMemory services a CPU read fault on GPU-tracked guest memory.
func (t *GpuTranslator) ReadMemory(address, size uintptr) bool {
	if !t.IsGpuMapped(address, size) {
		return false
	}

	downloaded := map[uintptr]struct{}{}
	for _, image := range t.CollectGpuResourcesInRange(address, size) {
		if !image.ShouldDownloadFromVkImage() {
			continue
		}
		if _, ok := downloaded[image.Address]; ok {
			continue
		}
		downloaded[image.Address] = struct{}{}
		_ = image.DownloadFromVkImage(&t.handles)
	}

	return true
}

// InvalidateMemory services a CPU write fault or guest mprotect.
func (t *GpuTranslator) InvalidateMemory(address, size uintptr) bool {
	if !t.IsGpuMapped(address, size) {
		return false
	}
	t.ReadMemory(address, size)

	images := t.CollectGpuResourcesInRange(address, size)
	for _, image := range images {
		image.MarkCpuModified()
	}

	return true
}

// IsGpuMapped reports whether the address range may contain GPU-tracked guest memory.
func (t *GpuTranslator) IsGpuMapped(address, size uintptr) bool {
	end := address + size
	for addr := address >> lib_structs.SystemPageShift; (addr << lib_structs.SystemPageShift) < end; addr++ {
		if page, ok := structs.GlobalMemoryManager.Pages[addr]; ok && page.Mapped {
			return true
		}
	}
	return false
}
