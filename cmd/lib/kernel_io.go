package lib

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/gookit/color"
)

const (
	// Standard POSIX
	FD_STDIN  = uintptr(0)
	FD_STDOUT = uintptr(1)
	FD_STDERR = uintptr(2)

	// PS4 Specific
	FD_CONSOLE = uintptr(10)
	FD_TTY     = uintptr(11)
	FD_GC      = uintptr(20)
)

var FileDescriptors = map[string]uintptr{
	"stdin":          FD_STDIN,
	"stdout":         FD_STDOUT,
	"stderr":         FD_STDERR,
	"/dev/console":   FD_CONSOLE,
	"/dev/deci_tty6": FD_TTY,
	"/dev/gc":        FD_GC,
}

var FileDescriptorNames = map[uintptr]string{
	FD_STDIN:   "stdin",
	FD_STDOUT:  "stdout",
	FD_STDERR:  "stderr",
	FD_CONSOLE: "/dev/console",
	FD_TTY:     "/dev/deci_tty6",
	FD_GC:      "/dev/gc",
}

var FileDescriptorColors = map[uintptr]color.Color{
	FD_STDIN:   color.Gray,
	FD_STDOUT:  color.Cyan,
	FD_STDERR:  color.Red,
	FD_CONSOLE: color.Cyan,
	FD_TTY:     color.Cyan,
	FD_GC:      color.Cyan,
}

// TODO: not sure correct errors
// 0x0000000000002750
// __int64 __fastcall open()
func libKernel__open(pathPtr uintptr, flags uintptr, mode uintptr) uintptr {
	if pathPtr == 0 {
		fmt.Printf("%-120s %s failed due to invalid path pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("_open"),
		)
		SetErrno(EFAULT)
		return ^uintptr(0)
	}

	path := ReadCString(pathPtr)
	fmt.Printf("%-120s %s opened file %s (flags=%s, mode=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("_open"),
		color.Blue.Sprint(path),
		color.Yellow.Sprintf("0x%X", flags),
		color.Yellow.Sprintf("0x%X", mode),
	)

	fd, ok := FileDescriptors[path]
	if !ok {
		SetErrno(ENOENT)
		return ^uintptr(0)
	}

	return fd
}

// 0x000000000000DD50
// __int64 __fastcall open(__m128 _XMM0, __m128 _XMM1, __m128 _XMM2, __m128 _XMM3, __m128 _XMM4, __m128 _XMM5, __m128 _XMM6, __m128 _XMM7, __int64, __int16, __int64, __int64, __int64, __int64, char)
func libKernel_open(pathPtr uintptr, flags uintptr, mode uintptr) uintptr {
	// TODO: Mark thread as entering blocking syscall
	// Call the syscall
	// Check for cancellation

	return libKernel__open(pathPtr, flags, mode)
}

// 0x0000000000002910
// __int64 __fastcall write()
func libKernel__write(fd uintptr, bufPtr uintptr, length uintptr) uintptr {
	if bufPtr == 0 {
		fmt.Printf("%-120s %s failed due to invalid buffer pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("_open"),
		)
		SetErrno(EFAULT)
		return ^uintptr(0)
	}

	buffSlice := unsafe.Slice((*byte)(unsafe.Pointer(bufPtr)), length)
	message := string(buffSlice)
	name, ok := FileDescriptorNames[fd]
	if !ok {
		fmt.Printf("%-120s %s %s",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("[_write on unknown fd]"),
			message,
		)
		if !strings.HasSuffix(message, "\n") {
			fmt.Println("")
		}
		return 0
	}

	fmt.Printf("%-120s %s %s",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("[_write on %s]", name),
		message,
	)
	if !strings.HasSuffix(message, "\n") {
		fmt.Println("")
	}
	return length
}

// 0x0000000000000970
// __int64 __fastcall ioctl()
func libKernel_ioctl(fd, request, argPtr uintptr) uintptr {
	name, _ := FileDescriptorNames[fd]
	switch fd {
	case FD_GC, FD_CONSOLE, FD_TTY:
		fmt.Printf("%-120s %s requested %s with argument at %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprintf("[ioctl on %s]", name),
			color.Yellow.Sprintf("0x%X", request),
			color.Yellow.Sprintf("0x%X", argPtr),
		)
		return 0
	}
	fmt.Printf("%-120s %s requested %s with argument at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("[ioctl on unknown fd]"),
		color.Yellow.Sprintf("0x%X", request),
		color.Yellow.Sprintf("0x%X", argPtr),
	)

	return 0
}
