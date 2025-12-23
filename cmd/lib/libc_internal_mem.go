package lib

import (
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x0000000000028D60
// __int64 malloc()
func libSceLibcInternal_malloc(size uintptr) uintptr {
	// Make sure to return a valid pointer, even for size 0.
	if size == 0 {
		size = 1
	}
	addr := libKernel_mmap(0, size, PROT_READ|PROT_WRITE, MAP_ANON|MAP_PRIVATE, ERR_PTR, 0)
	if addr == ERR_PTR {
		return 0
	}

	GlobalAllocator.Lock.Lock()
	defer GlobalAllocator.Lock.Unlock()
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
		fmt.Printf("%-120s %s failed due to invalid pointer %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("free"),
			color.Yellow.Sprintf("0x%X", ptr),
		)
		return
	}

	GlobalAllocator.Lock.Lock()
	defer GlobalAllocator.Lock.Unlock()
	size, ok := GlobalAllocator.Allocations[ptr]
	if !ok {
		fmt.Printf("%-120s %s failed freeing untracked pointer %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("free"),
			color.Yellow.Sprintf("0x%X", ptr),
		)
		return
	}
	ret := libKernel_munmap(ptr, size)
	if ret == ERR_PTR {
		return
	}

	delete(GlobalAllocator.Allocations, ptr)
	fmt.Printf("%-120s %s freed %s bytes at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("free"),
		color.Yellow.Sprintf("0x%X", size),
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

	GlobalAllocator.Lock.Lock()
	defer GlobalAllocator.Lock.Unlock()
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
	copy(
		unsafe.Slice((*byte)(unsafe.Pointer(newPtr)), copySize),
		unsafe.Slice((*byte)(unsafe.Pointer(ptr)), copySize),
	)

	fmt.Printf("%-120s %s reallocated %s bytes from %s to %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("realloc"),
		color.Yellow.Sprintf("0x%X", copySize),
		color.Yellow.Sprintf("0x%X", ptr),
		color.Yellow.Sprintf("0x%X", newPtr),
	)
	libSceLibcInternal_free(ptr)
	return newPtr
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
func libSceLibcInternal_sceLibcMspaceFree(mspace, ptr uintptr) {
	libSceLibcInternal_free(ptr)
}

// 0x0000000000034350
// __int64 __fastcall sceLibcMspaceRealloc(__int64, __int64 *, unsigned __int64, __m128)
func libSceLibcInternal_sceLibcMspaceRealloc(mspace, ptr, newSize uintptr) uintptr {
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
	if alignment > 4096 {
		fmt.Printf("%-120s %s ignored allocation alignment (wanted=%s, got=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceLibcMspaceMemalign"),
			color.Yellow.Sprintf("0x%X", alignment),
			color.Yellow.Sprintf("0x%X", 4096),
		)
	}
	return libSceLibcInternal_malloc(size)
}
