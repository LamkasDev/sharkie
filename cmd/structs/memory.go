package structs

var GlobalAllocator = NewAllocator()

const (
	SCE_KERNEL_ERROR_EINVAL  = 0x80020016
	SCE_KERNEL_ERROR_ENOMEM  = 0x8002000C
	SCE_KERNEL_ERROR_EACCESS = 0x8002000D

	// Alignment requirements (16KB).
	MEMORY_ALIGN_MASK = 0x3FFF
	MEMORY_ALIGN      = 0x4000
)

const (
	SCE_KERNEL_MTYPE_C_SHARED = 0xC // Onion (CPU/GPU shared)
	SCE_KERNEL_MTYPE_C        = 0x3 // Garlic (GPU optimized)
)

var MemoryTypeNames = map[uintptr]string{
	SCE_KERNEL_MTYPE_C_SHARED: "SCE_KERNEL_MTYPE_C_SHARED",
	SCE_KERNEL_MTYPE_C:        "SCE_KERNEL_MTYPE_C",
}

const (
	PROT_NONE  = 0x0
	PROT_READ  = 0x1
	PROT_WRITE = 0x2
	PROT_EXEC  = 0x4
)

const (
	MAP_FAILED  = ERR_PTR
	MAP_PRIVATE = 0x2
	MAP_FIXED   = 0x10
	MAP_ANON    = 0x1000
)

type Allocator struct {
	Allocations         map[uintptr]uintptr
	DirectMemoryBase    uintptr
	DirectMemoryCurrent uintptr
}

// NewAllocator creates a new instance of Allocator.
func NewAllocator() *Allocator {
	return &Allocator{
		DirectMemoryBase:    0xFE0000000,
		DirectMemoryCurrent: 0xFE0000000,
		Allocations:         map[uintptr]uintptr{},
	}
}
