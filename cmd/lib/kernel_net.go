package lib

import (
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x0000000000000B90
// __int64 __fastcall _sys_netcontrol()
func libKernel___sys_netcontrol(fd uintptr, op uintptr, resultPtr uintptr, length uintptr) uintptr {
	switch op {
	case NETC_GET_MEM_INFO:
		if length < 4 || resultPtr == 0 {
			logger.Printf("%-132s %s failed due to invalid result pointer.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("__sys_netcontrol"),
			)
			return EINVAL
		}
		resultSlice := unsafe.Slice((*byte)(unsafe.Pointer(resultPtr)), length)
		binary.LittleEndian.PutUint32(resultSlice, SocketBufferSize)

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
func libKernel___sys_socketex(namePtr uintptr, domain uintptr, sockType uintptr, protocol uintptr) uintptr {
	name := "unnamed-socket"
	if namePtr != 0 {
		name = ReadCString(namePtr)
	}

	file, err := GlobalFilesystem.Open(name, SCE_O_CREAT, 0)
	if err != nil {
		logger.Printf("%-132s %s failed due to open error on %s (%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("__sys_socketex"),
			color.Blue.Sprint(name),
			err.Error(),
		)
		SetErrno(ENOENT)
		return ERR_PTR
	}
	file.ExtraData = &Socket{
		Name:     name,
		Domain:   int32(domain),
		Type:     int32(sockType),
		Protocol: int32(protocol),
	}

	logger.Printf("%-132s %s created socket %s (name=%s, domain=%s, sockType=%s, protocol=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("__sys_socketex"),
		color.Yellow.Sprintf("0x%X", file.Descriptor),
		color.Blue.Sprint(name),
		color.Yellow.Sprintf("0x%X", domain),
		color.Yellow.Sprintf("0x%X", sockType),
		color.Yellow.Sprintf("0x%X", protocol),
	)
	return uintptr(file.Descriptor)
}

// 0x0000000000000C90
// __int64 __fastcall _sys_socketclose()
func libKernel___sys_socketclose(fd uintptr) uintptr {
	logger.Printf("%-132s %s tried closing socket %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("__sys_socketclose"),
		color.Yellow.Sprintf("0x%X", fd),
	)
	return 0
}
