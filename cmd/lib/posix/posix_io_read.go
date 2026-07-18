package posix

import (
	"io"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func Read(fd FileDescriptor, bufPtr uintptr, length uint64) int64 {
	return libScePosix_read(fd, bufPtr, length)
}

func libScePosix_read(fd FileDescriptor, bufPtr uintptr, length uint64) int64 {
	if bufPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid buffer pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("read"),
		)
		emu.SetErrno(EFAULT)
		return ERR_PTRI
	}

	buffer := unsafe.Slice((*byte)(unsafe.Pointer(bufPtr)), length)
	readBytes, err := GlobalFilesystem.ReadFd(fd, buffer)
	if err != nil && err != io.EOF {
		logger.Printf("%-132s %s failed due to read error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("read"),
			color.Yellow.Sprintf("0x%X", fd),
			err.Error(),
		)
		emu.SetErrno(EFAULT)
		return ERR_PTRI
	}

	if logger.LogFilesystem {
		logger.Printf("%-132s %s read %s bytes from %s (length=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("read"),
			color.Yellow.Sprintf("0x%X", readBytes),
			color.Yellow.Sprintf("0x%X", fd),
			color.Yellow.Sprintf("0x%X", length),
		)
	}
	return int64(readBytes)
}
