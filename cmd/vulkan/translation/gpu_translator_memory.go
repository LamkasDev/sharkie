package translation

import (
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/LamkasDev/sharkie/cmd/vulkan"
)

// ReadMemory services a CPU read fault on GPU-tracked guest memory.
func (t *GpuTranslator) ReadMemory(address, size uintptr) bool {
	err := vulkan.RunWithCommandBuffer(t.handles.DownloadQueue, t.handles, func(commandBuffer *vulkan.VulkanCommandBuffer) {
		err := t.DownloadRegionVkImages(address, size, commandBuffer)
		if err != nil {
			panic(err)
		}
	}, t.currentGuestFrame)
	if err != nil {
		panic(err)
	}

	return true
}

// InvalidateMemory services a CPU write fault or guest mprotect.
func (t *GpuTranslator) InvalidateMemory(address, size uintptr) bool {
	for _, image := range t.CollectGpuResourcesInRange(address, size) {
		image.MarkCpuModified(t.currentGuestFrame)
	}

	return true
}

// IsGpuMapped reports whether the address range may contain GPU-tracked guest memory.
func (t *GpuTranslator) IsGpuMapped(address, size uintptr) bool {
	end := address + size

	structs.GlobalMemoryManager.Lock.Lock()
	defer structs.GlobalMemoryManager.Lock.Unlock()
	for addr := address >> SystemPageShift; (addr << SystemPageShift) < end; addr++ {
		if page, ok := structs.GlobalMemoryManager.Pages[addr]; ok && page.Mapped {
			return true
		}
	}

	return false
}
