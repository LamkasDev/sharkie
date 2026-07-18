package kernel

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/lib/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
)

// 0x0000000000015990
// __int64 __fastcall sceKernelOpen(__int64, __int16, __int64, __int64, __int64, __int64, __m128, __m128, __m128, __m128, __m128, __m128, __m128, __m128)
func libKernel_sceKernelOpen(pathPtr Cstring, flags FileFlags, mode FileMode) int32 {
	fd := posix.Open(pathPtr, flags, mode)
	if fd == ERR_PTRI {
		return int32(emu.GetErrno() - SonyErrorOffset)
	}

	return fd
}

// 0x0000000000015930
// __int64 __fastcall sceKernelRead(__int64, __int64, __int64)
func libKernel_sceKernelRead(fd FileDescriptor, bufPtr uintptr, length uint64) int64 {
	readBytes := posix.Read(fd, bufPtr, length)
	if readBytes == ERR_PTRI {
		return int64(emu.GetErrno() - SonyErrorOffset)
	}

	return readBytes
}

// 0x00000000000165B0
// __int64 sceKernelLseek()
func libKernel_sceKernelLseek(fd FileDescriptor, offset int64, whence int32) int64 {
	newOffset := posix.Lseek(fd, offset, whence)
	if newOffset == ERR_PTRI {
		return int64(emu.GetErrno() - SonyErrorOffset)
	}

	return newOffset
}

// 0x00000000000159C0
// __int64 __fastcall sceKernelClose(__int64)
func libKernel_sceKernelClose(fd FileDescriptor) int32 {
	err := posix.Close(fd)
	if err != 0 {
		return int32(emu.GetErrno() - SonyErrorOffset)
	}

	return 0
}
