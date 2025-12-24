package structs

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/gookit/color"
	"github.com/langhuihui/gomem"
)

// GlobalAllocator should be used for explicit allocations (mmap, etc.)
var GlobalAllocator = NewAllocator()

// GlobalGoAllocator should be used for implicit allocations (inside init stubs, etc.)
var GlobalGoAllocator = NewGoAllocator()

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
	SCE_KERNEL_MTYPE_WB_ONION = 0x0 // Onion Bus (CPU/GPU shared)
	SCE_KERNEL_MTYPE_C_SHARED = 0xC // Onion (CPU optimized)
	SCE_KERNEL_MTYPE_C        = 0x3 // Garlic (GPU optimized)
)

var MemoryTypeNames = map[uintptr]string{
	SCE_KERNEL_MTYPE_WB_ONION: "SCE_KERNEL_MTYPE_WB_ONION",
	SCE_KERNEL_MTYPE_C_SHARED: "SCE_KERNEL_MTYPE_C_SHARED",
	SCE_KERNEL_MTYPE_C:        "SCE_KERNEL_MTYPE_C",
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
	MemoryPageSize          = uintptr(0x4000)      // 16KB
	GuardPageSize           = uintptr(4096)        // 4KB
)

type Allocator struct {
	Allocations         map[uintptr]uintptr
	DirectMemoryBase    uintptr
	DirectMemoryCurrent uintptr
	DirectMemorySize    uintptr
	Lock                sync.Mutex
}

type GoAllocator struct {
	Allocator *gomem.ScalableMemoryAllocator
}

// NewAllocator creates a new instance of Allocator.
func NewAllocator() *Allocator {
	var err error
	allocator := &Allocator{
		DirectMemorySize: DirectMemoryDefaultSize,
		Allocations:      map[uintptr]uintptr{},
		Lock:             sync.Mutex{},
	}
	allocator.DirectMemoryBase, err = ReserveKernelMemory(0x400000000, allocator.DirectMemorySize)
	if allocator.DirectMemoryBase == 0 {
		panic(err)
	}
	allocator.DirectMemoryCurrent = allocator.DirectMemoryBase
	fmt.Printf(
		"Reserved %s bytes for the global allocator at %s.\n",
		color.Yellow.Sprintf("0x%X", allocator.DirectMemorySize),
		color.Yellow.Sprintf("0x%X", allocator.DirectMemoryBase),
	)

	return allocator
}

// NewGoAllocator creates a new instance of GoAllocator.
func NewGoAllocator() *GoAllocator {
	goAllocator := &GoAllocator{
		Allocator: gomem.NewScalableMemoryAllocator(1024),
	}

	return goAllocator
}

func (allocator *GoAllocator) Malloc(size uintptr) uintptr {
	data := allocator.Allocator.Malloc(int(size))
	return (uintptr)(unsafe.Pointer(&data[0]))
}

func (allocator *GoAllocator) Free(address, size uintptr) bool {
	data := unsafe.Slice((*byte)(unsafe.Pointer(address)), size)
	return allocator.Allocator.Free(data)
}
