package lib

import (
	"crypto/rand"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x0000000000000970
// __int64 __fastcall ioctl()
func libKernel_ioctl(fd, request, mode uintptr) uintptr {
	return libKernel_sys_ioctl(fd, request, mode)
}

func libKernel_sys_ioctl(fd, request, argPtr uintptr) uintptr {
	file, ok := GlobalFilesystem.Descriptors[FileDescriptor(fd)]
	if !ok {
		logger.Printf("%-132s %s failed due to unknown file %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ioctl"),
			color.Yellow.Sprintf("0x%X", fd),
		)
		return ENOENT
	}

	switch request {
	case SCE_NET_IOCTL_INIT:
		logger.Printf("%-132s %s initialized socket.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ioctl"),
		)
		return 0
	case SCE_RNG_IOCTL_GET_ENTROPY:
		size := (request >> 16) & 0x1FFF

		argSlice := unsafe.Slice((*byte)(unsafe.Pointer(argPtr)), size)
		if _, err := rand.Read(argSlice); err != nil {
			return SCE_KERNEL_ERROR_EINVAL
		}

		logger.Printf("%-132s %s wrote %s random bytes to %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("ioctl"),
			color.Yellow.Sprintf("0x%X", size),
			color.Yellow.Sprintf("0x%X", argPtr),
		)
		return 0
	}

	logger.Printf("%-132s %s requested %s with argument at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("[ioctl on %s]", file.Path),
		color.Yellow.Sprintf("0x%X", request),
		color.Yellow.Sprintf("0x%X", argPtr),
	)
	return 0
}
