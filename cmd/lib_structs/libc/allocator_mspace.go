package libc

import (
	"fmt"
	"sync"
)

// GlobalMspaceAllocator tracks created mspaces.
var GlobalMspaceAllocator *MspaceAllocator

// MspaceAllocator holds handles and lock to created mspaces.
type MspaceAllocator struct {
	Mspaces map[uintptr]*MspaceInfo
	Lock    sync.Mutex
}

// MspaceInfo holds info about a mspace.
type MspaceInfo struct {
	Name    string
	Base    uintptr
	End     uintptr
	Current uintptr
	Mutex   sync.Mutex
}

// NewMspaceAllocator creates a new instance of MspaceAllocator.
func NewMspaceAllocator() *MspaceAllocator {
	return &MspaceAllocator{
		Mspaces: map[uintptr]*MspaceInfo{},
		Lock:    sync.Mutex{},
	}
}

// Alloc bump-allocates size bytes with given alignment from ms. Returns 0 if out of space.
func (ms *MspaceInfo) Alloc(alignment, size uintptr) (uintptr, error) {
	if alignment < 1 {
		alignment = 1
	}
	ms.Mutex.Lock()
	defer ms.Mutex.Unlock()

	alignedAddress := (ms.Current + alignment - 1) &^ (alignment - 1)
	if alignedAddress+size > ms.End {
		return 0, fmt.Errorf("lack of space")
	}
	ms.Current = alignedAddress + size

	return alignedAddress, nil
}

func SetupMspaceAllocator() {
	GlobalMspaceAllocator = NewMspaceAllocator()
}
