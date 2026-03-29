package lib

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs/fs"
	"github.com/gookit/color"
)

// 0x00000000000159C0
// __int64 __fastcall sceKernelClose(__int64)
func libKernel_sceKernelClose(fd uintptr) uintptr {
	err := libKernel_close(fd)
	if err != 0 {
		return GetErrno() - SonyErrorOffset
	}

	return 0
}

// 0x000000000000D950
// __int64 __fastcall close(unsigned int)
func libKernel_close(fd uintptr) uintptr {
	// TODO: Mark thread as entering blocking syscall
	// Call the syscall
	// Check for cancellation

	return libKernel__close(fd)
}

// 0x00000000000026B0
// __int64 __fastcall close()
func libKernel__close(fd uintptr) uintptr {
	return libKernel_sys_close(fd)
}

func libKernel_sys_close(fd uintptr) uintptr {
	GlobalFilesystem.Lock.Lock()
	file, ok := GlobalFilesystem.Descriptors[FileDescriptor(fd)]
	GlobalFilesystem.Lock.Unlock()
	if !ok {
		logger.Printf("%-132s %s failed due to unknown file %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("_close"),
			color.Yellow.Sprintf("0x%X", fd),
		)
		SetErrno(ENOENT)
		return ERR_PTR
	}

	if err := GlobalFilesystem.Close(FileDescriptor(fd)); err != nil {
		logger.Printf("%-132s %s failed due to close error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("_close"),
			color.Blue.Sprint(file.Path),
			err.Error(),
		)
		SetErrno(EFAULT)
		return ERR_PTR
	}

	logger.Printf("%-132s %s closed file %s (path=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("_close"),
		color.Yellow.Sprintf("0x%X", file.Descriptor),
		color.Blue.Sprint(file.Path),
	)
	return 0
}
