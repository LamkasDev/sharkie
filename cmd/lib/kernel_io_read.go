package lib

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x0000000000015930
// __int64 __fastcall sceKernelRead(__int64, __int64, __int64)
func libKernel_sceKernelRead(fd uintptr, bufPtr uintptr, length uintptr) uintptr {
	err := libKernel_read(fd, bufPtr, length)
	if err != 0 {
		return GetErrno() - 0x7FFE0000
	}

	return 0
}

// 0x000000000000E0A0
// __int64 __fastcall read(unsigned int, __int64, __int64)
func libKernel_read(fd uintptr, bufPtr uintptr, length uintptr) uintptr {
	// TODO: Mark thread as entering blocking syscall
	// Call the syscall
	// Check for cancellation

	return libKernel__read(fd, bufPtr, length)
}

// 0x00000000000027D0
// __int64 __fastcall read(_QWORD, _QWORD, _QWORD)
func libKernel__read(fd uintptr, bufPtr uintptr, length uintptr) uintptr {
	return libKernel_sys_read(fd, bufPtr, length)
}

func libKernel_sys_read(fd uintptr, bufPtr uintptr, length uintptr) uintptr {
	if bufPtr == 0 {
		logger.Printf("%-120s %s failed due to invalid buffer pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("_read"),
		)
		SetErrno(EFAULT)
		return 0
	}
	GlobalFilesystem.Lock.Lock()
	defer GlobalFilesystem.Lock.Unlock()

	file, ok := GlobalFilesystem.Descriptors[FileDescriptor(fd)]
	if !ok {
		logger.Printf("%-120s %s failed due to unknown file %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("_read"),
			color.Yellow.Sprintf("0x%X", fd),
		)
		SetErrno(ENOENT)
		return ERR_PTR
	}
	fileData, err := GlobalFilesystem.ReadFull(file.Path)
	if err != nil {
		logger.Printf("%-120s %s failed due to read error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("_read"),
			color.Blue.Sprint(file.Path),
			err.Error(),
		)
		SetErrno(EFAULT)
		return ERR_PTR
	}
	if file.Cursor >= uintptr(len(fileData)) {
		logger.Printf("%-120s %s ignored read of %s bytes from file %s (cursor EOF).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("_read"),
			color.Yellow.Sprintf("0x%X", length),
			color.Blue.Sprint(file.Path),
		)
		return 0
	}

	availableBytes := uintptr(len(fileData)) - file.Cursor
	readBytes := length
	if readBytes > availableBytes {
		readBytes = availableBytes
	}
	buffer := unsafe.Slice((*byte)(unsafe.Pointer(bufPtr)), readBytes)
	copy(buffer, fileData[file.Cursor:file.Cursor+readBytes])
	file.Cursor += readBytes

	logger.Printf("%-120s %s read %s bytes from file %s (length=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("_read"),
		color.Yellow.Sprintf("0x%X", readBytes),
		color.Blue.Sprint(file.Path),
		color.Yellow.Sprintf("0x%X", length),
	)
	return readBytes
}
