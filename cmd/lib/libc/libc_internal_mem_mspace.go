package libc

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/libc"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000033C20
// __int64 __fastcall sceLibcMspaceMalloc(int *, char *, __m128, __int64, __int64, char *)
func libSceLibcInternal_sceLibcMspaceMalloc(handle, size uintptr) uintptr {
	GlobalMspaceAllocator.Lock.Lock()
	defer GlobalMspaceAllocator.Lock.Unlock()
	if mspace, ok := GlobalMspaceAllocator.Mspaces[handle]; ok {
		address := mspace.Allocator.MallocAligned(size, AllocationAlignment)
		if address == 0 {
			logger.Printf("%-132s %s failed due to allocation error.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceLibcMspaceMalloc"),
			)
			emu.SetErrno(ENOMEM)
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
		address := mspace.Allocator.MallocAligned(total, AllocationAlignment)
		if address == 0 {
			logger.Printf("%-132s %s failed due to allocation error.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceLibcMspaceRealloc"),
			)
			emu.SetErrno(ENOMEM)
		} else {
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
	if mspace, ok := GlobalMspaceAllocator.Mspaces[handle]; ok {
		mspace.Allocator.Free(ptr)
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
		address := mspace.Allocator.Realloc(ptr, newSize)
		if address == 0 && newSize != 0 {
			logger.Printf("%-132s %s failed due to allocation error.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceLibcMspaceRealloc"),
			)
			emu.SetErrno(ENOMEM)
		}
		return address
	}

	return libSceLibcInternal_realloc(ptr, newSize)
}

// 0x00000000000345A0
// __int64 __fastcall sceLibcMspaceReallocalign(__int64, __int64, __int64, __int64)
func libSceLibcInternal_sceLibcMspaceReallocalign(handle, alignment, ptr, newSize uintptr) uintptr {
	GlobalMspaceAllocator.Lock.Lock()
	defer GlobalMspaceAllocator.Lock.Unlock()
	if mspace, ok := GlobalMspaceAllocator.Mspaces[handle]; ok {
		newAddress := mspace.Allocator.MallocAligned(newSize, alignment)
		if newAddress == 0 {
			logger.Printf("%-132s %s failed due to allocation error.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceLibcMspaceReallocalign"),
			)
			emu.SetErrno(ENOMEM)
		} else if ptr != 0 {
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

	var name string
	if namePtr != nil {
		name = GoString(namePtr)
	} else {
		name = fmt.Sprintf("0x%X", base)
	}
	mspace := &MspaceInfo{
		Name:      name,
		Base:      base,
		End:       base + capacity,
		Allocator: NewGoAllocatorFromBuffer(base, capacity),
		Mutex:     sync.Mutex{},
	}
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
		address := mspace.Allocator.MallocAligned(size, alignment)
		if address == 0 {
			logger.Printf("%-132s %s failed due to allocation error.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceLibcMspaceMemalign"),
			)
			emu.SetErrno(ENOMEM)
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
		address := mspace.Allocator.MallocAligned(size, alignment)
		if address == 0 {
			logger.Printf("%-132s %s failed due to allocation error.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceLibcMspacePosixMemalign"),
			)
			emu.SetErrno(ENOMEM)
		}

		return address
	}

	return libSceLibcInternal_malloc(size)
}

// 0x0000000000034890
// _BOOL8 __fastcall sceLibcMspaceIsHeapEmpty(__int64, __int64, __int64)
func libSceLibcInternal_sceLibcMspaceIsHeapEmpty(_ /*handle*/, _ /*heapPtr*/ uintptr) uintptr {
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

// 0x0000000000034840
// __int64 sceLibcMspaceMallocStatsFast()
func libSceLibcInternal_sceLibcMspaceMallocStatsFast() uintptr {
	return 0
}

// 0x0000000000035610
// _BOOL8 __fastcall sceLibcPafMspaceIsHeapEmpty(__int64, __int64, __int64)
func libSceLibcInternal_sceLibcPafMspaceIsHeapEmpty(handle, heapPtr uintptr) uintptr {
	return libSceLibcInternal_sceLibcMspaceIsHeapEmpty(handle, heapPtr)
}

// 0x0000000000034850
// unsigned __int64 __fastcall sceLibcMspaceMallocUsableSize(__int64)
func libSceLibcInternal_sceLibcMspaceMallocUsableSize(ptr uintptr) uintptr {
	if ptr == 0 {
		return 0
	}
	GlobalMspaceAllocator.Lock.Lock()
	defer GlobalMspaceAllocator.Lock.Unlock()
	for _, mspace := range GlobalMspaceAllocator.Mspaces {
		if size := mspace.Allocator.UsableSize(ptr); size != 0 {
			return size
		}
	}

	return GlobalGoAllocator.UsableSize(ptr)
}
