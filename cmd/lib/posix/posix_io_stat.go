package posix

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func Stat(pathPtr Cstring, stat *FileStat) int32 {
	return libScePosix_stat(pathPtr, stat)
}

func libScePosix_stat(pathPtr Cstring, stat *FileStat) int32 {
	if pathPtr == nil {
		logger.Printf("%-132s %s failed due to invalid path pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("stat"),
		)
		emu.SetErrno(EFAULT)
		return ERR_PTRI
	}

	path := GlobalFilesystem.GetUsablePath(GoString(pathPtr))
	fileStat, err := GlobalFilesystem.Stat(path)
	if err != nil {
		logger.Printf("%-132s %s failed due to stat error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("stat"),
			color.Blue.Sprint(path),
			err.Error(),
		)
		emu.SetErrno(FsToPosixError(err))
		return ERR_PTRI
	}
	*stat = *fileStat

	if logger.LogFilesystem {
		logger.Printf("%-132s %s returned file stat for %s (size=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("stat"),
			color.Blue.Sprint(path),
			color.Yellow.Sprintf("0x%X", stat.Size),
		)
	}
	return 0
}

func Fstat(fd FileDescriptor, stat *FileStat) int32 {
	return libScePosix_fstat(fd, stat)
}

func libScePosix_fstat(fd FileDescriptor, stat *FileStat) int32 {
	if stat == nil {
		logger.Printf("%-132s %s failed due to invalid stat pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fstat"),
		)
		emu.SetErrno(EFAULT)
		return ERR_PTRI
	}

	fileStat, err := GlobalFilesystem.StatFd(fd)
	if err != nil {
		logger.Printf("%-132s %s failed due to stat error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fstat"),
			color.Yellow.Sprintf("0x%X", fd),
			err.Error(),
		)
		emu.SetErrno(FsToPosixError(err))
		return ERR_PTRI
	}
	*stat = *fileStat

	if logger.LogFilesystem {
		logger.Printf("%-132s %s returned file stat for %s (size=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fstat"),
			color.Yellow.Sprintf("0x%X", fd),
			color.Yellow.Sprintf("0x%X", stat.Size),
		)
	}
	return 0
}
