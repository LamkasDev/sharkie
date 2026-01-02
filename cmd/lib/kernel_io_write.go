package lib

import (
	"strings"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x0000000000015960
// __int64 __fastcall sceKernelWrite(__int64, __int64, __int64)
func libKernel_sceKernelWrite(fd uintptr, bufPtr uintptr, length uintptr) uintptr {
	err := libKernel_write(fd, bufPtr, length)
	if err != 0 {
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
	return libKernel_sys_write(fd, bufPtr, length)
}

func libKernel_sys_write(fd uintptr, bufPtr uintptr, length uintptr) uintptr {
	if bufPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid buffer pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("_write"),
		)
		SetErrno(EFAULT)
		return 0
	}
	GlobalFilesystem.Lock.Lock()
	defer GlobalFilesystem.Lock.Unlock()

	file, ok := GlobalFilesystem.Descriptors[FileDescriptor(fd)]
	if !ok {
		logger.Printf("%-132s %s failed due to unknown file %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("_write"),
			color.Yellow.Sprintf("0x%X", fd),
		)
		SetErrno(ENOENT)
		return ERR_PTR
	}
	fileData, err := GlobalFilesystem.ReadFull(file.Path)
	if err != nil {
		logger.Printf("%-132s %s failed due to read error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("_write"),
			color.Blue.Sprint(file.Path),
			err.Error(),
		)
		SetErrno(EFAULT)
		return ERR_PTR
	}

	buffer := unsafe.Slice((*byte)(unsafe.Pointer(bufPtr)), length)
	wroteBytes := uintptr(len(buffer))
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
		return wroteBytes
	}

	// Write data.
	fileData = append(fileData, buffer...)
	if _, err = GlobalFilesystem.Write(file.Path, fileData); err != nil {
		logger.Printf("%-132s %s failed due to write error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("_write"),
			color.Blue.Sprint(file.Path),
			err.Error(),
		)
		SetErrno(EFAULT)
		return ERR_PTR
	}
	file.Cursor += wroteBytes

	logger.Printf("%-132s %s wrote %s bytes to file %s (length=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("_write"),
		color.Yellow.Sprintf("0x%X", wroteBytes),
		color.Blue.Sprint(file.Path),
		color.Yellow.Sprintf("0x%X", length),
	)
	return wroteBytes
}

// 0x0000000000012580
// __int64 ftruncate()
func libKernel_ftruncate(fd uintptr, length uintptr) uintptr {
	return libKernel_ftruncate_0(fd, length)
}

// 0x0000000000002950
// __int64 ftruncate_0()
func libKernel_ftruncate_0(fd uintptr, length uintptr) uintptr {
	file, ok := GlobalFilesystem.Descriptors[FileDescriptor(fd)]
	if !ok {
		logger.Printf("%-132s %s failed due to unknown file %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ftruncate_0"),
			color.Yellow.Sprintf("0x%X", fd),
		)
		SetErrno(ENOENT)
		return ERR_PTR
	}
	GlobalFilesystem.Lock.Lock()
	defer GlobalFilesystem.Lock.Unlock()

	fileData, err := GlobalFilesystem.ReadFull(file.Path)
	if err != nil {
		logger.Printf("%-132s %s failed due to read error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ftruncate_0"),
			color.Blue.Sprint(file.Path),
			err.Error(),
		)
		SetErrno(EFAULT)
		return ERR_PTR
	}
	if length == uintptr(len(fileData)) {
		return 0
	}

	var fileChunk []byte
	if length < uintptr(len(fileData)) {
		fileChunk = fileData[:length]
	} else {
		padding := make([]byte, length-uintptr(len(fileData)))
		fileChunk = append(fileData, padding...)
	}

	_, err = GlobalFilesystem.Write(file.Path, fileChunk)
	if err != nil {
		logger.Printf("%-132s %s failed due to write error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ftruncate_0"),
			color.Blue.Sprint(file.Path),
			err.Error(),
		)
		SetErrno(EFAULT)
		return ERR_PTR
	}

	logger.Printf("%-132s %s truncated file %s from %s to %s bytes.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("ftruncate_0"),
		color.Blue.Sprint(file.Path),
		color.Yellow.Sprintf("0x%X", len(fileData)),
		color.Yellow.Sprintf("0x%X", length),
	)
	return 0
}

// 0x0000000000016550
// __int64 sceKernelPwrite()
func libKernel_sceKernelPwrite(fd uintptr, bufPtr uintptr, length uintptr, offset uintptr) uintptr {
	err := libKernel_pwrite(fd, bufPtr, length, offset)
	if err != 0 {
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
	return libKernel_sys_pwrite(fd, bufPtr, length, offset)
}

func libKernel_sys_pwrite(fd uintptr, bufPtr uintptr, length uintptr, offset uintptr) uintptr {
	if bufPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid buffer pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pwrite_0"),
		)
		SetErrno(EFAULT)
		return 0
	}
	GlobalFilesystem.Lock.Lock()
	defer GlobalFilesystem.Lock.Unlock()

	file, ok := GlobalFilesystem.Descriptors[FileDescriptor(fd)]
	if !ok {
		logger.Printf("%-132s %s failed due to unknown file %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pwrite_0"),
			color.Yellow.Sprintf("0x%X", fd),
		)
		SetErrno(ENOENT)
		return ERR_PTR
	}
	fileData, err := GlobalFilesystem.ReadFull(file.Path)
	if err != nil {
		logger.Printf("%-132s %s failed due to read error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pwrite_0"),
			color.Blue.Sprint(file.Path),
			err.Error(),
		)
		SetErrno(EFAULT)
		return ERR_PTR
	}

	buffer := unsafe.Slice((*byte)(unsafe.Pointer(bufPtr)), length)
	wroteBytes := uintptr(len(buffer))

	// Ensure the file is large enough for a write at the offset.
	requiredSize := offset + wroteBytes
	currentSize := uintptr(len(fileData))
	if requiredSize > currentSize {
		// Expand the file with zeros if we are writing past EOF.
		expansion := make([]byte, requiredSize-currentSize)
		fileData = append(fileData, expansion...)
	}

	// Write data.
	copy(fileData[offset:], buffer)
	if _, err = GlobalFilesystem.Write(file.Path, fileData); err != nil {
		logger.Printf("%-132s %s failed due to write error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pwrite_0"),
			color.Blue.Sprint(file.Path),
			err.Error(),
		)
		SetErrno(EFAULT)
		return ERR_PTR
	}

	logger.Printf("%-132s %s wrote %s bytes to file %s at offset %s (length=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pwrite_0"),
		color.Yellow.Sprintf("0x%X", wroteBytes),
		color.Blue.Sprint(file.Path),
		color.Yellow.Sprintf("0x%X", offset),
		color.Yellow.Sprintf("0x%X", length),
	)
	return wroteBytes
}
