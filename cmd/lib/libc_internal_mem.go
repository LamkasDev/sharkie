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
