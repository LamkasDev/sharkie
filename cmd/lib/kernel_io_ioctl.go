package lib

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs/fs"
	"github.com/gookit/color"
)

// 0x0000000000000970
// __int64 __fastcall ioctl()
func libKernel_ioctl(fd, request, mode uintptr) uintptr {
	return libKernel_sys_ioctl(fd, request, mode)
}

func libKernel_sys_ioctl(fd, request, argPtr uintptr) uintptr {
	file, ok := GlobalFilesystem.Descriptors[FileDescriptor(fd)]
	if !ok {
		logger.Printf("%-132s %s failed due to unknown file %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ioctl"),
			color.Yellow.Sprintf("0x%X", fd),
		)
		SetErrno(ENOENT)
		return ENOENT
	}

	err := file.File.Ioctl(uint32(request), argPtr)
	if err != nil {
		logger.Printf("%-132s %s command %s on %s with argument %s failed (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ioctl"),
			color.Yellow.Sprintf("0x%X", request),
			color.Yellow.Sprintf("0x%X", argPtr),
			color.Blue.Sprint(file.Path),
			err.Error(),
		)
		if false {
			SetErrno(EFAULT)
			return ERR_PTR
		}
	}

	return 0
}
