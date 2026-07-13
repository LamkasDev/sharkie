package np_auth

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterNpAuthStubs() {
	elf.RegisterStub("libSceNpAuth", "sceNpAuthDeleteRequest", libSceNpAuth_stub)
	elf.RegisterStub("libSceNpAuth", "sceNpAuthGetAuthorizationCode", libSceNpAuth_stub)
	elf.RegisterStub("libSceNpAuth", "sceNpAuthCreateRequest", libSceNpAuth_stub)
}

func libSceNpAuth_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
