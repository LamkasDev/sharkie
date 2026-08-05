package posix

import (
	"fmt"
	"sync"

	"github.com/goki/vulkan"
)

// GlobalAllocator should be used for explicit allocations (mmap, etc.)
var GlobalAllocator *Allocator

// GlobalGpuAllocator should be used for GPU-memory allocations.
var GlobalGpuAllocator *Allocator

const SystemPageShift = 12
const SystemPageSize = uintptr(1 << SystemPageShift)

var (
	HookMap            func(addr uintptr, length uint64, prot int32)
	HookMapDirect      func(addr uintptr, length uint64, offset uint64, memType int32, prot int32)
	HookUnmap          func(addr uintptr, length uintptr)
	HookProtect        func(addr uintptr, length uintptr, prot int32)
	HookAllocateDirect func(offset uintptr, length uint64, memType int32)

	HookAllocateMemoryVulkan func(offset uintptr, length uint64)
	HookFreeMemoryVulkan     func(offset uintptr)
	HookMapMemoryVulkan      func(addr uintptr, length uint64, offset uintptr)
)

const (
	SCE_KERNEL_MTYPE_WB_ONION  = 0x0 // Onion Bus (CPU shared)
	SCE_KERNEL_MTYPE_WC_GARLIC = 0x3 // Garlic Bus (CPU/GPU optimized)
	SCE_KERNEL_MTYPE_WB_GARLIC = 0xA // Garlic Bus (GPU optimized)
)

var MemoryTypeNames = map[int32]string{
	SCE_KERNEL_MTYPE_WB_ONION:  "SCE_KERNEL_MTYPE_WB_ONION",
	SCE_KERNEL_MTYPE_WC_GARLIC: "SCE_KERNEL_MTYPE_WC_GARLIC",
	SCE_KERNEL_MTYPE_WB_GARLIC: "SCE_KERNEL_MTYPE_WB_GARLIC",
}

const (
	PROT_NONE      = 0x0
	PROT_READ      = 0x1
	PROT_WRITE     = 0x2
	PROT_EXEC      = 0x4
	PROT_GPU_READ  = 0x10
	PROT_GPU_WRITE = 0x20
)

const (
	MAP_PRIVATE = 0x2
	MAP_FIXED   = 0x10
	MAP_ANON    = 0x1000
	MAP_SYSTEM  = 0x2000
)

const (
	DirectMemoryDefaultSize = uint64(0x100000000) // 4GB
	GpuMemoryDefaultSize    = uint64(0x080000000) // 2GB
	MemoryPageSize          = uint64(0x4000)      // 16KB
	GuardPageSize           = uint64(4096)        // 4KB
)

type VirtualQueryInfo struct {
	Start      uint64
	End        uint64
	Offset     uint64
	Protection int32
	MemoryType int32
	Bitfield   uint8 // is_flexible:1, is_direct:1, is_stack:1, is_pooled:1, is_committed:1
	_          [3]uint8
	_          [4]uint8
	Name       [32]byte
}

type Allocator struct {
	Base          uintptr
	Current       uintptr
	Size          uint64
	Buffer        vulkan.Buffer
	DeviceAddress uint64
	Lock          sync.Mutex
}

func SetupAllocator() {
	GlobalAllocator = NewAllocator(0x400000000, DirectMemoryDefaultSize)
	GlobalGpuAllocator = NewAllocator(0xFE0000000, GpuMemoryDefaultSize)
}

// NewAllocator creates a new instance of Allocator.
func NewAllocator(base uintptr, size uint64) *Allocator {
	return &Allocator{
		Base:    base,
		Current: base,
		Size:    size,
		Lock:    sync.Mutex{},
	}
}

func (allocator *Allocator) GetNextAlignedAddress(alignment, length uint64) uintptr {
	allocator.Lock.Lock()
	defer allocator.Lock.Unlock()
	if alignment == 0 {
		alignment = MemoryPageSize
	}

	alignedLength := (length + (alignment - 1)) &^ (alignment - 1)
	addr := (allocator.Current + uintptr(alignment-1)) &^ uintptr(alignment-1)
	allocator.Current = addr + uintptr(alignedLength)

	return addr
}

func MemoryProtName(prot int32) string {
	name := ""
	if (prot&PROT_READ) != 0 || (prot&PROT_GPU_READ) != 0 {
		name = fmt.Sprintf("%sR", name)
	}
	if (prot&PROT_WRITE) != 0 || (prot&PROT_GPU_WRITE) != 0 {
		name = fmt.Sprintf("%sW", name)
	}
	if (prot & PROT_EXEC) != 0 {
		name = fmt.Sprintf("%sE", name)
	}
	if name == "" {
		name = "NO_ACCESS"
	}

	return name
}
