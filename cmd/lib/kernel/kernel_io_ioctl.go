package kernel

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000000970
// __int64 __fastcall ioctl()
func libKernel_ioctl(fd FileDescriptor, request uint64, argPtr uintptr) int32 {
	err := GlobalFilesystem.IoctlFd(fd, request, argPtr)
	if err != nil {
		if err.Error() == "invalid file descriptor" {
			logger.Printf("%-132s %s failed due to unknown file %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("ioctl"),
				color.Yellow.Sprintf("0x%X", fd),
			)
			emu.SetErrno(ENOENT)
			return ENOENT
		}
		logger.Printf("%-132s %s command %s on fd %s with argument %s failed (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ioctl"),
			color.Yellow.Sprintf("0x%X", request),
			color.Yellow.Sprintf("0x%X", fd),
			color.Yellow.Sprintf("0x%X", argPtr),
			err.Error(),
		)

		if false {
			emu.SetErrno(EFAULT)
			return ERR_PTRI
		}
	}

	return 0
}
