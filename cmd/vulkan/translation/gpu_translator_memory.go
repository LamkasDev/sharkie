package translation

import (
	"github.com/LamkasDev/sharkie/cmd/lib_structs"
	"github.com/LamkasDev/sharkie/cmd/structs"
)

// ReadMemory services a CPU read fault on GPU-tracked guest memory.
func (t *GpuTranslator) ReadMemory(address, size uintptr) bool {
	err := t.DownloadRegionVkImages(address, size)
	if err != nil {
		panic(err)
	}

	return true
}

// InvalidateMemory services a CPU write fault or guest mprotect.
func (t *GpuTranslator) InvalidateMemory(address, size uintptr) bool {
	t.ReadMemory(address, size)
	for _, image := range t.CollectGpuResourcesInRange(address, size) {
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
