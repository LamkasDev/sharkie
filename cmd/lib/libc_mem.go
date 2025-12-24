package lib

// 0x0000000000027A00
// __int64 malloc_init(void)
func libc__malloc_init() uintptr {
	return libSceLibcInternal__malloc_init()
}

// 0x0000000000027950
// __int64 malloc()
func libc_malloc(size uintptr) uintptr {
	return libSceLibcInternal_malloc(size)
}

// 0x0000000000027970
// __int64 calloc()
func libc_calloc(nmemb, size uintptr) uintptr {
	return libSceLibcInternal_calloc(nmemb, size)
}

// 0x0000000000027960
// __int64 free()
func libc_free(ptr uintptr) {
	libSceLibcInternal_free(ptr)
}

// 0x0000000000027980
// __int64 realloc()
func libc_realloc(ptr, newSize uintptr) uintptr {
	return libSceLibcInternal_realloc(ptr, newSize)
}

// 0x0000000000030FB0
// __int64 __fastcall sceLibcMspaceMalloc(int *, char *, __m128, __int64, __int64, char *)
func libc_sceLibcMspaceMalloc(mspace, size uintptr) uintptr {
	return libSceLibcInternal_sceLibcMspaceMalloc(mspace, size)
}

// 0x00000000000311F0
// __int64 __fastcall sceLibcMspaceCalloc(__int64, unsigned __int64, unsigned __int64, __int64)
func libc_sceLibcMspaceCalloc(mspace, nmemb, size uintptr) uintptr {
	return libSceLibcInternal_sceLibcMspaceCalloc(mspace, nmemb, size)
}

// 0x0000000000030FC0
// __int64 __fastcall sceLibcMspaceFree(__int64, __int64 *, __int64, __int64, __m128)
func libc_sceLibcMspaceFree(mspace, ptr uintptr) {
	libSceLibcInternal_sceLibcMspaceFree(mspace, ptr)
}

// 0x0000000000031270
// __int64 __fastcall sceLibcMspaceRealloc(__int64, __int64 *, unsigned __int64, __m128)
func libc_sceLibcMspaceRealloc(mspace, ptr, newSize uintptr) uintptr {
	return libSceLibcInternal_sceLibcMspaceRealloc(mspace, ptr, newSize)
}

// 0x0000000000030F90
// __int64 __fastcall sceLibcMspaceCreate(__int64, __int64, __int64, __int64)
func libc_sceLibcMspaceCreate(namePtr, base, capacity, flags uintptr) uintptr {
	return libSceLibcInternal_sceLibcMspaceCreate(namePtr, base, capacity, flags)
}

// 0x0000000000030FA0
// __int64 __fastcall sceLibcMspaceDestroy(__int64, __int64, __int64, __int64)
func libc_sceLibcMspaceDestroy() uintptr {
	return libSceLibcInternal_sceLibcMspaceDestroy()
}

// 0x0000000000031320
// __int64 __fastcall sceLibcMspaceMemalign(__int64, __int64, __int64)
func libc_sceLibcMspaceMemalign(mspace, alignment, size uintptr) uintptr {
	return libSceLibcInternal_sceLibcMspaceMemalign(mspace, alignment, size)
}

// 0x00000000000314C0
// _BOOL8 __fastcall sceLibcMspaceIsHeapEmpty(__int64, __int64, __int64)
func libc_sceLibcMspaceIsHeapEmpty(mspace, heapPtr uintptr) uintptr {
	return libSceLibcInternal_sceLibcMspaceIsHeapEmpty(mspace, heapPtr)
}
