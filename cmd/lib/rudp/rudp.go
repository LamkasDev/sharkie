package rudp

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterRudpStubs() {
	elf.RegisterStub("libSceRudp", "sceRudpInitiate", libSceRudp_stub)
	elf.RegisterStub("libSceRudp", "sceRudpBind", libSceRudp_stub)
	elf.RegisterStub("libSceRudp", "sceRudpCreateContext", libSceRudp_stub)
	elf.RegisterStub("libSceRudp", "sceRudpSetEventHandler", libSceRudp_stub)
	elf.RegisterStub("libSceRudp", "sceRudpEnd", libSceRudp_stub)
	elf.RegisterStub("libSceRudp", "sceRudpEnableInternalIOThread", libSceRudp_stub)
	elf.RegisterStub("libSceRudp", "sceRudpInit", libSceRudp_stub)
	elf.RegisterStub("libSceRudp", "sceRudpTerminate", libSceRudp_stub)
	elf.RegisterStub("libSceRudp", "sceRudpRead", libSceRudp_stub)
	elf.RegisterStub("libSceRudp", "sceRudpGetSizeReadable", libSceRudp_stub)
	elf.RegisterStub("libSceRudp", "sceRudpWrite", libSceRudp_stub)
}

func libSceRudp_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
