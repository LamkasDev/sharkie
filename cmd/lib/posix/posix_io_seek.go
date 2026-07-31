package posix

import (
	"io"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func Lseek(fd FileDescriptor, offset int64, whence int32) int64 {
	return libScePosix_lseek(fd, offset, whence)
}

func libScePosix_lseek(fd FileDescriptor, offset int64, whence int32) int64 {
	var goWhence int
	switch whence {
	case 0:
		goWhence = io.SeekStart
	case 1:
		goWhence = io.SeekCurrent
	case 2:
		goWhence = io.SeekEnd
	default:
		emu.SetErrno(EINVAL)
		return ERR_PTRI
	}

	newOffset, err := GlobalFilesystem.SeekFd(fd, offset, goWhence)
	if err != nil {
		logger.Printf("%-132s %s failed due to seek error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("lseek"),
			color.Yellow.Sprintf("0x%X", fd),
			err.Error(),
		)
		emu.SetErrno(FsToPosixError(err))
		return ERR_PTRI
	}

	if logger.LogFilesystem {
		logger.Printf("%-132s %s moved cursor to %s (fd=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("lseek"),
			color.Yellow.Sprintf("0x%X", newOffset),
			color.Yellow.Sprintf("0x%X", fd),
		)
	}
	return newOffset
}
