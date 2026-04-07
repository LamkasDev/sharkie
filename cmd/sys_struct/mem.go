package sys_struct

import (
	"os"
	"sync/atomic"
)

var NextAddress uintptr = 0x3100000000

// GrowGoStack grows the current goroutine stack by kb kilobytes, essentially pre-allocating space.
func GrowGoStack(kb int) {
	var dummy [1024]byte
	if kb > 0 {
		GrowGoStack(kb - 1)
	}
	_ = dummy
}

func GetNextAlignedAddress(addr uintptr, length uint64) uintptr {
	pageSize := uint64(os.Getpagesize())
	alignedLength := (length + (pageSize - 1)) &^ (pageSize - 1)
	if addr == 0 {
		addr = (atomic.LoadUintptr(&NextAddress) + uintptr(pageSize-1)) &^ uintptr(pageSize-1)
		atomic.StoreUintptr(&NextAddress, addr+uintptr(alignedLength))
	} else {
		addr = addr &^ uintptr(pageSize-1)
	}

	return addr
}
