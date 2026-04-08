package lib

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs/fs"
	"github.com/gookit/color"
)

// 0x0000000000001750
// __int64 __fastcall shm_open()
func libKernel_shm_open(pathPtr Cstring, oflag FileFlags, mode FileMode) int32 {
	if pathPtr == nil {
		logger.Printf("%-132s %s failed due to invalid path pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("shm_open"),
		)
		return 0
	}

	path := GetUsablePath(GoString(pathPtr))
	fd, err := GlobalFilesystem.Open(path, oflag, mode)
	if err != nil {
		logger.Printf("%-132s %s failed due to open error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("shm_open"),
			color.Blue.Sprint(path),
			err.Error(),
		)
		SetErrno(EFAULT)
		return ERR_PTRI
	}

	if logger.LogFilesystem {
		logger.Printf("%-132s %s opened file %s (path=%s, oflag=%s, mode=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("shm_open"),
			color.Yellow.Sprintf("0x%X", fd),
			color.Blue.Sprint(path),
			color.Yellow.Sprintf("0x%X", oflag),
			color.Yellow.Sprintf("0x%X", mode),
		)
	}
	return int32(fd)
}
