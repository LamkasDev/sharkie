package structs

import (
	"encoding/binary"
	"fmt"
	"sync"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
	"github.com/langhuihui/gomem"
)

// GlobalAllocator should be used for explicit allocations (mmap, etc.)
var GlobalAllocator *Allocator

// GlobalGoAllocator should be used for implicit allocations (inside init stubs, etc.)
var GlobalGoAllocator *GoAllocator

const (
	SCE_KERNEL_ERROR_ENOTSUP      = 0x80020001
	SCE_KERNEL_ERROR_ENOENT       = 0x80020002
	SCE_KERNEL_ERROR_EINVAL       = 0x80020016
	SCE_KERNEL_ERROR_ENOMEM       = 0x8002000C
	SCE_KERNEL_ERROR_EACCESS      = 0x8002000D
	SCE_KERNEL_ERROR_ENAMETOOLONG = 0x80020042

	// Alignment requirements (16KB).
	MEMORY_ALIGN_MASK = 0x3FFF
	MEMORY_ALIGN      = 0x4000
)

const (
	SCE_KERNEL_MTYPE_WB_ONION  = 0x0 // Onion Bus (CPU shared)
	SCE_KERNEL_MTYPE_WC_GARLIC = 0x3 // Garlic Bus (CPU/GPU optimized)
	SCE_KERNEL_MTYPE_WB_GARLIC = 0xA // Garlic Bus (GPU optimized)
)

var MemoryTypeNames = map[uintptr]string{
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
	DirectMemoryDefaultSize = uintptr(0x100000000) // 4GB
	GpuMemoryDefaultSize    = uintptr(0x080000000) // 2GB
	MemoryPageSize          = uintptr(0x4000)      // 16KB
	GuardPageSize           = uintptr(4096)        // 4KB
)

const (
	AllocationAlignment  = 16
	AllocationHeaderSize = 16
)

type Allocator struct {
	DirectMemoryBase    uintptr
	DirectMemoryCurrent uintptr
	DirectMemorySize    uintptr
	GpuMemoryBase       uintptr
	GpuMemoryCurrent    uintptr
	GpuMemorySize       uintptr
}

type GoAllocator struct {
	Allocator *gomem.ScalableMemoryAllocator
	Lock      sync.Mutex
}

func SetupAllocator() {
	GlobalAllocator = NewAllocator()
	GlobalGoAllocator = NewGoAllocator()
}

// NewAllocator creates a new instance of Allocator.
func NewAllocator() *Allocator {
	var err error
	allocator := &Allocator{
		DirectMemorySize: DirectMemoryDefaultSize,
		GpuMemorySize:    GpuMemoryDefaultSize,
	}
	allocator.DirectMemoryBase, err = ReserveKernelMemory(0x400000000, allocator.DirectMemorySize)
	if allocator.DirectMemoryBase == 0 {
		panic(err)
	}
	allocator.DirectMemoryCurrent = allocator.DirectMemoryBase
	allocator.GpuMemoryBase, err = ReserveKernelMemory(0xFE0000000, allocator.GpuMemorySize)
	if allocator.GpuMemoryBase == 0 {
		panic(err)
	}
	allocator.GpuMemoryCurrent = allocator.GpuMemoryBase
	logger.Printf(
		"Reserved %s of direct memory (%s) and %s bytes of graphics memory (%s).\n",
		color.Yellow.Sprintf("0x%X", allocator.DirectMemorySize),
		color.Yellow.Sprintf("0x%X", allocator.DirectMemoryBase),
		color.Yellow.Sprintf("0x%X", allocator.GpuMemorySize),
		color.Yellow.Sprintf("0x%X", allocator.GpuMemoryBase),
	)

	return allocator
}

// NewGoAllocator creates a new instance of GoAllocator.
func NewGoAllocator() *GoAllocator {
	goAllocator := &GoAllocator{
		Allocator: gomem.NewScalableMemoryAllocator(1024),
		Lock:      sync.Mutex{},
	}

	return goAllocator
}

func (allocator *GoAllocator) Malloc(size uintptr) uintptr {
	if size == 0 {
		size = 1
	}

	// We need 16-bytes for header and 15-bytes for worst case alignment.
	allocatedSize := size + AllocationHeaderSize + (AllocationAlignment - 1)
	allocator.Lock.Lock()
	dataSlice := allocator.Allocator.Malloc(int(allocatedSize))
	allocator.Lock.Unlock()
	if len(dataSlice) == 0 {
		return 0
	}
	address := uintptr(unsafe.Pointer(&dataSlice[0]))
	alignedAddress := (address + AllocationHeaderSize + (AllocationAlignment - 1)) & ^uintptr(AllocationAlignment-1)
	headerAddress := alignedAddress - AllocationHeaderSize

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

	// Read header (0 - original pointer, 8 - allocated size).
	headerAddr := ptr - AllocationHeaderSize
	address := *(*uintptr)(unsafe.Pointer(headerAddr))
	allocatedSize := *(*uintptr)(unsafe.Pointer(headerAddr + 8))
	dataSlice := unsafe.Slice((*byte)(unsafe.Pointer(address)), allocatedSize)

	allocator.Lock.Lock()
	defer allocator.Lock.Unlock()
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

func MemoryProtName(prot uintptr) string {
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
