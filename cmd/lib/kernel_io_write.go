package lib

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

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

	buffSlice := unsafe.Slice((*byte)(unsafe.Pointer(bufPtr)), length)
	message := string(buffSlice)
	name, ok := FileDescriptorNames[fd]
	if !ok {
		fmt.Printf("%-120s %s %s",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprintf("[write on unknown %d]", fd),
			FileDescriptorColors[fd].Sprint(message),
		)
		if !strings.HasSuffix(message, "\n") {
			fmt.Println("")
		}
		return 0
	}

	fmt.Printf("%-120s %s %s",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("[write on %s]", name),
		message,
	)
	if !strings.HasSuffix(message, "\n") {
		fmt.Println("")
	}
	return length
}
