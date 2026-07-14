package kernel

import (
	"io"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x00000000000165B0
// __int64 sceKernelLseek()
func libKernel_sceKernelLseek(fd FileDescriptor, offset int64, whence int32) int64 {
	newOffset := libKernel_lseek(fd, offset, whence)
	if newOffset == ERR_PTRI {
		return int64(emu.GetErrno() - SonyErrorOffset)
	}

	return newOffset
}

// 0x0000000000012590
// __int64 lseek(void)
func libKernel_lseek(fd FileDescriptor, offset int64, whence int32) int64 {
	return libKernel_lseek_0(fd, offset, whence)
}

// 0x0000000000002970
// __int64 __fastcall lseek_0()
func libKernel_lseek_0(fd FileDescriptor, offset int64, whence int32) int64 {
	var goWhence int
	switch whence {
	case 0:
		goWhence = io.SeekStart
	case 1:
		goWhence = io.SeekCurrent
	case 2:
		goWhence = io.SeekEnd
	default:
		emu.SetErrno(EINVAL)
		return ERR_PTRI
	}

	newOffset, err := GlobalFilesystem.SeekFd(fd, offset, goWhence)
	if err != nil {
		logger.Printf("%-132s %s failed due to seek error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("lseek_0"),
			color.Yellow.Sprintf("0x%X", fd),
			err.Error(),
		)

		if err.Error() == "invalid file descriptor" {
			emu.SetErrno(ENOENT)
		} else {
			emu.SetErrno(ESPIPE)
		}
		return ERR_PTRI
	}

	if logger.LogFilesystem {
		logger.Printf("%-132s %s moved %s cursor to %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("lseek_0"),
			color.Yellow.Sprintf("0x%X", fd),
			color.Yellow.Sprintf("0x%X", newOffset),
		)
	}
	return newOffset
}
