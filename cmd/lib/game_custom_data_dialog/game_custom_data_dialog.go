package game_custom_data_dialog

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterGameCustomDataDialogStubs() {
	elf.RegisterStub("libSceGameCustomDataDialog", "sceGameCustomDataDialogOpen", libSceGameCustomDataDialog_stub)
	elf.RegisterStub("libSceGameCustomDataDialog", "sceGameCustomDataDialogInitialize", libSceGameCustomDataDialog_stub)
	elf.RegisterStub("libSceGameCustomDataDialog", "sceGameCustomDataDialogTerminate", libSceGameCustomDataDialog_stub)
	elf.RegisterStub("libSceGameCustomDataDialog", "sceGameCustomDataDialogClose", libSceGameCustomDataDialog_stub)
	elf.RegisterStub("libSceGameCustomDataDialog", "sceGameCustomDataDialogGetResult", libSceGameCustomDataDialog_stub)
	elf.RegisterStub("libSceGameCustomDataDialog", "sceGameCustomDataDialogUpdateStatus", libSceGameCustomDataDialog_stub)
}

func libSceGameCustomDataDialog_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
