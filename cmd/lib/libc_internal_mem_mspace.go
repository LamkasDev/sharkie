package lib

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
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
	base    uintptr
	end     uintptr
	current uintptr
	mu      sync.Mutex
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
	ms.mu.Lock()
	defer ms.mu.Unlock()

	alignedAddress := (ms.current + alignment - 1) &^ (alignment - 1)
	if alignedAddress+size > ms.end {
		return 0, fmt.Errorf("lack of space")
	}
	ms.current = alignedAddress + size

	return alignedAddress, nil
}

// 0x0000000000033C20
// __int64 __fastcall sceLibcMspaceMalloc(int *, char *, __m128, __int64, __int64, char *)
func libSceLibcInternal_sceLibcMspaceMalloc(handle, size uintptr) uintptr {
	GlobalMspaceAllocator.Lock.Lock()
	defer GlobalMspaceAllocator.Lock.Unlock()
	if mspace, ok := GlobalMspaceAllocator.Mspaces[handle]; ok {
		address, err := mspace.Alloc(16, size)
		if err != nil {
			logger.Printf("%-132s %s failed due to alloc error (%s).\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceLibcMspaceMalloc"),
				err.Error(),
			)
		}

		return address
	}

	return libSceLibcInternal_malloc(size)
}

// 0x0000000000034200
// __int64 __fastcall sceLibcMspaceCalloc(__int64, unsigned __int64, unsigned __int64, __int64)
func libSceLibcInternal_sceLibcMspaceCalloc(handle, nmemb, size uintptr) uintptr {
	GlobalMspaceAllocator.Lock.Lock()
	defer GlobalMspaceAllocator.Lock.Unlock()
	total := nmemb * size
	if mspace, ok := GlobalMspaceAllocator.Mspaces[handle]; ok {
		address, err := mspace.Alloc(16, total)
		if err != nil {
			logger.Printf("%-132s %s failed due to alloc error (%s).\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceLibcMspaceRealloc"),
				err.Error(),
			)
		}
		if address != 0 {
			dstSlice := unsafe.Slice((*byte)(unsafe.Pointer(address)), total)
			for i := range dstSlice {
				dstSlice[i] = 0
			}
		}

		return address
	}

	return libSceLibcInternal_calloc(nmemb, size)
}

// 0x0000000000033CF0
// __int64 __fastcall sceLibcMspaceFree(__int64, __int64 *, __int64, __int64, __m128)
func libSceLibcInternal_sceLibcMspaceFree(handle, ptr uintptr) uintptr {
	GlobalMspaceAllocator.Lock.Lock()
	defer GlobalMspaceAllocator.Lock.Unlock()
	if _, ok := GlobalMspaceAllocator.Mspaces[handle]; ok {
		// TODO: add free if we ever change from bump-allocator.
		return 0
	}

	return libSceLibcInternal_free(ptr)
}

// 0x0000000000034350
// __int64 __fastcall sceLibcMspaceRealloc(__int64, __int64 *, unsigned __int64, __m128)
func libSceLibcInternal_sceLibcMspaceRealloc(handle, ptr, newSize uintptr) uintptr {
	GlobalMspaceAllocator.Lock.Lock()
	defer GlobalMspaceAllocator.Lock.Unlock()
	if mspace, ok := GlobalMspaceAllocator.Mspaces[handle]; ok {
		newAddress, err := mspace.Alloc(16, newSize)
		if err != nil {
			logger.Printf("%-132s %s failed due to alloc error (%s).\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceLibcMspaceRealloc"),
				err.Error(),
			)
		}
		if newAddress != 0 && ptr != 0 {
			oldSlice := unsafe.Slice((*byte)(unsafe.Pointer(ptr)), newSize)
			newSlice := unsafe.Slice((*byte)(unsafe.Pointer(newAddress)), newSize)
			copy(newSlice, oldSlice)
		}

		return newAddress
	}

	return libSceLibcInternal_realloc(ptr, newSize)
}

// 0x00000000000345A0
// __int64 __fastcall sceLibcMspaceReallocalign(__int64, __int64, __int64, __int64)
func libSceLibcInternal_sceLibcMspaceReallocalign(handle, alignment, ptr, newSize uintptr) uintptr {
	GlobalMspaceAllocator.Lock.Lock()
	defer GlobalMspaceAllocator.Lock.Unlock()
	if mspace, ok := GlobalMspaceAllocator.Mspaces[handle]; ok {
		newAddress, err := mspace.Alloc(alignment, newSize)
		if err != nil {
			logger.Printf("%-132s %s failed due to alloc error (%s).\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceLibcMspaceReallocalign"),
				err.Error(),
			)
		}
		if newAddress != 0 && ptr != 0 {
			oldSlice := unsafe.Slice((*byte)(unsafe.Pointer(ptr)), newSize)
			newSlice := unsafe.Slice((*byte)(unsafe.Pointer(newAddress)), newSize)
			copy(newSlice, oldSlice)
		}

		return newAddress
	}

	return libSceLibcInternal_realloc(ptr, newSize)
}

// 0x000000000002F390
// __int64 __fastcall sceLibcMspaceCreate(__int64, __int64, __int64, __int64)
func libSceLibcInternal_sceLibcMspaceCreate(namePtr Cstring, base, capacity, _ /*flags*/ uintptr) uintptr {
	if base == 0 || capacity == 0 {
		logger.Printf("%-132s %s failed due to invalid base or zero capacity.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceLibcMspaceCreate"),
		)
		return 0
	}
	GlobalMspaceAllocator.Lock.Lock()
	defer GlobalMspaceAllocator.Lock.Unlock()

	mspace := &MspaceInfo{
		base:    base,
		end:     base + capacity,
		current: base,
		mu:      sync.Mutex{},
	}
	var name string
	if namePtr != nil {
		name = GoString(namePtr)
	} else {
		name = fmt.Sprintf("0x%X", base)
	}
	mspace.Name = name
	GlobalMspaceAllocator.Mspaces[base] = mspace

	logger.Printf("%-132s %s created mspace %s (base=%s, capacity=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceLibcMspaceCreate"),
		color.Blue.Sprint(mspace.Name),
		color.Yellow.Sprintf("0x%X", base),
		color.Yellow.Sprintf("0x%X", capacity),
	)
	return base
}

// 0x0000000000033C10
// __int64 __fastcall sceLibcMspaceDestroy(__int64, __m128)
func libSceLibcInternal_sceLibcMspaceDestroy() uintptr {
	return 0
}

// 0x00000000000344A0
// __int64 __fastcall sceLibcMspaceMemalign(__int64, _QWORD *, __int64, __int64)
func libSceLibcInternal_sceLibcMspaceMemalign(handle, alignment, size uintptr) uintptr {
	GlobalMspaceAllocator.Lock.Lock()
	defer GlobalMspaceAllocator.Lock.Unlock()
	if mspace, ok := GlobalMspaceAllocator.Mspaces[handle]; ok {
		address, err := mspace.Alloc(alignment, size)
		if err != nil {
			logger.Printf("%-132s %s failed due to alloc error (%s).\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceLibcMspaceMemalign"),
				err.Error(),
			)
		}

		return address
	}

	return libSceLibcInternal_malloc(size)
}

// 0x00000000000313C0
// __int64 __fastcall sceLibcMspacePosixMemalign(__int64, __int64 *, unsigned __int64, unsigned __int64)
func libSceLibcInternal_sceLibcMspacePosixMemalign(handle, alignment, size uintptr) uintptr {
	GlobalMspaceAllocator.Lock.Lock()
	defer GlobalMspaceAllocator.Lock.Unlock()
	if mspace, ok := GlobalMspaceAllocator.Mspaces[handle]; ok {
		address, err := mspace.Alloc(alignment, size)
		if err != nil {
			logger.Printf("%-132s %s failed due to alloc error (%s).\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceLibcMspacePosixMemalign"),
				err.Error(),
			)
		}

		return address
	}

	return libSceLibcInternal_malloc(size)
}

// 0x0000000000034890
// _BOOL8 __fastcall sceLibcMspaceIsHeapEmpty(__int64, __int64, __int64)
func libSceLibcInternal_sceLibcMspaceIsHeapEmpty(_ /*mspace*/, _ /*heapPtr*/ uintptr) uintptr {
	isEmpty := uintptr(0)
	logger.Printf("%-132s %s returned %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceLibcMspaceIsHeapEmpty"),
		color.Yellow.Sprintf("0x%X", isEmpty),
	)
	return isEmpty
}

// 0x0000000000034830
// __int64 sceLibcMspaceMallocStats()
func libSceLibcInternal_sceLibcMspaceMallocStats() uintptr {
	return 0
}

// 0x:0000000000034840
// __int64 sceLibcMspaceMallocStatsFast()
func libSceLibcInternal_sceLibcMspaceMallocStatsFast() uintptr {
	return 0
}

// 0x0000000000035610
// _BOOL8 __fastcall sceLibcPafMspaceIsHeapEmpty(__int64, __int64, __int64)
func libSceLibcInternal_sceLibcPafMspaceIsHeapEmpty(mspace, heapPtr uintptr) uintptr {
	return libSceLibcInternal_sceLibcMspaceIsHeapEmpty(mspace, heapPtr)
}

func SetupMspaceAllocator() {
	GlobalMspaceAllocator = NewMspaceAllocator()
}
