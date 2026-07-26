package posix

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func Close(fd FileDescriptor) int32 {
	return libScePosix_close(fd)
}

func libScePosix_close(fd FileDescriptor) int32 {
	GlobalFilesystem.Lock.Lock()
	file, ok := GlobalFilesystem.Descriptors[fd]
	GlobalFilesystem.Lock.Unlock()
	if !ok {
		logger.Printf("%-132s %s failed due to unknown file %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("close"),
			color.Yellow.Sprintf("0x%X", fd),
		)
		emu.SetErrno(ENOENT)
		return ERR_PTRI
	}

	if err := GlobalFilesystem.Close(fd); err != nil {
		logger.Printf("%-132s %s failed due to close error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("close"),
			color.Blue.Sprint(file.Path),
			err.Error(),
		)
		emu.SetErrno(EFAULT)
		return ERR_PTRI
	}

	if logger.LogFilesystem {
		logger.Printf("%-132s %s closed file %s (path=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("close"),
			color.Yellow.Sprintf("0x%X", file.Descriptor),
			color.Blue.Sprint(file.Path),
		)
	}
	return 0
}
