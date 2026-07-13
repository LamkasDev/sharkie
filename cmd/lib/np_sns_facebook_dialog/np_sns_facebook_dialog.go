package np_sns_facebook_dialog

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterNpSnsFacebookDialogStubs() {
	elf.RegisterStub("libSceNpSnsFacebookDialog", "sceNpSnsFacebookDialogTerminate", libSceNpSnsFacebookDialog_stub)
	elf.RegisterStub("libSceNpSnsFacebookDialog", "sceNpSnsFacebookDialogGetResult", libSceNpSnsFacebookDialog_stub)
	elf.RegisterStub("libSceNpSnsFacebookDialog", "sceNpSnsFacebookDialogUpdateStatus", libSceNpSnsFacebookDialog_stub)
	elf.RegisterStub("libSceNpSnsFacebookDialog", "sceNpSnsFacebookDialogOpen", libSceNpSnsFacebookDialog_stub)
	elf.RegisterStub("libSceNpSnsFacebookDialog", "sceNpSnsFacebookDialogInitialize", libSceNpSnsFacebookDialog_stub)
}

func libSceNpSnsFacebookDialog_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
