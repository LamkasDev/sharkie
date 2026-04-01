package lib

import (
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

const ErrnoTcbOffset = 0x188

// GetErrnoAddress returns address of the errno variable for current thread.
func GetErrnoAddress() uintptr {
	thread := emu.GetCurrentThread()
	return uintptr(unsafe.Pointer(thread.Tcb)) + ErrnoTcbOffset
}

// GetErrno returns error number of current thread.
func GetErrno() uintptr {
	errNoAddr := GetErrnoAddress()
	errNoSlice := unsafe.Slice((*byte)(unsafe.Pointer(errNoAddr)), 8)
	return uintptr(binary.LittleEndian.Uint64(errNoSlice))
}

// SetErrno sets error number for current thread.
func SetErrno(err uintptr) {
	errNoAddr := GetErrnoAddress()
	errNoSlice := unsafe.Slice((*byte)(unsafe.Pointer(errNoAddr)), 8)
	binary.LittleEndian.PutUint64(errNoSlice, uint64(err))
}

// 0x0000000000002C70
// void *_error()
func libKernel___error() uintptr {
	address := GetErrnoAddress()
	if logger.LogErrorRet {
		logger.Printf("%-132s %s returned %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("_error"),
			color.Red.Sprintf("0x%X", address),
		)
	}

	return address
}

// 0x0000000000014E50
// __int64 __fastcall sceKernelError(int)
func libKernel_sceKernelError(err uintptr) uintptr {
	if err != 0 {
		err -= SonyErrorOffset
	}
	if logger.LogErrorRet {
		logger.Printf("%-132s %s returned %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelError"),
			color.Red.Sprintf("0x%X", err),
		)
	}

	return err
}

// 0x0000000000022D40
// __int64 __fastcall sceKernelDebugRaiseException(__int64, __int64)
func libKernel_sceKernelDebugRaiseException(err, argsPtr uintptr) uintptr {
	logger.Printf("%-132s %s called with %s, exiting...\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelDebugRaiseException"),
		color.Red.Sprintf("0x%X", err),
	)
	logger.CleanupAndExit()

	return 0
}
