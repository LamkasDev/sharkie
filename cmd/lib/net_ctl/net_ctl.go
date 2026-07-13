package net_ctl

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterNetCtlStubs() {
	elf.RegisterStub("libSceNetCtl", "sceNetCtlTerm", libSceNetCtl_stub)
	elf.RegisterStub("libSceNetCtl", "sceNetCtlGetNatInfo", libSceNetCtl_stub)
	elf.RegisterStub("libSceNetCtl", "sceNetCtlGetInfo", libSceNetCtl_sceNetCtlGetInfo)
	elf.RegisterStub("libSceNetCtl", "sceNetCtlUnregisterCallback", libSceNetCtl_stub)
	elf.RegisterStub("libSceNetCtl", "sceNetCtlGetResult", libSceNetCtl_sceNetCtlGetResult)
	elf.RegisterStub("libSceNetCtl", "sceNetCtlCheckCallback", libSceNetCtl_sceNetCtlCheckCallback)
	elf.RegisterStub("libSceNetCtl", "sceNetCtlRegisterCallback", libSceNetCtl_sceNetCtlRegisterCallback)
	elf.RegisterStub("libSceNetCtl", "sceNetCtlInit", libSceNetCtl_stub)
	elf.RegisterStub("libSceNetCtl", "sceNetCtlGetState", libSceNetCtl_sceNetCtlGetState)
}

func libSceNetCtl_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
