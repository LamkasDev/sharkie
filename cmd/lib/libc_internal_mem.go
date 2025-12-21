package lib

import (
	"fmt"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x0000000000028D60
// __int64 malloc()
func libSceLibcInternal_malloc(size uintptr) uintptr {
	addr := libKernel_mmap(0, size, PROT_READ|PROT_WRITE, MAP_ANON|MAP_PRIVATE, ERR_PTR, 0)
	if addr == ERR_PTR {
		return 0
	}
	GlobalAllocator.Allocations[addr] = size

	return addr
}

// 0x0000000000028D80
// __int64 calloc()
func libSceLibcInternal_calloc(nmemb, size uintptr) uintptr {
	size *= nmemb
	return libSceLibcInternal_malloc(size)
}

// 0x0000000000028D70
// __int64 __fastcall free(_QWORD)
func libSceLibcInternal_free(ptr uintptr) {
	if ptr == 0 {
		fmt.Printf("%-120s %s failed due to invalid pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("free"),
		)
		return
	}

	_, ok := GlobalAllocator.Allocations[ptr]
	if !ok {
		fmt.Printf("%-120s %s failed freeing untracked pointer %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("free"),
			color.Yellow.Sprintf("0x%X", ptr),
		)
		return
	}

	delete(GlobalAllocator.Allocations, ptr)
	fmt.Printf("%-120s %s freed memory at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("free"),
		color.Yellow.Sprintf("0x%X", ptr),
	)
}

// 0x0000000000028D90
// __int64 realloc()
func libSceLibcInternal_realloc(ptr, newSize uintptr) uintptr {
	if ptr == 0 {
		return libSceLibcInternal_malloc(newSize)
	}
	if newSize == 0 {
		libSceLibcInternal_free(ptr)
		return 0
	}

	oldSize, ok := GlobalAllocator.Allocations[ptr]
	if !ok {
		fmt.Printf("%-120s %s failed reallocating untracked pointer %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("realloc"),
			color.Yellow.Sprintf("0x%X", ptr),
		)
		return 0
	}

	newPtr := libSceLibcInternal_malloc(newSize)
	if newPtr == 0 {
		return 0
	}
	copySize := oldSize
	if newSize < oldSize {
		copySize = newSize
	}

	fmt.Printf("%-120s %s reallocating %s xd.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("realloc"),
		color.Yellow.Sprintf("0x%X", copySize),
	)

	libSceLibcInternal_free(ptr)

	return newPtr
}

// 0x00000000000311F0
// __int64 __fastcall sceLibcMspaceCalloc(__int64, unsigned __int64, unsigned __int64, __int64)
func libSceLibcInternal_sceLibcMspaceCalloc(mspace, nmemb, size uintptr) uintptr {
	return libSceLibcInternal_calloc(nmemb, size)
}

// 0x0000000000033CF0
// __int64 __fastcall sceLibcMspaceFree(__int64, __int64 *, __int64, __int64, __m128)
func libSceLibcInternal_sceLibcMspaceFree(ptr uintptr) {
	libSceLibcInternal_free(ptr)
}
