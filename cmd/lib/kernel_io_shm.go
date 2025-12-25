package lib

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x0000000000001750
// __int64 __fastcall shm_open()
func libKernel_shm_open(pathPtr uintptr, oflag uintptr, mode uintptr) uintptr {
	if pathPtr == 0 {
		logger.Printf("%-120s %s failed due to invalid path pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("shm_open"),
		)
		return 0
	}
	GlobalFilesystem.Lock.Lock()
	defer GlobalFilesystem.Lock.Unlock()

	path := GetUsablePath(ReadCString(pathPtr))
	file, err := GlobalFilesystem.Open(path, oflag, mode)
	if err != nil {
		logger.Printf("%-120s %s failed due to open error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("shm_open"),
			color.Blue.Sprint(path),
			err.Error(),
		)
		SetErrno(EFAULT)
		return ERR_PTR
	}

	logger.Printf("%-120s %s opened file %s (path=%s, oflag=%s, mode=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("shm_open"),
		color.Yellow.Sprintf("0x%X", file.Descriptor),
		color.Blue.Sprint(path),
		color.Yellow.Sprintf("0x%X", oflag),
		color.Yellow.Sprintf("0x%X", mode),
	)
	return uintptr(file.Descriptor)
}
