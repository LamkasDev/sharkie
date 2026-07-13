package save_data_dialog

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterSaveDataDialogStubs() {
	elf.RegisterStub("libSceSaveDataDialog", "sceSaveDataDialogTerminate", libSceSaveDataDialog_stub)
	elf.RegisterStub("libSceSaveDataDialog", "sceSaveDataDialogGetResult", libSceSaveDataDialog_stub)
	elf.RegisterStub("libSceSaveDataDialog", "sceSaveDataDialogUpdateStatus", libSceSaveDataDialog_stub)
	elf.RegisterStub("libSceSaveDataDialog", "sceSaveDataDialogOpen", libSceSaveDataDialog_stub)
	elf.RegisterStub("libSceSaveDataDialog", "sceSaveDataDialogInitialize", libSceSaveDataDialog_stub)
}

func libSceSaveDataDialog_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
