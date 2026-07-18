package libc

import . "github.com/LamkasDev/sharkie/cmd/lib_structs"

// 0x000000000000A390
// __int64 __fastcall fopen(__int64, __int64)
func libc_fopen(pathPtr, modePtr Cstring) uintptr {
	return libSceLibcInternal_fopen(pathPtr, modePtr)
}

// 0x000000000000A9A0
// unsigned __int64 __fastcall fread(_BYTE *, unsigned __int64, unsigned __int64, __int64)
func libc_fread(ptr, size, n, filePtr uintptr) uintptr {
	return libSceLibcInternal_fread(ptr, size, n, filePtr)
}

// 0x000000000000ACA0
// __int64 __fastcall fseek(__int64, __int64, unsigned int)
func libc_fseek(filePtr, offset, whence uintptr) uintptr {
	return libSceLibcInternal_fseek(filePtr, offset, whence)
}

// 0x0000000000009F50
// __int64 __fastcall fgetpos(__int64, __int64)
func libc_fgetpos(filePtr, posPtr uintptr) uintptr {
	return libSceLibcInternal_fgetpos(filePtr, posPtr)
}

// 0x00000000000104C0
// __int64 __fastcall setvbuf(__int16 *, __int64, int, unsigned __int64)
func libc_setvbuf(filePtr, bufferPtr, mode, size uintptr) uintptr {
	return libSceLibcInternal_setvbuf(filePtr, bufferPtr, mode, size)
}

// 0x00000000000099F0
// __int64 __fastcall fclose(__int64)
func libc_fclose(fd uintptr) uintptr {
	return libSceLibcInternal_fclose(fd)
}
