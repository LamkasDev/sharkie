package lib

import (
	"time"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x0000000000028E70
// __int64 __fastcall malloc_init(__int64)
func libSceLibcInternal__malloc_init() uintptr {
	logger.Printf("%-132s %s initialized allocator.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("_malloc_init"),
	)
	return 0
}

var timeTemp = time.Now()

// 0x0000000000028D60
// __int64 malloc()
func libSceLibcInternal_malloc(size uintptr) uintptr {
	address := GlobalGoAllocator.Malloc(size)
	if address == 0 {
		logger.Printf("%-132s %s failed due to allocation error.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("malloc"),
		)
		return 0
	}
	if time.Since(timeTemp) > time.Second*5 {
		slot := unsafe.Slice((*uintptr)(unsafe.Add(unsafe.Pointer(emu.GetCurrentThread().Tcb.Self), -8)), 1)[0]
		logger.Printf("slot=0x%X\n", slot)
	}

	if logger.LogAlloc {
		logger.Printf("%-132s %s allocated %s bytes at %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("malloc"),
			color.Yellow.Sprintf("0x%X", size),
			color.Yellow.Sprintf("0x%X", address),
		)
	}
	return address
}

// 0x0000000000026AD0
// unsigned __int64 __fastcall memcpy(unsigned __int64 _RDI, __int64 _RSI, unsigned __int64 _RDX, __int64, __int64, __int64, char)
func libSceLibcInternal_memcpy(dst, src, n uintptr) uintptr {
	if n == 0 {
		return dst
	}

	dstSlice := unsafe.Slice((*byte)(unsafe.Pointer(dst)), n)
	srcSlice := unsafe.Slice((*byte)(unsafe.Pointer(src)), n)
	copy(dstSlice, srcSlice)

	return dst
}

// 0x0000000000027350
// unsigned __int64 __fastcall memset(unsigned __int64 _RDI, int _ESI, unsigned __int64 _RDX, double, __m128 _XMM1, __int64, __int64, __int64, char)
func libSceLibcInternal_memset(dst, c, n uintptr) uintptr {
	if n == 0 {
		return dst
	}

	dstSlice := unsafe.Slice((*byte)(unsafe.Pointer(dst)), n)
	fillValue := byte(c)
	for i := range dstSlice {
		dstSlice[i] = fillValue
	}

	return dst
}

// 0x0000000000028D80
// __int64 calloc()
func libSceLibcInternal_calloc(nmemb, size uintptr) uintptr {
	size *= nmemb
	return libSceLibcInternal_malloc(size)
}

// 0x0000000000028D70
// __int64 __fastcall free(_QWORD)
func libSceLibcInternal_free(ptr uintptr) uintptr {
	if ptr == 0 {
		logger.Printf("%-132s %s failed due to invalid pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("free"),
		)
		return 0
	}

	if !GlobalGoAllocator.Free(ptr) {
		logger.Printf("%-132s %s failed freeing untracked pointer %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("free"),
			color.Yellow.Sprintf("0x%X", ptr),
		)
		return 0
	}
	if logger.LogAlloc {
		logger.Printf("%-132s %s freed %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("free"),
			color.Yellow.Sprintf("0x%X", ptr),
		)
	}

	return 0
}

// 0x0000000000028D90
// __int64 realloc()
func libSceLibcInternal_realloc(ptr, newSize uintptr) uintptr {
	address := GlobalGoAllocator.Realloc(ptr, newSize)
	if address == 0 {
		logger.Printf("%-132s %s failed due to allocation error.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("realloc"),
		)
		return 0
	}

	if logger.LogAlloc {
		logger.Printf("%-132s %s reallocated %s to %s (newSize=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("realloc"),
			color.Yellow.Sprintf("0x%X", ptr),
			color.Yellow.Sprintf("0x%X", address),
			color.Yellow.Sprintf("0x%X", newSize),
		)
	}
	return address
}

// 0x0000000000033C20
// __int64 __fastcall sceLibcMspaceMalloc(int *, char *, __m128, __int64, __int64, char *)
func libSceLibcInternal_sceLibcMspaceMalloc(mspace, size uintptr) uintptr {
	return libSceLibcInternal_malloc(size)
}

// 0x0000000000034200
// __int64 __fastcall sceLibcMspaceCalloc(__int64, unsigned __int64, unsigned __int64, __int64)
func libSceLibcInternal_sceLibcMspaceCalloc(mspace, nmemb, size uintptr) uintptr {
	return libSceLibcInternal_calloc(nmemb, size)
}

// 0x0000000000033CF0
// __int64 __fastcall sceLibcMspaceFree(__int64, __int64 *, __int64, __int64, __m128)
func libSceLibcInternal_sceLibcMspaceFree(mspace, ptr uintptr) uintptr {
	return libSceLibcInternal_free(ptr)
}

// 0x0000000000034350
// __int64 __fastcall sceLibcMspaceRealloc(__int64, __int64 *, unsigned __int64, __m128)
func libSceLibcInternal_sceLibcMspaceRealloc(mspace, ptr, newSize uintptr) uintptr {
	return libSceLibcInternal_realloc(ptr, newSize)
}

// TODO: THIS IS JUST A PLACEHOLDER
// 0x00000000000345A0
// __int64 __fastcall sceLibcMspaceReallocalign(__int64, __int64, __int64, __int64)
func libSceLibcInternal_sceLibcMspaceReallocalign(mspace, alignment, ptr, newSize uintptr) uintptr {
	// TODO: handle actual alignment
	if alignment >= 4096 {
		logger.Printf("%-132s %s ignored allocation alignment (wanted=%s, got=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceLibcMspaceReallocalign"),
			color.Yellow.Sprintf("0x%X", alignment),
			color.Yellow.Sprintf("0x%X", 4096),
		)
	}
	return libSceLibcInternal_realloc(ptr, newSize)
}

// 0x000000000002F390
// __int64 __fastcall sceLibcMspaceCreate(__int64, __int64, __int64, __int64)
func libSceLibcInternal_sceLibcMspaceCreate(namePtr, base, capacity, flags uintptr) uintptr {
	return 0xCAFEBABE
}

// 0x0000000000033C10
// __int64 __fastcall sceLibcMspaceDestroy(__int64, __m128)
func libSceLibcInternal_sceLibcMspaceDestroy() uintptr {
	return 0
}

// 0x00000000000344A0
// __int64 __fastcall sceLibcMspaceMemalign(__int64, _QWORD *, __int64, __int64)
func libSceLibcInternal_sceLibcMspaceMemalign(mspace, alignment, size uintptr) uintptr {
	// TODO: handle actual alignment
	if alignment >= 4096 {
		logger.Printf("%-132s %s ignored allocation alignment (wanted=%s, got=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceLibcMspaceMemalign"),
			color.Yellow.Sprintf("0x%X", alignment),
			color.Yellow.Sprintf("0x%X", 4096),
		)
	}
	return libSceLibcInternal_malloc(size)
}

// TODO: THIS IS JUST A PLACEHOLDER
// 0x00000000000313C0
//
//	__int64 __fastcall sceLibcMspacePosixMemalign(__int64, __int64 *, unsigned __int64, unsigned __int64)
func libSceLibcInternal_sceLibcMspacePosixMemalign(mspace, alignment, size uintptr) uintptr {
	// TODO: handle actual alignment
	if alignment >= 4096 {
		logger.Printf("%-132s %s ignored allocation alignment (wanted=%s, got=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceLibcMspaceMemalign"),
			color.Yellow.Sprintf("0x%X", alignment),
			color.Yellow.Sprintf("0x%X", 4096),
		)
	}
	return libSceLibcInternal_malloc(size)
}

// 0x0000000000034890
// _BOOL8 __fastcall sceLibcMspaceIsHeapEmpty(__int64, __int64, __int64)
func libSceLibcInternal_sceLibcMspaceIsHeapEmpty(mspace, heapPtr uintptr) uintptr {
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
