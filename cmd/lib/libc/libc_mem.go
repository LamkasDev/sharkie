package libc

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

// 0x000000000001C850
// unsigned __int64 __fastcall memcpy(unsigned __int64 _RDI, __int64 _RSI, unsigned __int64 _RDX, _DWORD, _DWORD, _DWORD, char)
func libc_memcpy(dst, src, n uintptr) uintptr {
	return libSceLibcInternal_memcpy(dst, src, n)
}

// 0x000000000001D0D0
// unsigned __int64 __fastcall memset(unsigned __int64 _RDI, int _ESI, unsigned __int64 _RDX, _DWORD, _DWORD, _DWORD, double, __m128 _XMM1, char)
func libc_memset(dst, c, n uintptr) uintptr {
	return libSceLibcInternal_memset(dst, c, n)
}

// 0x0000000000027970
// __int64 calloc()
func libc_calloc(nmemb, size uintptr) uintptr {
	return libSceLibcInternal_calloc(nmemb, size)
}

// 0x0000000000027960
// __int64 free()
func libc_free(ptr uintptr) uintptr {
	return libSceLibcInternal_free(ptr)
}

// 0x0000000000027980
// __int64 realloc()
func libc_realloc(ptr, newSize uintptr) uintptr {
	return libSceLibcInternal_realloc(ptr, newSize)
}

// 0x00000000000279A0
// __int64 memalign()
func libc_memalign(alignment, size uintptr) uintptr {
	return libSceLibcInternal_memalign(alignment, size)
}

// 0x0000000000027990
// __int64 __fastcall aligned_alloc(_QWORD, _QWORD)
func libc_aligned_alloc(alignment, size uintptr) uintptr {
	return libSceLibcInternal_aligned_alloc(alignment, size)
}

// __int64 reallocalign()
func libc_reallocalign(ptr, newSize, alignment uintptr) uintptr {
	return libSceLibcInternal_reallocalign(ptr, newSize, alignment)
}
