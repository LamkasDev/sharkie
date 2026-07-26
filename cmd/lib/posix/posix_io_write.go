package posix

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func Write(fd FileDescriptor, bufPtr uintptr, length uint64) int64 {
	return libScePosix_write(fd, bufPtr, length)
}

func libScePosix_write(fd FileDescriptor, bufPtr uintptr, length uint64) int64 {
	if bufPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid buffer pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("write"),
		)
		emu.SetErrno(EFAULT)
		return ERR_PTRI
	}

	GlobalFilesystem.Lock.Lock()
	file, ok := GlobalFilesystem.Descriptors[fd]
	GlobalFilesystem.Lock.Unlock()
	if !ok {
		logger.Printf("%-132s %s failed due to unknown file %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("write"),
			color.Yellow.Sprintf("0x%X", fd),
		)
		emu.SetErrno(ENOENT)
		return ERR_PTRI
	}

	// Write data.
	buffer := unsafe.Slice((*byte)(unsafe.Pointer(bufPtr)), length)
	wroteBytes, err := file.Write(buffer)
	if err != nil {
		logger.Printf("%-132s %s failed due to write error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("write"),
			color.Yellow.Sprintf("0x%X", fd),
			err.Error(),
		)
		emu.SetErrno(EFAULT)
		return ERR_PTRI
	}

	if logger.LogFilesystem {
		logger.Printf("%-132s %s wrote %s bytes (fd=%s, length=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("write"),
			color.Yellow.Sprintf("0x%X", wroteBytes),
			color.Yellow.Sprintf("0x%X", fd),
			color.Yellow.Sprintf("0x%X", length),
		)
	}
	return int64(wroteBytes)
}

func Pwrite(fd FileDescriptor, bufPtr uintptr, length uint64, offset int64) int64 {
	return libScePosix_pwrite(fd, bufPtr, length, offset)
}

func libScePosix_pwrite(fd FileDescriptor, bufPtr uintptr, length uint64, offset int64) int64 {
	if bufPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid buffer pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pwrite"),
		)
		emu.SetErrno(EFAULT)
		return ERR_PTRI
	}

	buffer := unsafe.Slice((*byte)(unsafe.Pointer(bufPtr)), length)
	wroteBytes, err := GlobalFilesystem.PwriteFd(fd, buffer, offset)
	if err != nil {
		logger.Printf("%-132s %s failed due to write error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pwrite"),
			color.Yellow.Sprintf("0x%X", fd),
			err.Error(),
		)

		if err.Error() == "invalid file descriptor" {
			emu.SetErrno(ENOENT)
		} else if err.Error() == "illegal seek" {
			emu.SetErrno(ESPIPE)
		} else if err.Error() == "file not opened for writing" {
			emu.SetErrno(EFAULT) // TODO: EBADF?
		} else {
			emu.SetErrno(EFAULT)
		}
		return ERR_PTRI
	}

	if logger.LogFilesystem {
		logger.Printf("%-132s %s wrote %s bytes at offset %s (fd=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pwrite"),
			color.Yellow.Sprintf("0x%X", wroteBytes),
			color.Yellow.Sprintf("0x%X", offset),
			color.Yellow.Sprintf("0x%X", fd),
		)
	}
	return int64(wroteBytes)
}
