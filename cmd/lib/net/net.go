package net

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	structsNet "github.com/LamkasDev/sharkie/cmd/lib_structs/net"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterNetStubs() {
	elf.RegisterStub("libSceNet", "sceNetTerm", libSceNet_stub)
	elf.RegisterStub("libSceNet", "sceNetInit", libSceNet_stub)
	elf.RegisterStub("libSceNet", "sceNetPoolDestroy", libSceNet_stub)
	elf.RegisterStub("libSceNet", "sceNetGetMacAddress", libSceNet_sceNetGetMacAddress)
	elf.RegisterStub("libSceNet", "sceNetSocketClose", libSceNet_stub)
	elf.RegisterStub("libSceNet", "sceNetBind", libSceNet_stub)
	elf.RegisterStub("libSceNet", "sceNetSetsockopt", libSceNet_stub)
	elf.RegisterStub("libSceNet", "sceNetSocket", libSceNet_stub)
	elf.RegisterStub("libSceNet", "sceNetHtons", libSceNet_stub)
	elf.RegisterStub("libSceNet", "sceNetPoolCreate", libSceNet_stub)
	elf.RegisterStub("libSceNet", "sceNetRecvfrom", libSceNet_stub2)
}

func libSceNet_stub2() uintptr {
	return 0
}

func libSceNet_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}

func libSceNet_sceNetGetMacAddress(addrPtr uintptr, flags int32) uintptr {
	if addrPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid addr pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceNetGetMacAddress"),
		)
		return 0x80410116
	}

	addr := (*[6]byte)(unsafe.Pointer(addrPtr))
	*addr = structsNet.GlobalNetConnectionInstance.MacAddress

	if logger.LogMisc {
		logger.Printf("%-132s %s returned %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceNetGetMacAddress"),
			color.Yellow.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", addr[0], addr[1], addr[2], addr[3], addr[4], addr[5]),
		)
	}

	return 0
}
