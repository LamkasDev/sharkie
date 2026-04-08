package lib

import (
	"io"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs/fs"
	"github.com/gookit/color"
)

// 0x00000000000165B0
// __int64 sceKernelLseek()
func libKernel_sceKernelLseek(fd FileDescriptor, offset int64, whence int32) int64 {
	newOffset := libKernel_lseek(fd, offset, whence)
	if newOffset == ERR_PTRI {
		return int64(GetErrno() - SonyErrorOffset)
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
	GlobalFilesystem.Lock.Lock()
	file, ok := GlobalFilesystem.Descriptors[fd]
	GlobalFilesystem.Lock.Unlock()
	if !ok {
		logger.Printf("%-132s %s failed due to unknown file %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pread_0"),
			color.Yellow.Sprintf("0x%X", fd),
		)
		SetErrno(ENOENT)
		return ERR_PTRI
	}

	var goWhence int
	switch whence {
	case 0:
		goWhence = io.SeekStart
	case 1:
		goWhence = io.SeekCurrent
	case 2:
		goWhence = io.SeekEnd
	default:
		SetErrno(EINVAL)
		return ERR_PTRI
	}

	newOffset, err := file.File.Seek(offset, goWhence)
	if err != nil {
		logger.Printf("%-132s %s failed due to seek error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("lseek_0"),
			color.Yellow.Sprintf("0x%X", fd),
			err.Error(),
		)
		SetErrno(ESPIPE)
		return -1
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
