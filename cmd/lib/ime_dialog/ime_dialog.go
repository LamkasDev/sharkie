package ime_dialog

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterImeDialogStubs() {
	elf.RegisterStub("libSceImeDialog", "sceImeDialogAbort", libSceImeDialog_stub)
	elf.RegisterStub("libSceImeDialog", "sceImeDialogTerm", libSceImeDialog_stub)
	elf.RegisterStub("libSceImeDialog", "sceImeDialogGetResult", libSceImeDialog_stub)
	elf.RegisterStub("libSceImeDialog", "sceImeDialogGetStatus", libSceImeDialog_stub)
	elf.RegisterStub("libSceImeDialog", "sceImeDialogInit", libSceImeDialog_stub)
}

func libSceImeDialog_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
