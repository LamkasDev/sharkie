package posix

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func Stat(pathPtr Cstring, statPtr uintptr) int32 {
	return libScePosix_stat(pathPtr, statPtr)
}

func libScePosix_stat(pathPtr Cstring, statPtr uintptr) int32 {
	if pathPtr == nil {
		logger.Printf("%-132s %s failed due to invalid path pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("stat"),
		)
		return 0
	}

	path := GetUsablePath(GoString(pathPtr))
	fileStat, err := GlobalFilesystem.Stat(path)
	if err != nil {
		logger.Printf("%-132s %s failed due to stat error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("stat"),
			color.Blue.Sprint(path),
			err.Error(),
		)
		emu.SetErrno(ENOENT)
		return ERR_PTRI
	}
	stat := (*FileStat)(unsafe.Pointer(statPtr))
	*stat = *fileStat

	return 0
}
