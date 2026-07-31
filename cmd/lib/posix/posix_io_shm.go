package posix

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func Shm_open(pathPtr Cstring, flags FileFlags, mode FileMode) int32 {
	return libScePosix_shm_open(pathPtr, flags, mode)
}

func libScePosix_shm_open(pathPtr Cstring, flags FileFlags, mode FileMode) int32 {
	if pathPtr == nil {
		logger.Printf("%-132s %s failed due to invalid path pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("shm_open"),
		)
		return 0
	}

	path := GlobalFilesystem.GetUsablePath(GoString(pathPtr))
	fd, err := GlobalFilesystem.Open(path, flags, mode)
	if err != nil {
		logger.Printf("%-132s %s failed due to open error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("shm_open"),
			color.Blue.Sprint(path),
			err.Error(),
		)
		emu.SetErrno(FsToPosixError(err))
		return ERR_PTRI
	}

	if logger.LogFilesystem {
		logger.Printf("%-132s %s opened file %s (path=%s, oflag=%s, mode=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("shm_open"),
			color.Yellow.Sprintf("0x%X", fd),
			color.Blue.Sprint(path),
			color.Yellow.Sprintf("0x%X", flags),
			color.Yellow.Sprintf("0x%X", mode),
		)
	}
	return int32(fd)
}
