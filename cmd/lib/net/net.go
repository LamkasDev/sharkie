package net

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterNetStubs() {
	elf.RegisterStub("libSceNet", "sceNetTerm", libSceNet_stub)
	elf.RegisterStub("libSceNet", "sceNetInit", libSceNet_stub)
	elf.RegisterStub("libSceNet", "sceNetPoolDestroy", libSceNet_stub)
	elf.RegisterStub("libSceNet", "sceNetGetMacAddress", libSceNet_stub)
	elf.RegisterStub("libSceNet", "sceNetSocketClose", libSceNet_stub)
	elf.RegisterStub("libSceNet", "sceNetBind", libSceNet_stub)
	elf.RegisterStub("libSceNet", "sceNetSetsockopt", libSceNet_stub)
	elf.RegisterStub("libSceNet", "sceNetSocket", libSceNet_stub)
	elf.RegisterStub("libSceNet", "sceNetHtons", libSceNet_stub)
	elf.RegisterStub("libSceNet", "sceNetPoolCreate", libSceNet_stub)
}

func libSceNet_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
