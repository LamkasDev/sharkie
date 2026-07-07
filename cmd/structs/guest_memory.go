package structs

import (
	"sync"

	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
)

// GuestVMA tracks one contiguous guest virtual mapping.
type GuestVMA struct {
	Base       uintptr
	Size       uintptr
	Type       VMAType
	Prot       int32
	MemType    int32
	PhysOffset uintptr
	Mapped     bool
}

func (v GuestVMA) End() uintptr {
	return v.Base + v.Size
}

func (v GuestVMA) Contains(address, size uintptr) bool {
	if address < v.Base {
		return false
	}
	end := address + size

	return end <= v.End()
}

func (v GuestVMA) Overlaps(address, size uintptr) bool {
	end := address + size
	return address < v.End() && end > v.Base
}

func hasGpuProt(prot int32) bool {
	return (prot & (PROT_GPU_READ | PROT_GPU_WRITE)) != 0
}

// GuestMemory tracks guest virtual mappings for kernel hooks and GPU visibility queries.
type GuestMemory struct {
	regions map[uintptr]GuestVMA
	mu      sync.RWMutex
}

func newGuestMemory() *GuestMemory {
	return &GuestMemory{regions: make(map[uintptr]GuestVMA)}
}

func alignGuestSize(length uint64) uintptr {
	pageSize := uintptr(MemoryPageSize)
	return (uintptr(length) + pageSize - 1) &^ (pageSize - 1)
}

// RegisterDirectAllocation records physical direct memory carved by AllocateDirectMemory.
func (g *GuestMemory) RegisterDirectAllocation(base uintptr, size uint64, memType int32) {
	var physOffset uintptr
	if memType == SCE_KERNEL_MTYPE_WC_GARLIC || memType == SCE_KERNEL_MTYPE_WB_GARLIC {
		physOffset = base - GlobalGpuAllocator.Base
	} else {
		physOffset = base - GlobalAllocator.Base
	}

	g.mu.Lock()
	defer g.mu.Unlock()
	g.regions[base] = GuestVMA{
		Base:       base,
		Size:       uintptr(alignGuestSize(size)),
		Type:       VMATypeDirect,
		MemType:    memType,
		PhysOffset: physOffset,
	}
}

// MapDirect marks a direct-memory range mapped with the given protection.
func (g *GuestMemory) MapDirect(base uintptr, size uint64, prot int32) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if vma, ok := g.regions[base]; ok {
		vma.Prot = prot
		vma.Mapped = true
		g.regions[base] = vma
		return
	}

	memType := int32(SCE_KERNEL_MTYPE_WB_ONION)
	physOffset := base - GlobalAllocator.Base
	if base >= GlobalGpuAllocator.Base {
		memType = SCE_KERNEL_MTYPE_WC_GARLIC
		physOffset = base - GlobalGpuAllocator.Base
	}

	g.regions[base] = GuestVMA{
		Base:       base,
		Size:       uintptr(alignGuestSize(size)),
		Type:       VMATypeDirect,
		Prot:       prot,
		MemType:    memType,
		PhysOffset: physOffset,
		Mapped:     true,
	}
}

// MapAnonymous records a flexible or mmap-backed guest region.
func (g *GuestMemory) MapAnonymous(base uintptr, size uint64, prot int32, vmaType VMAType) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.regions[base] = GuestVMA{
		Base:   base,
		Size:   uintptr(alignGuestSize(size)),
		Type:   vmaType,
		Prot:   prot,
		Mapped: true,
	}
}

// Unmap removes guest mappings overlapping [addr, addr+size).
func (g *GuestMemory) Unmap(address uintptr, size uint64) {
	size = uint64(alignGuestSize(size))
	end := address + uintptr(size)

	g.mu.Lock()
	defer g.mu.Unlock()
	for base, vma := range g.regions {
		vmaEnd := vma.End()
		if address >= vmaEnd || end <= base {
			continue
		}
		if address <= base && end >= vmaEnd {
			delete(g.regions, base)
			continue
		}

		// Partial unmap: shrink or split is rare for our callers; drop the whole VMA if partially touched.
		delete(g.regions, base)
	}
}

// Protect updates protection on a mapped range.
func (g *GuestMemory) Protect(address uintptr, size uint64, prot int32) {
	size = uint64(alignGuestSize(size))

	g.mu.Lock()
	defer g.mu.Unlock()
	for base, vma := range g.regions {
		if vma.Overlaps(address, uintptr(size)) {
			vma.Prot = prot
			g.regions[base] = vma
		}
	}
}

// Lookup returns the VMA containing addr, if any.
func (g *GuestMemory) Lookup(address uintptr) (GuestVMA, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, vma := range g.regions {
		if address >= vma.Base && address < vma.End() {
			return vma, true
		}
	}

	return GuestVMA{}, false
}

// ForEachOverlap calls fn for VMAs intersecting [addr, addr+size).
func (g *GuestMemory) ForEachOverlap(address, size uintptr, fn func(GuestVMA)) {
	end := address + size

	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, vma := range g.regions {
		if vma.Overlaps(address, size) && address < vma.End() && vma.Base < end {
			fn(vma)
		}
	}
}

// IsGpuVisible reports whether any overlapping mapped VMA is GPU-accessible.
func (g *GuestMemory) IsGpuVisible(address, size uintptr) bool {
	visible := false
	g.ForEachOverlap(address, size, func(vma GuestVMA) {
		if vma.Mapped && hasGpuProt(vma.Prot) {
			visible = true
		}
	})

	return visible
}

// RegisterPreMapped records the onion/garlic Vulkan-backed regions at emulator init.
func (g *GuestMemory) RegisterPreMapped() {
	g.mu.Lock()
	defer g.mu.Unlock()

	gpuProt := int32(PROT_READ | PROT_WRITE | PROT_GPU_READ | PROT_GPU_WRITE)
	g.regions[GlobalAllocator.Base] = GuestVMA{
		Base:       GlobalAllocator.Base,
		Size:       uintptr(GlobalAllocator.Size),
		Type:       VMATypeDirect,
		Prot:       gpuProt,
		MemType:    SCE_KERNEL_MTYPE_WB_ONION,
		PhysOffset: 0,
		Mapped:     true,
	}
	g.regions[GlobalGpuAllocator.Base] = GuestVMA{
		Base:       GlobalGpuAllocator.Base,
		Size:       uintptr(GlobalGpuAllocator.Size),
		Type:       VMATypeDirect,
		Prot:       gpuProt,
		MemType:    SCE_KERNEL_MTYPE_WC_GARLIC,
		PhysOffset: 0,
		Mapped:     true,
	}
}
