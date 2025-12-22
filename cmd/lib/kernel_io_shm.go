package lib

import (
	"fmt"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x0000000000001750
// __int64 __fastcall shm_open()
func libKernel_shm_open(namePtr uintptr, oflag int32, mode int32) uintptr {
	if namePtr == 0 {
		fmt.Printf("%-120s %s failed due to invalid name pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("shm_open"),
		)
		return 0
	}
	GlobalShmFilesystemLock.Lock()
	defer GlobalShmFilesystemLock.Unlock()

	name := ReadCString(namePtr)
	file, err := GlobalShmFilesystem.Open(name, oflag, mode)
	if err != nil {
		fmt.Printf("%-120s %s failed to open file %s: %+v\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("shm_open"),
			color.Blue.Sprint(name),
			err.Error(),
		)
		return ERR_PTR
	}

	fmt.Printf("%-120s %s opened file %s (name=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("shm_open"),
		color.Yellow.Sprintf("0x%X", file.Descriptor),
		color.Blue.Sprint(name),
	)
	return uintptr(file.Descriptor)
}
