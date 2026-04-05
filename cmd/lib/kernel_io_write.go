package lib

import (
	"io"
	"strings"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs/fs"
	"github.com/gookit/color"
)

// 0x0000000000015960
// __int64 __fastcall sceKernelWrite(__int64, __int64, __int64)
func libKernel_sceKernelWrite(fd uintptr, bufPtr uintptr, length uintptr) uintptr {
	err := libKernel_write(fd, bufPtr, length)
	if err == ERR_PTR {
		return GetErrno() - SonyErrorOffset
	}

	return 0
}

// 0x000000000000E610
// __int64 __fastcall write()
func libKernel_write(fd uintptr, bufPtr uintptr, length uintptr) uintptr {
	// TODO: Mark thread as entering blocking syscall
	// Call the syscall
	// Check for cancellation

	return libKernel__write(fd, bufPtr, length)
}

// 0x0000000000002910
// __int64 __fastcall write()
func libKernel__write(fd uintptr, bufPtr uintptr, length uintptr) uintptr {
	if bufPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid buffer pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("_write"),
		)
		SetErrno(EFAULT)
		return 0
	}

	GlobalFilesystem.Lock.Lock()
	file, ok := GlobalFilesystem.Descriptors[FileDescriptor(fd)]
	GlobalFilesystem.Lock.Unlock()
	if !ok {
		logger.Printf("%-132s %s failed due to unknown file %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("_write"),
			color.Yellow.Sprintf("0x%X", fd),
		)
		SetErrno(ENOENT)
		return ERR_PTR
	}

	buffer := unsafe.Slice((*byte)(unsafe.Pointer(bufPtr)), length)
	if file.Path == "stdout" || file.Path == "stderr" || file.Path == "/dev/console" || file.Path == "/dev/deci_tty6" {
		message := string(buffer)
		outputColor, ok := FileDescriptorColors[file.Path]
		if !ok {
			outputColor = color.White
		}
		logger.Printf("%-132s %s %s",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprintf("[write on %s]", file.Path),
			outputColor.Sprint(message),
		)
		if !strings.HasSuffix(message, "\n") {
			logger.Println()
		}
		return length
	}

	// Write data.
	wroteBytes, err := GlobalFilesystem.WriteFd(FileDescriptor(fd), buffer)
	if err != nil {
		logger.Printf("%-132s %s failed due to write error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("_write"),
			color.Yellow.Sprintf("0x%X", fd),
			err.Error(),
		)
		SetErrno(EFAULT)
		return ERR_PTR
	}

	logger.Printf("%-132s %s wrote %s bytes to %s (length=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("_write"),
		color.Yellow.Sprintf("0x%X", wroteBytes),
		color.Yellow.Sprintf("0x%X", fd),
		color.Yellow.Sprintf("0x%X", length),
	)
	return uintptr(wroteBytes)
}

// 0x0000000000016550
// __int64 sceKernelPwrite()
func libKernel_sceKernelPwrite(fd uintptr, bufPtr uintptr, length uintptr, offset uintptr) uintptr {
	err := libKernel_pwrite(fd, bufPtr, length, offset)
	if err == ERR_PTR {
		return GetErrno() - SonyErrorOffset
	}

	return 0
}

// 0x00000000000125C0
// __int64 pwrite()
func libKernel_pwrite(fd uintptr, bufPtr uintptr, length uintptr, offset uintptr) uintptr {
	return libKernel_pwrite_0(fd, bufPtr, length, offset)
}

// 0x00000000000029D0
// __int64 pwrite_0()
func libKernel_pwrite_0(fd uintptr, bufPtr uintptr, length uintptr, offset uintptr) uintptr {
	if bufPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid buffer pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pwrite_0"),
		)
		SetErrno(EFAULT)
		return 0
	}

	GlobalFilesystem.Lock.Lock()
	file, ok := GlobalFilesystem.Descriptors[FileDescriptor(fd)]
	GlobalFilesystem.Lock.Unlock()
	if !ok {
		logger.Printf("%-132s %s failed due to unknown file %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pwrite_0"),
			color.Yellow.Sprintf("0x%X", fd),
		)
		SetErrno(ENOENT)
		return ERR_PTR
	}

	buffer := unsafe.Slice((*byte)(unsafe.Pointer(bufPtr)), length)
	currentOffset, _ := file.File.Seek(0, io.SeekCurrent)
	_, _ = file.File.Seek(int64(offset), io.SeekStart)
	wroteBytes, err := file.File.Write(buffer)
	_, _ = file.File.Seek(currentOffset, io.SeekStart)
	if err != nil {
		logger.Printf("%-132s %s failed due to write error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pwrite_0"),
			color.Yellow.Sprintf("0x%X", fd),
			err.Error(),
		)
		SetErrno(EFAULT)
		return ERR_PTR
	}

	logger.Printf("%-132s %s wrote %s bytes to %s at offset %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pwrite_0"),
		color.Yellow.Sprintf("0x%X", wroteBytes),
		color.Yellow.Sprintf("0x%X", fd),
		color.Yellow.Sprintf("0x%X", offset),
	)
	return uintptr(wroteBytes)
}
