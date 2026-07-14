package kernel

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000015960
// __int64 __fastcall sceKernelWrite(__int64, __int64, __int64)
func libKernel_sceKernelWrite(fd FileDescriptor, bufPtr uintptr, length uint64) int64 {
	wroteBytes := libKernel_write(fd, bufPtr, length)
	if wroteBytes == ERR_PTRI {
		return int64(emu.GetErrno() - SonyErrorOffset)
	}

	return wroteBytes
}

// 0x000000000000E610
// __int64 __fastcall write()
func libKernel_write(fd FileDescriptor, bufPtr uintptr, length uint64) int64 {
	// TODO: Mark thread as entering blocking syscall
	// Call the syscall
	// Check for cancellation

	return libKernel__write(fd, bufPtr, length)
}

// 0x0000000000002910
// __int64 __fastcall write()
func libKernel__write(fd FileDescriptor, bufPtr uintptr, length uint64) int64 {
	if bufPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid buffer pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("_write"),
		)
		emu.SetErrno(EFAULT)
		return 0
	}

	GlobalFilesystem.Lock.Lock()
	file, ok := GlobalFilesystem.Descriptors[fd]
	GlobalFilesystem.Lock.Unlock()
	if !ok {
		logger.Printf("%-132s %s failed due to unknown file %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("_write"),
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
			color.Magenta.Sprint("_write"),
			color.Yellow.Sprintf("0x%X", fd),
			err.Error(),
		)
		emu.SetErrno(EFAULT)
		return ERR_PTRI
	}

	if logger.LogFilesystem {
		logger.Printf("%-132s %s wrote %s bytes to %s (length=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("_write"),
			color.Yellow.Sprintf("0x%X", wroteBytes),
			color.Yellow.Sprintf("0x%X", fd),
			color.Yellow.Sprintf("0x%X", length),
		)
	}
	return int64(wroteBytes)
}

// 0x0000000000016550
// __int64 sceKernelPwrite()
func libKernel_sceKernelPwrite(fd FileDescriptor, bufPtr uintptr, length uint64, offset int64) int64 {
	wroteBytes := libKernel_pwrite(fd, bufPtr, length, offset)
	if wroteBytes == ERR_PTRI {
		return int64(emu.GetErrno() - SonyErrorOffset)
	}

	return wroteBytes
}

// 0x00000000000125C0
// __int64 pwrite()
func libKernel_pwrite(fd FileDescriptor, bufPtr uintptr, length uint64, offset int64) int64 {
	return libKernel_pwrite_0(fd, bufPtr, length, offset)
}

// 0x00000000000029D0
// __int64 pwrite_0()
func libKernel_pwrite_0(fd FileDescriptor, bufPtr uintptr, length uint64, offset int64) int64 {
	if bufPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid buffer pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pwrite_0"),
		)
		emu.SetErrno(EFAULT)
		return 0
	}

	buffer := unsafe.Slice((*byte)(unsafe.Pointer(bufPtr)), length)
	wroteBytes, err := GlobalFilesystem.PwriteFd(fd, buffer, offset)
	if err != nil {
		logger.Printf("%-132s %s failed due to write error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pwrite_0"),
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
		logger.Printf("%-132s %s wrote %s bytes to %s at offset %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pwrite_0"),
			color.Yellow.Sprintf("0x%X", wroteBytes),
			color.Yellow.Sprintf("0x%X", fd),
			color.Yellow.Sprintf("0x%X", offset),
		)
	}
	return int64(wroteBytes)
}
