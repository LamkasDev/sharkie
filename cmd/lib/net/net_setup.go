package net

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/net"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x00000000000037F0
// __int64 __fastcall sceNetGetMacAddress(__int64, unsigned int)
func libSceNet_sceNetGetMacAddress(addrPtr uintptr, flags int32) uintptr {
	if addrPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid addr pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceNetGetMacAddress"),
		)
		return 0x80410116
	}

	addr := (*[6]byte)(unsafe.Pointer(addrPtr))
	*addr = GlobalNetConnectionInstance.MacAddress

	if logger.LogMisc {
		logger.Printf("%-132s %s returned %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceNetGetMacAddress"),
			color.Yellow.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", addr[0], addr[1], addr[2], addr[3], addr[4], addr[5]),
		)
	}

	return 0
}

var NextNetPoolId = uint32(0x1001)

// 0x0000000000001B90
// __int64 __fastcall sceNetPoolCreate(__int64, unsigned int, unsigned int)
func libSceNet_sceNetPoolCreate(name Cstring, size uint64, flags int32) uintptr {
	id := NextNetPoolId
	NextNetPoolId++

	logger.Printf("%-132s %s returned %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceNetPoolCreate"),
		color.Yellow.Sprintf("0x%X", id),
	)
	return uintptr(id)
}
