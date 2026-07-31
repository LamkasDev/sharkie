package kernel

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/lib/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
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

// 0x0000000000016520
// __int64 sceKernelPread()
func libKernel_sceKernelPread(fd FileDescriptor, bufPtr uintptr, length uint64, offset int64) int64 {
	readBytes := posix.Pread(fd, bufPtr, length, offset)
	if readBytes == ERR_PTRI {
		return int64(emu.GetErrno() - SonyErrorOffset)
	}

	return readBytes
}

// 0x0000000000015960
// __int64 __fastcall sceKernelWrite(__int64, __int64, __int64)
func libKernel_sceKernelWrite(fd FileDescriptor, bufPtr uintptr, length uint64) int64 {
	wroteBytes := posix.Write(fd, bufPtr, length)
	if wroteBytes == ERR_PTRI {
		return int64(emu.GetErrno() - SonyErrorOffset)
	}

	return wroteBytes
}

// 0x0000000000016550
// __int64 sceKernelPwrite()
func libKernel_sceKernelPwrite(fd FileDescriptor, bufPtr uintptr, length uint64, offset int64) int64 {
	wroteBytes := posix.Pwrite(fd, bufPtr, length, offset)
	if wroteBytes == ERR_PTRI {
		return int64(emu.GetErrno() - SonyErrorOffset)
	}

	return wroteBytes
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

// 0x00000000000163D0
// __int64 __fastcall sceKernelStat(__int64, __int64)
func libKernel_sceKernelStat(pathPtr Cstring, stat *FileStat) int32 {
	err := posix.Stat(pathPtr, stat)
	if err != 0 {
		return int32(emu.GetErrno() - SonyErrorOffset)
	}

	return 0
}

// 0x0000000000016400
// __int64 __fastcall sceKernelFstat(__int64, __int64)
func libKernel_sceKernelFstat(fd FileDescriptor, stat *FileStat) int32 {
	err := posix.Fstat(fd, stat)
	if err != 0 {
		return int32(emu.GetErrno() - SonyErrorOffset)
	}

	return 0
}

// 0x00000000000165E0
// __int64 sceKernelTruncate()
func libKernel_sceKernelTruncate(pathPtr Cstring, length int64) int32 {
	err := posix.Truncate(pathPtr, length)
	if err != 0 {
		return int32(emu.GetErrno() - SonyErrorOffset)
	}

	return 0
}

// 0x0000000000016610
// __int64 sceKernelFtruncate()
func libKernel_sceKernelFtruncate(fd FileDescriptor, length int64) int32 {
	err := posix.Ftruncate(fd, length)
	if err != 0 {
		return int32(emu.GetErrno() - SonyErrorOffset)
	}

	return 0
}

// 0x0000000000000970
// __int64 __fastcall ioctl()
func libKernel_ioctl(fd FileDescriptor, request uint64, argPtr uintptr) int32 {
	err := posix.Ioctl(fd, request, argPtr)
	if err != 0 {
		return int32(emu.GetErrno() - SonyErrorOffset)
	}

	return 0
}

// 0x0000000000001750
// __int64 __fastcall shm_open()
func libKernel_shm_open(pathPtr Cstring, flags FileFlags, mode FileMode) int32 {
	err := posix.Shm_open(pathPtr, flags, mode)
	if err != 0 {
		return int32(emu.GetErrno() - SonyErrorOffset)
	}

	return 0
}

// 0x0000000000016640
// __int64 __fastcall sceKernelCheckReachability(char *)
func libKernel_sceKernelCheckReachability(pathPtr Cstring) uintptr {
	if pathPtr == nil {
		logger.Printf("%-132s %s failed due to invalid path pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelCheckReachability"),
		)
		return SCE_KERNEL_ERROR_ENOENT
	}

	path := GlobalFilesystem.GetUsablePath(GoString(pathPtr))
	exists := GlobalFilesystem.Exists(path)
	if !exists {
		return SCE_KERNEL_ERROR_ENOENT
	}

	return 0
}
