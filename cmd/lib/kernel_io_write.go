package lib

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x0000000000015960
// __int64 __fastcall sceKernelWrite(__int64, __int64, __int64)
func libKernel_sceKernelWrite(fd uintptr, bufPtr uintptr, length uintptr) uintptr {
	err := libKernel_write(fd, bufPtr, length)
	if err != 0 {
		return GetErrno() - 0x7FFE0000
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
		fmt.Printf("%-120s %s failed due to invalid buffer pointer.\n",
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
		fmt.Printf("%-120s %s failed due to unknown file %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("_write"),
			color.Yellow.Sprintf("0x%X", fd),
		)
		SetErrno(ENOENT)
		return ERR_PTR
	}
	fileData, err := GlobalFilesystem.ReadFull(file.Path)
	if err != nil {
		fmt.Printf("%-120s %s failed due to read error on %s (%s).\n",
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
		fmt.Printf("%-120s %s %s",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprintf("[write on %s]", file.Path),
			outputColor.Sprint(message),
		)
		if !strings.HasSuffix(message, "\n") {
			fmt.Println("")
		}
		return wroteBytes
	}
	fileData = append(fileData, buffer...)
	if _, err = GlobalFilesystem.Write(file.Path, fileData); err != nil {
		fmt.Printf("%-120s %s failed due to write error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("_write"),
			color.Blue.Sprint(file.Path),
			err.Error(),
		)
		SetErrno(EFAULT)
		return ERR_PTR
	}
	file.Cursor += wroteBytes

	fmt.Printf("%-120s %s wrote %s bytes to file %s (length=%s).\n",
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
		fmt.Printf("%-120s %s failed due to unknown file %s.\n",
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
		fmt.Printf("%-120s %s failed due to read error on %s (%s).\n",
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
		fmt.Printf("%-120s %s failed due to write error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ftruncate_0"),
			color.Blue.Sprint(file.Path),
			err.Error(),
		)
		SetErrno(EFAULT)
		return ERR_PTR
	}

	fmt.Printf("%-120s %s truncated file %s from %s to %s bytes.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("ftruncate_0"),
		color.Blue.Sprint(file.Path),
		color.Yellow.Sprintf("0x%X", len(fileData)),
		color.Yellow.Sprintf("0x%X", length),
	)
	return 0
}
