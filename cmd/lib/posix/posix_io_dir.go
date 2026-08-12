package posix

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func Mkdir(pathPtr Cstring, mode uint16) int64 {
	return libScePosix_mkdir(pathPtr, mode)
}

func libScePosix_mkdir(pathPtr Cstring, mode uint16) int64 {
	if pathPtr == nil {
		logger.Printf("%-132s %s failed due to invalid path pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("mkdir"),
		)
		emu.SetErrno(EFAULT)
		return ERR_PTRI
	}
	path := GoString(pathPtr)
	err := GlobalFilesystem.Mkdir(path, mode)
	if err != nil {
		logger.Printf("%-132s %s failed due to mkdir error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("mkdir"),
			color.Yellow.Sprint(path),
			err.Error(),
		)
		emu.SetErrno(FsToPosixError(err))
		return ERR_PTRI
	}

	if logger.LogFilesystem {
		logger.Printf("%-132s %s created directory %s (mode=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("mkdir"),
			color.Yellow.Sprint(path),
			color.Green.Sprintf("0%o", mode),
		)
	}
	return 0
}

func Getdents(fd FileDescriptor, bufPtr uintptr, nbytes uint64) int64 {
	return libScePosix_getdents(fd, bufPtr, nbytes)
}

func libScePosix_getdents(fd FileDescriptor, bufPtr uintptr, nbytes uint64) int64 {
	return libScePosix_getdirentries(fd, bufPtr, nbytes, 0)
}

func Getdirentries(fd FileDescriptor, bufPtr uintptr, nbytes uint64, basep uintptr) int64 {
	return libScePosix_getdirentries(fd, bufPtr, nbytes, basep)
}

func libScePosix_getdirentries(fd FileDescriptor, bufPtr uintptr, nbytes uint64, basep uintptr) int64 {
	if bufPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid buffer pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("getdirentries"),
		)
		emu.SetErrno(EFAULT)
		return ERR_PTRI
	}
	if nbytes < 512 {
		logger.Printf("%-132s %s failed due to invalid size.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("getdirentries"),
		)
		emu.SetErrno(EINVAL)
		return ERR_PTRI
	}

	// If the guest provided a basep pointer, save the offset before reading.
	if basep != 0 {
		offset, err := GlobalFilesystem.GetOffsetFd(fd)
		if err == nil {
			*(*int64)(unsafe.Pointer(basep)) = offset
		}
	}

	// Get directory entries.
	data, err := GlobalFilesystem.GetdentsFd(fd, nbytes)
	if err != nil {
		logger.Printf("%-132s %s failed due to iterate error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("getdirentries"),
			color.Yellow.Sprintf("0x%X", fd),
			err.Error(),
		)
		emu.SetErrno(FsToPosixError(err))
		return ERR_PTRI
	}

	// Copy the resulting bytes.
	bytesWritten := int64(len(data))
	if bytesWritten > 0 {
		buffer := unsafe.Slice((*byte)(unsafe.Pointer(bufPtr)), bytesWritten)
		copy(buffer, data)
	}

	if logger.LogFilesystem {
		logger.Printf("%-132s %s read %s bytes (fd=%s, bufPtr=%s, nbytes=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("getdirentries"),
			color.Yellow.Sprintf("0x%X", bytesWritten),
			color.Yellow.Sprintf("0x%X", fd),
			color.Yellow.Sprintf("0x%X", bufPtr),
			color.Yellow.Sprintf("0x%X", nbytes),
		)
	}
	return bytesWritten
}
