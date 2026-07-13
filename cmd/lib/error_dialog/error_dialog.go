package error_dialog

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterErrorDialogStubs() {
	elf.RegisterStub("libSceErrorDialog", "sceErrorDialogTerminate", libSceErrorDialog_stub)
	elf.RegisterStub("libSceErrorDialog", "sceErrorDialogUpdateStatus", libSceErrorDialog_stub)
	elf.RegisterStub("libSceErrorDialog", "sceErrorDialogOpen", libSceErrorDialog_stub)
	elf.RegisterStub("libSceErrorDialog", "sceErrorDialogInitialize", libSceErrorDialog_stub)
}

func libSceErrorDialog_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
