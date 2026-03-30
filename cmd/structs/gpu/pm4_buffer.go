package gpu

import (
	"fmt"
	"unsafe"
)

const GNM_MAX_CB_SIZE_DWORDS = uint32(0x3FFFFD)

type PM4IndirectBuffer struct {
	Header      uint32
	AddressLow  uint32
	AddressHigh uint32
	SizeDW      uint32
}

const PM4IndirectBufferSize = unsafe.Sizeof(PM4IndirectBuffer{})

func NewPM4IndirectBuffer(gpuAddr uintptr, sizeBytes uint32, isCCB bool) PM4IndirectBuffer {
	opcode := uint32(PM4_IT_INDIRECT_BUFFER)
	if isCCB {
		opcode = PM4_IT_INDIRECT_BUFFER_CNST
	}

	return PM4IndirectBuffer{
		Header:      NewPM4TypedHeader(opcode, 3),
		AddressLow:  uint32(gpuAddr),
		AddressHigh: uint32(gpuAddr >> 32),
		SizeDW:      (sizeBytes >> 2) & 0x000FFFFF,
	}
}

// BuildPM4IndirectBuffers packs DCB/CCB pairs into PM4 indirect buffers.
// CCBs are prepended before their paired DCBs.
func BuildPM4IndirectBuffers(count, dcbGpuAddrsPtr, dcbSizesPtr, ccbGpuAddrsPtr, ccbSizesPtr uintptr) ([]PM4IndirectBuffer, error) {
	dcbAddresses := unsafe.Slice((*uintptr)(unsafe.Pointer(dcbGpuAddrsPtr)), count)
	dcbSizes := unsafe.Slice((*uint32)(unsafe.Pointer(dcbSizesPtr)), count)

	// CCBs are optional.
	hasCcbs := ccbGpuAddrsPtr != 0 && ccbSizesPtr != 0
	var ccbAddrs []uintptr
	var ccbSizes []uint32
	if hasCcbs {
		ccbAddrs = unsafe.Slice((*uintptr)(unsafe.Pointer(ccbGpuAddrsPtr)), count)
		ccbSizes = unsafe.Slice((*uint32)(unsafe.Pointer(ccbSizesPtr)), count)
	}

	buffers := make([]PM4IndirectBuffer, 0, int(count)*2)
	for i := range count {
		dcbSize := dcbSizes[i]
		if dcbSize == 0 {
			return nil, fmt.Errorf("DCB %d has zero size", i)
		}
		if dcbSize>>2 > GNM_MAX_CB_SIZE_DWORDS {
			return nil, fmt.Errorf("DCB %d size 0x%X exceeds limit", i, dcbSize)
		}

		// CCBs are optional, prepend them before DCBs.
		if hasCcbs && ccbSizes[i] != 0 {
			if ccbSizes[i]>>2 > GNM_MAX_CB_SIZE_DWORDS {
				return nil, fmt.Errorf("CCB %d size 0x%X exceeds limit", i, ccbSizes[i])
			}
			buffers = append(buffers, NewPM4IndirectBuffer(ccbAddrs[i], ccbSizes[i], true))
		}

		buffers = append(buffers, NewPM4IndirectBuffer(dcbAddresses[i], dcbSize, false))
	}

	return buffers, nil
}
