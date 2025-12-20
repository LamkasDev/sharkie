package lib

import (
	"fmt"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x0000000000015990
// __int64 __fastcall sceKernelOpen(__int64, __int16, __int64, __int64, __int64, __int64, __m128, __m128, __m128, __m128, __m128, __m128, __m128, __m128)
func libKernel_sceKernelOpen(pathPtr uintptr, flags uintptr, mode uintptr) uintptr {
	err := libKernel_open(pathPtr, flags, mode)
	if err == ERR_PTR {
		return GetErrno() - 0x7FFE0000
	}

	return 0
}

// 0x000000000000DD50
// __int64 __fastcall open(__m128 _XMM0, __m128 _XMM1, __m128 _XMM2, __m128 _XMM3, __m128 _XMM4, __m128 _XMM5, __m128 _XMM6, __m128 _XMM7, __int64, __int16, __int64, __int64, __int64, __int64, char)
func libKernel_open(pathPtr uintptr, flags uintptr, mode uintptr) uintptr {
	// TODO: Mark thread as entering blocking syscall
	// Call the syscall
	// Check for cancellation

	return libKernel__open(pathPtr, flags, mode)
}

// 0x0000000000002750
// __int64 __fastcall open()
func libKernel__open(pathPtr uintptr, flags uintptr, mode uintptr) uintptr {
	return libKernel_sys_open(pathPtr, flags, mode)
}

func libKernel_sys_open(pathPtr uintptr, flags uintptr, mode uintptr) uintptr {
	if pathPtr == 0 {
		fmt.Printf("%-120s %s failed due to invalid path pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("_open"),
		)
		SetErrno(EFAULT)
		return ERR_PTR
	}

	path := ReadCString(pathPtr)
	fmt.Printf("%-120s %s opened file %s (flags=%s, mode=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("_open"),
		color.Blue.Sprint(path),
		color.Yellow.Sprintf("0x%X", flags),
		color.Yellow.Sprintf("0x%X", mode),
	)

	fd, ok := FileDescriptors[path]
	if !ok {
		SetErrno(ENOENT)
		return ERR_PTR
	}

	return fd
}
