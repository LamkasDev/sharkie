package posix

import (
	"io"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
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
		emu.SetErrno(FsToPosixError(err))
		return ERR_PTRI
	}

	if length > 1 && logger.LogFilesystem {
		logger.Printf("%-132s %s read %s bytes (fd=%s, bufPtr=%s, length=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("read"),
			color.Yellow.Sprintf("0x%X", readBytes),
			color.Yellow.Sprintf("0x%X", fd),
			color.Yellow.Sprintf("0x%X", bufPtr),
			color.Yellow.Sprintf("0x%X", length),
		)
	}
	return int64(readBytes)
}

func Pread(fd FileDescriptor, bufPtr uintptr, length uint64, offset int64) int64 {
	return libScePosix_pread(fd, bufPtr, length, offset)
}

func libScePosix_pread(fd FileDescriptor, bufPtr uintptr, length uint64, offset int64) int64 {
	if bufPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid buffer pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pread"),
		)
		emu.SetErrno(EFAULT)
		return ERR_PTRI
	}

	buffer := unsafe.Slice((*byte)(unsafe.Pointer(bufPtr)), length)
	readBytes, err := GlobalFilesystem.PreadFd(fd, buffer, offset)
	if err != nil && err != io.EOF {
		logger.Printf("%-132s %s failed due to read error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pread"),
			color.Yellow.Sprintf("0x%X", fd),
			err.Error(),
		)
		emu.SetErrno(FsToPosixError(err))
		return ERR_PTRI
	}

	if length > 1 && logger.LogFilesystem {
		logger.Printf("%-132s %s read %s bytes at offset %s (fd=%s, bufPtr=%s, length=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pread"),
			color.Yellow.Sprintf("0x%X", readBytes),
			color.Yellow.Sprintf("0x%X", offset),
			color.Yellow.Sprintf("0x%X", fd),
			color.Yellow.Sprintf("0x%X", bufPtr),
			color.Yellow.Sprintf("0x%X", length),
		)
	}
	return int64(readBytes)
}
