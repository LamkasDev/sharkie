package libc

import (
	"sync"
	"unsafe"

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
	goAllocator.Allocator = gomem.NewScalableMemoryAllocator(1025)

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

func SetupGoAllocator() {
	GlobalGoAllocator = NewGoAllocator()
}
