package kernel

import (
	"io"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000015930
// __int64 __fastcall sceKernelRead(__int64, __int64, __int64)
func libKernel_sceKernelRead(fd FileDescriptor, bufPtr uintptr, length uint64) int64 {
	readBytes := libKernel_read(fd, bufPtr, length)
	if readBytes == ERR_PTRI {
		return int64(emu.GetErrno() - SonyErrorOffset)
	}

	return readBytes
}

// 0x000000000000E0A0
// __int64 __fastcall read(unsigned int, __int64, __int64)
func libKernel_read(fd FileDescriptor, bufPtr uintptr, length uint64) int64 {
	// TODO: Mark thread as entering blocking syscall
	// Call the syscall
	// Check for cancellation

	return libKernel__read(fd, bufPtr, length)
}

// 0x00000000000027D0
// __int64 __fastcall read(_QWORD, _QWORD, _QWORD)
func libKernel__read(fd FileDescriptor, bufPtr uintptr, length uint64) int64 {
	if bufPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid buffer pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("_read"),
		)
		emu.SetErrno(EFAULT)
		return 0
	}

	buffer := unsafe.Slice((*byte)(unsafe.Pointer(bufPtr)), length)
	readBytes, err := GlobalFilesystem.ReadFd(fd, buffer)
	if err != nil && err != io.EOF {
		logger.Printf("%-132s %s failed due to read error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("_read"),
			color.Yellow.Sprintf("0x%X", fd),
			err.Error(),
		)
		emu.SetErrno(EFAULT)
		return ERR_PTRI
	}

	if logger.LogFilesystem {
		logger.Printf("%-132s %s read %s bytes from %s (length=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("_read"),
			color.Yellow.Sprintf("0x%X", readBytes),
			color.Yellow.Sprintf("0x%X", fd),
			color.Yellow.Sprintf("0x%X", length),
		)
	}
	return int64(readBytes)
}

// 0x0000000000016520
// __int64 sceKernelPread()
func libKernel_sceKernelPread(fd FileDescriptor, bufPtr uintptr, length uint64, offset int64) int64 {
	readBytes := libKernel_pread(fd, bufPtr, length, offset)
	if readBytes == ERR_PTRI {
		return int64(emu.GetErrno() - SonyErrorOffset)
	}

	return readBytes
}

// 0x00000000000125B0
// __int64 pread()
func libKernel_pread(fd FileDescriptor, bufPtr uintptr, length uint64, offset int64) int64 {
	return libKernel_pread_0(fd, bufPtr, length, offset)
}

// 0x00000000000029B0
// __int64 pread_0()
func libKernel_pread_0(fd FileDescriptor, bufPtr uintptr, length uint64, offset int64) int64 {
	if bufPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid buffer pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pread_0"),
		)
		emu.SetErrno(EFAULT)
		return 0
	}

	buffer := unsafe.Slice((*byte)(unsafe.Pointer(bufPtr)), length)
	readBytes, err := GlobalFilesystem.PreadFd(fd, buffer, offset)
	if err != nil && err != io.EOF {
		logger.Printf("%-132s %s failed due to read error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pread_0"),
			color.Yellow.Sprintf("0x%X", fd),
			err.Error(),
		)

		if err.Error() == "invalid file descriptor" {
			emu.SetErrno(ENOENT)
		} else if err.Error() == "illegal seek" {
			emu.SetErrno(ESPIPE)
		} else {
			emu.SetErrno(EFAULT)
		}
		return ERR_PTRI
	}

	if logger.LogFilesystem {
		logger.Printf("%-132s %s read %s bytes from %s at offset %s (length=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pread_0"),
			color.Yellow.Sprintf("0x%X", readBytes),
			color.Yellow.Sprintf("0x%X", fd),
			color.Yellow.Sprintf("0x%X", offset),
			color.Yellow.Sprintf("0x%X", length),
		)
	}
	return int64(readBytes)
}
