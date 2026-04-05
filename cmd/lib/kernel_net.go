package lib

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs/fs"
	. "github.com/LamkasDev/sharkie/cmd/structs/net"
	"github.com/gookit/color"
)

// 0x0000000000000B90
// __int64 __fastcall _sys_netcontrol()
func libKernel___sys_netcontrol(fd FileDescriptor, op, resultPtr, length uintptr) uintptr {
	switch op {
	case NETC_GET_MEM_INFO:
		if length < NetworkMemoryInfoSize || resultPtr == 0 {
			logger.Printf("%-132s %s failed due to invalid result pointer.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("__sys_netcontrol"),
			)
			return EINVAL
		}
		memoryInfo := (*NetworkMemoryInfo)(unsafe.Pointer(resultPtr))
		memoryInfo.BufferSize = SocketBufferSize

		logger.Printf("%-132s %s returned network memory info.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("__sys_netcontrol"),
		)
		return 0
	}

	logger.Printf("%-132s %s failed due to unknown operation %s (fd=%s, resultPtr=%s, length=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("__sys_netcontrol"),
		color.Yellow.Sprintf("0x%X", op),
		color.Yellow.Sprintf("0x%X", fd),
		color.Yellow.Sprintf("0x%X", resultPtr),
		color.Yellow.Sprintf("0x%X", length),
	)
	return EINVAL
}

// 0x0000000000000C70
// __int64 __fastcall _sys_socketex()
func libKernel___sys_socketex(namePtr Cstring, domain, sockType, protocol uintptr) uintptr {
	name := "unnamed-socket"
	if namePtr != nil {
		name = GoString(namePtr)
	}

	socket := &Socket{
		Name:     name,
		Domain:   int32(domain),
		Type:     int32(sockType),
		Protocol: int32(protocol),
	}
	fd := GlobalFilesystem.AllocateFd(name, socket)

	logger.Printf("%-132s %s created socket %s (name=%s, domain=%s, sockType=%s, protocol=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("__sys_socketex"),
		color.Yellow.Sprintf("0x%X", fd),
		color.Blue.Sprint(name),
		color.Yellow.Sprintf("0x%X", domain),
		color.Yellow.Sprintf("0x%X", sockType),
		color.Yellow.Sprintf("0x%X", protocol),
	)
	return uintptr(fd)
}

// 0x0000000000000C90
// __int64 __fastcall _sys_socketclose()
func libKernel___sys_socketclose(fd FileDescriptor) uintptr {
	logger.Printf("%-132s %s tried closing socket %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("__sys_socketclose"),
		color.Yellow.Sprintf("0x%X", fd),
	)
	return 0
}
