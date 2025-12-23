package lib

import (
	"fmt"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x0000000000001750
// __int64 __fastcall shm_open()
func libKernel_shm_open(pathPtr uintptr, oflag int32, mode int32) uintptr {
	if pathPtr == 0 {
		fmt.Printf("%-120s %s failed due to invalid path pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("shm_open"),
		)
		return 0
	}
	GlobalFilesystem.Lock.Lock()
	defer GlobalFilesystem.Lock.Unlock()

	path := ReadCString(pathPtr)
	file, err := GlobalFilesystem.Open(path, oflag, mode)
	if err != nil {
		fmt.Printf("%-120s %s failed due to unknown file %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("shm_open"),
			color.Blue.Sprint(path),
			err.Error(),
		)
		SetErrno(ENOENT)
		return ERR_PTR
	}

	fmt.Printf("%-120s %s opened file %s (path=%s, oflag=%s, mode=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("shm_open"),
		color.Yellow.Sprintf("0x%X", file.Descriptor),
		color.Blue.Sprint(path),
		color.Yellow.Sprintf("0x%X", oflag),
		color.Yellow.Sprintf("0x%X", mode),
	)
	return uintptr(file.Descriptor)
}
