package lib

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs/fs"
	"github.com/gookit/color"
)

// 0x0000000000015990
// __int64 __fastcall sceKernelOpen(__int64, __int16, __int64, __int64, __int64, __int64, __m128, __m128, __m128, __m128, __m128, __m128, __m128, __m128)
func libKernel_sceKernelOpen(pathPtr Cstring, flags FileFlags, mode FileMode) int32 {
	fd := libKernel_open(pathPtr, flags, mode)
	if fd == ERR_PTRI {
		return int32(GetErrno() - SonyErrorOffset)
	}

	return fd
}

// 0x000000000000DD50
// __int64 __fastcall open(__m128 _XMM0, __m128 _XMM1, __m128 _XMM2, __m128 _XMM3, __m128 _XMM4, __m128 _XMM5, __m128 _XMM6, __m128 _XMM7, __int64, __int16, __int64, __int64, __int64, __int64, char)
func libKernel_open(pathPtr Cstring, flags FileFlags, mode FileMode) int32 {
	// TODO: Mark thread as entering blocking syscall
	// Call the syscall
	// Check for cancellation

	return libKernel__open(pathPtr, flags, mode)
}

// 0x0000000000002750
// __int64 __fastcall open()
func libKernel__open(pathPtr Cstring, flags FileFlags, mode FileMode) int32 {
	if pathPtr == nil {
		logger.Printf("%-132s %s failed due to invalid path pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("_open"),
		)
		SetErrno(EFAULT)
		return ERR_PTRI
	}

	path := GetUsablePath(GoString(pathPtr))
	fd, err := GlobalFilesystem.Open(path, 0, mode)
	if err != nil {
		logger.Printf("%-132s %s failed due to open error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("_open"),
			color.Blue.Sprint(path),
			err.Error(),
		)
		SetErrno(ENOENT)
		return ERR_PTRI
	}

	if logger.LogFilesystem {
		logger.Printf("%-132s %s opened file %s (path=%s, flags=%s, mode=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("_open"),
			color.Yellow.Sprintf("0x%X", fd),
			color.Blue.Sprint(path),
			color.Yellow.Sprintf("0x%X", flags),
			color.Yellow.Sprintf("0x%X", mode),
		)
	}
	return int32(fd)
}
