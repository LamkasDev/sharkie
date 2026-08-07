package translation

import (
	"fmt"
	"sort"

	"github.com/LamkasDev/sharkie/cmd/vulkan"
	vk "github.com/goki/vulkan"
)

// DirectAllocation represents a chunk of direct memory mapped to a Vulkan buffer.
type DirectAllocation struct {
	Buffer        vk.Buffer
	Memory        vk.DeviceMemory
	DeviceAddress uint64
	Length        uint64
}

type AddressTranslationEntry struct {
	GuestBase     uint64
	GuestEnd      uint64
	DeviceAddress uint64
	Pad           uint64
}

func (t *GpuTranslator) GetBufferAddress(buffer vk.Buffer) uint64 {
	return vulkan.GetBufferDeviceAddress(t.handles.Instance, t.handles.Device, buffer)
}

func (t *GpuTranslator) GetLinearBuffer(address uintptr) (vk.Buffer, uintptr, error) {
	t.directAllocationsMutex.Lock()
	defer t.directAllocationsMutex.Unlock()
	for base, alloc := range t.directAllocations {
		if address >= base && address < base+uintptr(alloc.Length) {
			return alloc.Buffer, address - base, nil
		}
	}

	return vk.NullBuffer, 0, fmt.Errorf("address 0x%X not in any known allocator", address)
}

func (t *GpuTranslator) TranslateToHostAddress(address uintptr) uintptr {
	t.directAllocationsMutex.Lock()
	defer t.directAllocationsMutex.Unlock()
	for base, alloc := range t.directAllocations {
		if address >= base && address < base+uintptr(alloc.Length) {
			return address
		}
	}

	return 0
}

func (t *GpuTranslator) updateAddressTranslationSSBO() {
	if t.addressTranslationMap == nil {
		return
	}
	offsets := make([]uintptr, 0, len(t.directAllocations))
	for offset := range t.directAllocations {
		offsets = append(offsets, offset)
	}
	sort.Slice(offsets, func(i, j int) bool {
		return offsets[i] < offsets[j]
	})

	for i, offset := range offsets {
		alloc := t.directAllocations[offset]
		t.addressTranslationMap[i].DeviceAddress = alloc.DeviceAddress
		t.addressTranslationMap[i].GuestEnd = uint64(offset) + alloc.Length
		t.addressTranslationMap[i].GuestBase = uint64(offset)
		if offset == 0 || alloc.DeviceAddress == 0 {
			panic("nope")
		}
	}
	t.addressTranslationMap[len(offsets)].GuestBase = ^uint64(0)
}
