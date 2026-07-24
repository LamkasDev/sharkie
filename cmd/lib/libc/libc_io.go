package libc

import (
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
)

// 0x000000000000A390
// __int64 __fastcall fopen(__int64, __int64)
func libc_fopen(pathPtr, modePtr Cstring) uintptr {
	return libSceLibcInternal_fopen(pathPtr, modePtr)
}

// 0x0000000000027530
// __int64 __fastcall fdopen(__int64, __int64)
func libc_fdopen(fd FileDescriptor, modePtr Cstring) uintptr {
	return libSceLibcInternal_fdopen(fd, modePtr)
}

// 0x000000000000A9A0
// unsigned __int64 __fastcall fread(_BYTE *, unsigned __int64, unsigned __int64, __int64)
func libc_fread(ptr, size, n, filePtr uintptr) uintptr {
	return libSceLibcInternal_fread(ptr, size, n, filePtr)
}

// 0x0000000000009DD0
// __int64 __fastcall fgetc(__int64)
func libc_fgetc(filePtr uintptr) uintptr {
	return libSceLibcInternal_fgetc(filePtr)
}

// 0x0000000000012480
// __int64 __fastcall ungetc(unsigned int, __int64)
func libc_ungetc(c, filePtr uintptr) uintptr {
	return libSceLibcInternal_ungetc(c, filePtr)
}

// 0x000000000000AF90
// unsigned __int64 __fastcall fwrite(__int64, unsigned __int64, unsigned __int64, __int64)
func libc_fwrite(ptr, size, n, filePtr uintptr) uintptr {
	return libSceLibcInternal_fwrite(ptr, size, n, filePtr)
}

// 0x000000000000A510
// __int64 __fastcall fputc(unsigned __int8, __int64)
func libc_fputc(c, filePtr uintptr) uintptr {
	return libSceLibcInternal_fputc(c, filePtr)
}

// 0x000000000000A6B0
// __int64 __fastcall fputs(_BYTE *, __int64)
func libc_fputs(sPtr Cstring, filePtr uintptr) uintptr {
	return libSceLibcInternal_fputs(sPtr, filePtr)
}

// 0x000000000000C3D0
// __int64 __fastcall putchar(unsigned __int8)
func libc_putchar(c uintptr) uintptr {
	return libSceLibcInternal_putchar(c)
}

// 0x000000000000C3F0
// __int64 __fastcall puts(_BYTE *)
func libc_puts(sPtr Cstring) uintptr {
	return libSceLibcInternal_puts(sPtr)
}

// 0x0000000000009B80
// __int64 __fastcall fflush(__int16 *)
func libc_fflush(filePtr uintptr) uintptr {
	return libSceLibcInternal_fflush(filePtr)
}

// 0x000000000000ACA0
// __int64 __fastcall fseek(__int64, __int64, unsigned int)
func libc_fseek(filePtr, offset, whence uintptr) uintptr {
	return libSceLibcInternal_fseek(filePtr, offset, whence)
}

// 0x000000000000AD80
// __int64 __fastcall ftell(__int64)
func libc_ftell(filePtr uintptr) uintptr {
	return libSceLibcInternal_ftell(filePtr)
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
func libc_fclose(filePtr uintptr) uintptr {
	return libSceLibcInternal_fclose(filePtr)
}

// 0x0000000000009AB0
// __int64 __fastcall feof(_WORD *)
func libc_feof(filePtr uintptr) uintptr {
	return libSceLibcInternal_feof(filePtr)
}

func libc__Lockfilelock(filePtr uintptr) uintptr {
	return libSceLibcInternal__Lockfilelock(filePtr)
}

func libc__Unlockfilelock(filePtr uintptr) uintptr {
	return libSceLibcInternal__Unlockfilelock(filePtr)
}
