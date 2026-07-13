package app_content

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterAppContentStubs() {
	elf.RegisterStub("libSceAppContent", "sceAppContentGetEntitlementKey", libSceAppContent_stub)
	elf.RegisterStub("libSceAppContent", "sceAppContentGetAddcontInfoList", libSceAppContent_stub)
	elf.RegisterStub("libSceAppContent", "sceAppContentAppParamGetInt", libSceAppContent_stub)
	elf.RegisterStub("libSceAppContent", "sceAppContentInitialize", libSceAppContent_stub)
}

func libSceAppContent_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
