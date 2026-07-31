package posix

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func Ioctl(fd FileDescriptor, request uint64, argPtr uintptr) int32 {
	return libScePosix_ioctl(fd, request, argPtr)
}

func libScePosix_ioctl(fd FileDescriptor, request uint64, argPtr uintptr) int32 {
	err := GlobalFilesystem.IoctlFd(fd, request, argPtr)
	if err != nil {
		if err.Error() == "invalid file descriptor" {
			logger.Printf("%-132s %s failed due to unknown file descriptor %s.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("ioctl"),
				color.Yellow.Sprintf("0x%X", fd),
			)
			emu.SetErrno(EBADF)
			return ERR_PTRI
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
