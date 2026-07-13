package sysmodule

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterSysmoduleStubs() {
	elf.RegisterStub("libSceSysmodule", "sceSysmoduleIsLoaded", libSceSysmodule_stub)
	elf.RegisterStub("libSceSysmodule", "sceSysmoduleLoadModule", libSceSysmodule_stub)
	elf.RegisterStub("libSceSysmodule", "sceSysmoduleUnloadModule", libSceSysmodule_stub)
}

func libSceSysmodule_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
