package libc

import (
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
	Name      string
	Base      uintptr
	End       uintptr
	Allocator *GoAllocator
	Mutex     sync.Mutex
}

// NewMspaceAllocator creates a new instance of MspaceAllocator.
func NewMspaceAllocator() *MspaceAllocator {
	return &MspaceAllocator{
		Mspaces: map[uintptr]*MspaceInfo{},
		Lock:    sync.Mutex{},
	}
}

func SetupMspaceAllocator() {
	GlobalMspaceAllocator = NewMspaceAllocator()
}
