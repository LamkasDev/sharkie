package posix

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func Truncate(pathPtr Cstring, length int64) int32 {
	return libScePosix_truncate(pathPtr, length)
}

func libScePosix_truncate(pathPtr Cstring, length int64) int32 {
	if pathPtr == nil {
		logger.Printf("%-132s %s failed due to invalid path pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("truncate"),
		)
		emu.SetErrno(ENOENT)
		return ERR_PTRI
	}

	path := GlobalFilesystem.GetUsablePath(GoString(pathPtr))
	err := GlobalFilesystem.Truncate(path, length)
	if err != nil {
		logger.Printf("%-132s %s failed due to truncate error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("truncate"),
			color.Blue.Sprint(path),
			err.Error(),
		)
		emu.SetErrno(FsToPosixError(err))
		return ERR_PTRI
	}

	if logger.LogFilesystem {
		logger.Printf("%-132s %s truncated %s to %s bytes.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("truncate"),
			color.Blue.Sprint(path),
			color.Yellow.Sprintf("0x%X", length),
		)
	}
	return 0
}

func Ftruncate(fd FileDescriptor, length int64) int32 {
	return libScePosix_ftruncate(fd, length)
}

func libScePosix_ftruncate(fd FileDescriptor, length int64) int32 {
	err := GlobalFilesystem.TruncateFd(fd, length)
	if err != nil {
		logger.Printf("%-132s %s failed due to truncate error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ftruncate"),
			color.Yellow.Sprintf("0x%X", fd),
			err.Error(),
		)
		emu.SetErrno(FsToPosixError(err))
		return ERR_PTRI
	}

	if logger.LogFilesystem {
		logger.Printf("%-132s %s truncated %s to %s bytes.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ftruncate"),
			color.Yellow.Sprintf("0x%X", fd),
			color.Yellow.Sprintf("0x%X", length),
		)
	}
	return 0
}
