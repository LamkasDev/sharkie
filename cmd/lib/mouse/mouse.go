package mouse

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterMouseStubs() {
	elf.RegisterStub("libSceMouse", "sceMouseInit", libSceMouse_stub)
	elf.RegisterStub("libSceMouse", "sceMouseOpen", libSceMouse_stub)
	elf.RegisterStub("libSceMouse", "sceMouseClose", libSceMouse_stub)

	// Read functions.
	elf.RegisterStub("libSceMouse", "sceMouseRead", libSceMouse_sceMouseRead)
}

func libSceMouse_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
