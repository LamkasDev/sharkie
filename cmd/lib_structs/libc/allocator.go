package libc

import (
	"sync"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/langhuihui/gomem"
)

// GlobalGoAllocator should be used for implicit allocations (inside init stubs, etc.)
var GlobalGoAllocator *GoAllocator

// Default alignment is 32 bytes.
// https://lists.llvm.org/pipermail/cfe-commits/Week-of-Mon-20180305/221024.html.
const (
	AllocationAlignment  = 32
	AllocationHeaderSize = 16
)

type GoAllocator struct {
	Allocator   *gomem.ScalableMemoryAllocator
	Allocations map[uintptr][]byte
	Lock        sync.Mutex
}

// NewGoAllocator creates a new instance of GoAllocator.
func NewGoAllocator() *GoAllocator {
	goAllocator := &GoAllocator{
		Allocations: map[uintptr][]byte{},
		Lock:        sync.Mutex{},
	}
	goAllocator.Allocator = gomem.NewScalableMemoryAllocator(1<<30, 0)
	goMmapHint := uintptr(0x200000000)
	goMmapHintMu := sync.Mutex{}
	goAllocator.Allocator.CustomAllocator = func(size int, hint uintptr) []byte {
		if posix.HookAllocateLibcVulkan != nil {
			if hint == 0 {
				goMmapHintMu.Lock()
				hint = goMmapHint
				goMmapHint = (goMmapHint + uintptr(size) + 0x1FFFFF) &^ 0x1FFFFF
				goMmapHintMu.Unlock()
			}
			return posix.HookAllocateLibcVulkan(size, hint)
		}
		return nil
	}

	return goAllocator
}

// NewGoAllocatorFromBuffer creates a new instance of GoAllocator using an already mapped buffer, and makes it non-expandable.
func NewGoAllocatorFromBuffer(base uintptr, capacity uintptr) *GoAllocator {
	goAllocator := &GoAllocator{
		Allocations: map[uintptr][]byte{},
		Lock:        sync.Mutex{},
	}
	goAllocator.Allocator = gomem.NewScalableMemoryAllocator(int(capacity), base)
	goAllocator.Allocator.Expandable = false
	goAllocator.Allocator.AddPreallocated(unsafe.Slice((*byte)(unsafe.Pointer(base)), int(capacity)))

	return goAllocator
}

func (allocator *GoAllocator) Malloc(size uintptr) uintptr {
	return allocator.MallocAligned(size, AllocationAlignment)
}

func (allocator *GoAllocator) MallocAligned(size, alignment uintptr) uintptr {
	if size == 0 {
		size = 1
	}
	allocator.Lock.Lock()
	defer allocator.Lock.Unlock()

	// We need 16-bytes for header and 15-bytes for worst case alignment.
	allocatedSize := size + AllocationHeaderSize + (alignment - 1)
	dataSlice := allocator.Allocator.Malloc(int(allocatedSize))
	if len(dataSlice) == 0 {
		return 0
	}
	address := uintptr(unsafe.Pointer(&dataSlice[0]))
	alignedAddress := (address + AllocationHeaderSize + (alignment - 1)) & ^uintptr(alignment-1)
	headerAddress := alignedAddress - AllocationHeaderSize
	allocator.Allocations[address] = dataSlice

	// Write header (0 - original pointer, 8 - allocated size).
	*(*uintptr)(unsafe.Pointer(headerAddress)) = address
	*(*uintptr)(unsafe.Pointer(headerAddress + 8)) = uintptr(allocatedSize)

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
	dataSlice, exists := allocator.Allocations[address]
	if !exists {
		return false
		panic("double free or unallocated address")
	}
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

func (allocator *GoAllocator) UsableSize(ptr uintptr) uintptr {
	if ptr == 0 {
		return 0
	}
	allocator.Lock.Lock()
	defer allocator.Lock.Unlock()

	// Read header (0 - original pointer, 8 - allocated size).
	headerAddr := ptr - AllocationHeaderSize
	address := *(*uintptr)(unsafe.Pointer(headerAddr))
	if _, exists := allocator.Allocations[address]; !exists {
		return 0
	}
	allocatedSize := *(*uintptr)(unsafe.Pointer(headerAddr + 8))

	padding := ptr - address
	return allocatedSize - padding
}

func SetupGoAllocator() {
	GlobalGoAllocator = NewGoAllocator()
}
