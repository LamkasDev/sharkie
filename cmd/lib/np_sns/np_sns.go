package np_sns

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterNpSnsStubs() {
	elf.RegisterStub("libSceNpSns", "sceNpSnsFacebookDeleteRequest", libSceNpSns_stub)
	elf.RegisterStub("libSceNpSns", "sceNpSnsFacebookGetAccessToken", libSceNpSns_stub)
	elf.RegisterStub("libSceNpSns", "sceNpSnsFacebookCreateRequest", libSceNpSns_stub)
}

func libSceNpSns_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
