package structs

import (
	"encoding/binary"
	"fmt"
	"sync"
	"unsafe"

	"github.com/goki/vulkan"
	"github.com/langhuihui/gomem"
)

// GlobalAllocator should be used for explicit allocations (mmap, etc.)
var GlobalAllocator *Allocator

// GlobalGpuAllocator should be used for GPU-memory allocations.
var GlobalGpuAllocator *Allocator

// GlobalGoAllocator should be used for implicit allocations (inside init stubs, etc.)
var GlobalGoAllocator *GoAllocator

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

const (
	AllocationAlignment  = 16
	AllocationHeaderSize = 16
)

type Allocator struct {
	Base    uintptr
	Current uintptr
	Size    uint64
	Alloc   func(size uint64) (vulkan.Buffer, uintptr, error)
	Map     func(addr uintptr, length uint64, handle uintptr) error
	Ranges  []AllocatorMemoryRange
	Lock    sync.Mutex
}

type AllocatorMemoryRange struct {
	Base   uintptr
	Size   uint64
	Buffer vulkan.Buffer
}

type GoAllocator struct {
	Allocator   *gomem.ScalableMemoryAllocator
	Allocations map[uintptr][]byte
	Lock        sync.Mutex
}

func SetupAllocator() {
	GlobalAllocator = NewAllocator(0x400000000, DirectMemoryDefaultSize)
	GlobalGpuAllocator = NewAllocator(0xFE0000000, GpuMemoryDefaultSize)
	GlobalGoAllocator = NewGoAllocator()
}

// NewAllocator creates a new instance of Allocator.
func NewAllocator(base uintptr, size uint64) *Allocator {
	return &Allocator{
		Base:    base,
		Current: base,
		Size:    size,
		Ranges:  []AllocatorMemoryRange{},
		Lock:    sync.Mutex{},
	}
}

func (allocator *Allocator) FindRange(addr uintptr) *AllocatorMemoryRange {
	allocator.Lock.Lock()
	defer allocator.Lock.Unlock()

	for i := range allocator.Ranges {
		r := &allocator.Ranges[i]
		if addr >= r.Base && addr < r.Base+uintptr(r.Size) {
			return r
		}
	}

	return nil
}

// NewGoAllocator creates a new instance of GoAllocator.
func NewGoAllocator() *GoAllocator {
	goAllocator := &GoAllocator{
		Allocations: map[uintptr][]byte{},
		Lock:        sync.Mutex{},
	}
	goAllocator.Allocator = gomem.NewScalableMemoryAllocator(1025)

	return goAllocator
}

func (allocator *GoAllocator) Malloc(size uintptr) uintptr {
	if size == 0 {
		size = 1
	}
	allocator.Lock.Lock()
	defer allocator.Lock.Unlock()

	// We need 16-bytes for header and 15-bytes for worst case alignment.
	allocatedSize := size + AllocationHeaderSize + (AllocationAlignment - 1)
	dataSlice := allocator.Allocator.Malloc(int(allocatedSize))
	if len(dataSlice) == 0 {
		return 0
	}
	address := uintptr(unsafe.Pointer(&dataSlice[0]))
	alignedAddress := (address + AllocationHeaderSize + (AllocationAlignment - 1)) & ^uintptr(AllocationAlignment-1)
	headerAddress := alignedAddress - AllocationHeaderSize
	allocator.Allocations[address] = dataSlice

	// Write header (0 - original pointer, 8 - allocated size).
	headerSlice := unsafe.Slice((*byte)(unsafe.Pointer(headerAddress)), AllocationHeaderSize)
	binary.LittleEndian.PutUint64(headerSlice, uint64(address))
	binary.LittleEndian.PutUint64(headerSlice[8:], uint64(allocatedSize))

	return alignedAddress
}

func (allocator *GoAllocator) Free(ptr uintptr) bool {
	if ptr == 0 {
		return true
	}
	allocator.Lock.Lock()
	defer allocator.Lock.Unlock()

	// Read header (0 - original pointer, 8 - allocated size).
	headerAddr := ptr - AllocationHeaderSize
	address := *(*uintptr)(unsafe.Pointer(headerAddr))
	allocatedSize := *(*uintptr)(unsafe.Pointer(headerAddr + 8))
	dataSlice := unsafe.Slice((*byte)(unsafe.Pointer(address)), allocatedSize)
	delete(allocator.Allocations, address)

	return allocator.Allocator.Free(dataSlice)
}

func (allocator *GoAllocator) Realloc(ptr uintptr, newSize uintptr) uintptr {
	if ptr == 0 {
		return allocator.Malloc(newSize)
	}
	if newSize == 0 {
		allocator.Free(ptr)
		return 0
	}

	// Read header (0 - original pointer, 8 - allocated size).
	headerAddr := ptr - AllocationHeaderSize
	address := *(*uintptr)(unsafe.Pointer(headerAddr))
	allocatedSize := *(*uintptr)(unsafe.Pointer(headerAddr + 8))

	// Allocate new block.
	padding := ptr - address
	oldUserSize := allocatedSize - padding
	newAddress := allocator.Malloc(newSize)
	if newAddress == 0 {
		return 0
	}

	// Copy contents.
	copySize := oldUserSize
	if newSize < copySize {
		copySize = newSize
	}
	copy(
		unsafe.Slice((*byte)(unsafe.Pointer(newAddress)), copySize),
		unsafe.Slice((*byte)(unsafe.Pointer(ptr)), copySize),
	)
	allocator.Free(ptr)

	return newAddress
}

func (allocator *Allocator) GetNextAlignedAddress(alignment, length uint64) uintptr {
	allocator.Lock.Lock()
	defer allocator.Lock.Unlock()

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
