package posix

import (
	"sync/atomic"
)

var NextAddress uintptr = 0x3100000000

func GetNextAlignedAddress(addr uintptr, length uint64, alignment uintptr) uintptr {
	if addr == 0 {
		for {
			// Align the current bump pointer.
			current := atomic.LoadUintptr(&NextAddress)
			alignedAddr := (current + alignment - 1) & ^(alignment - 1)
			alignedSize := (length + uint64(MemoryPageSize) - 1) & ^uint64(MemoryPageSize-1)
			next := alignedAddr + uintptr(alignedSize)
			if atomic.CompareAndSwapUintptr(&NextAddress, current, next) {
				addr = alignedAddr
				break
			}
		}
	} else {
		addr = addr & ^(alignment - 1)
	}

	return addr
}
