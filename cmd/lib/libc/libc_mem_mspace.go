package libc

import . "github.com/LamkasDev/sharkie/cmd/lib_structs"

// 0x0000000000030FB0
// __int64 __fastcall sceLibcMspaceMalloc(int *, char *, __m128, __int64, __int64, char *)
func libc_sceLibcMspaceMalloc(handle, size uintptr) uintptr {
	return libSceLibcInternal_sceLibcMspaceMalloc(handle, size)
}

// 0x00000000000311F0
// __int64 __fastcall sceLibcMspaceCalloc(__int64, unsigned __int64, unsigned __int64, __int64)
func libc_sceLibcMspaceCalloc(handle, nmemb, size uintptr) uintptr {
	return libSceLibcInternal_sceLibcMspaceCalloc(handle, nmemb, size)
}

// 0x0000000000030FC0
// __int64 __fastcall sceLibcMspaceFree(__int64, __int64 *, __int64, __int64, __m128)
func libc_sceLibcMspaceFree(handle, ptr uintptr) uintptr {
	return libSceLibcInternal_sceLibcMspaceFree(handle, ptr)
}

// 0x0000000000031270
// __int64 __fastcall sceLibcMspaceRealloc(__int64, __int64 *, unsigned __int64, __m128)
func libc_sceLibcMspaceRealloc(handle, ptr, newSize uintptr) uintptr {
	return libSceLibcInternal_sceLibcMspaceRealloc(handle, ptr, newSize)
}

// 0x0000000000031350
// __int64 __fastcall sceLibcMspaceReallocalign(__int64, __int64, __int64, __int64)
func libc_sceLibcMspaceReallocalign(handle, alignment, ptr, newSize uintptr) uintptr {
	return libSceLibcInternal_sceLibcMspaceReallocalign(handle, alignment, ptr, newSize)
}

// 0x0000000000030F90
// __int64 __fastcall sceLibcMspaceCreate(__int64, __int64, __int64, __int64)
func libc_sceLibcMspaceCreate(namePtr Cstring, base, capacity, flags uintptr) uintptr {
	return libSceLibcInternal_sceLibcMspaceCreate(namePtr, base, capacity, flags)
}

// 0x0000000000030FA0
// __int64 __fastcall sceLibcMspaceDestroy(__int64, __int64, __int64, __int64)
func libc_sceLibcMspaceDestroy() uintptr {
	return libSceLibcInternal_sceLibcMspaceDestroy()
}

// 0x0000000000031320
// __int64 __fastcall sceLibcMspaceMemalign(__int64, __int64, __int64)
func libc_sceLibcMspaceMemalign(handle, alignment, size uintptr) uintptr {
	return libSceLibcInternal_sceLibcMspaceMemalign(handle, alignment, size)
}

// 0x00000000000313C0
// __int64 __fastcall sceLibcMspacePosixMemalign(unsigned __int64, __int64 *, unsigned __int64, unsigned __int64, __m128)
func libc_sceLibcMspacePosixMemalign(handle, alignment, size uintptr) uintptr {
	return libSceLibcInternal_sceLibcMspacePosixMemalign(handle, alignment, size)
}

// 0x00000000000314C0
// _BOOL8 __fastcall sceLibcMspaceIsHeapEmpty(__int64, __int64, __int64)
func libc_sceLibcMspaceIsHeapEmpty(handle, heapPtr uintptr) uintptr {
	return libSceLibcInternal_sceLibcMspaceIsHeapEmpty(handle, heapPtr)
}

// 0x0000000000031460
// __int64 sceLibcMspaceMallocStats()
func libc_sceLibcMspaceMallocStats() uintptr {
	return libSceLibcInternal_sceLibcMspaceMallocStats()
}

// 0x0000000000031460
// __int64 sceLibcMspaceMallocStatsFast()
func libc_sceLibcMspaceMallocStatsFast() uintptr {
	return libSceLibcInternal_sceLibcMspaceMallocStatsFast()
}

// 0x0000000000031480
// unsigned __int64 __fastcall sceLibcMspaceMallocUsableSize(__int64)
func libc_sceLibcMspaceMallocUsableSize(ptr uintptr) uintptr {
	return libSceLibcInternal_sceLibcMspaceMallocUsableSize(ptr)
}
