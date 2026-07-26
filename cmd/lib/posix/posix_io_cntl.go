package posix

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func Fcntl(fd FileDescriptor, op FileOperation, arg uintptr) int64 {
	return libScePosix_fcntl(fd, op, arg)
}

// TODO: this
func libScePosix_fcntl(fd FileDescriptor, op FileOperation, arg uintptr) int64 {
	GlobalFilesystem.Lock.Lock()
	file, ok := GlobalFilesystem.Descriptors[fd]
	GlobalFilesystem.Lock.Unlock()
	if !ok {
		logger.Printf("%-132s %s failed due to unknown file %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fcntl"),
			color.Yellow.Sprintf("0x%X", fd),
		)
		emu.SetErrno(EBADF)
		return ERR_PTRI
	}

	if logger.LogFilesystem {
		logger.Printf("%-132s %s manipulated %s (op=%s, arg=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("fcntl"),
			color.Yellow.Sprintf("0x%X", file.Descriptor),
			color.Yellow.Sprintf("0x%X", op),
			color.Yellow.Sprintf("0x%X", arg),
		)
	}
	return 0
}
