package ime_dialog

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
)

func RegisterImeDialogStubs() {
	// Setup functions.
	elf.RegisterStub("libSceImeDialog", "sceImeDialogInit", libSceImeDialog_sceImeDialogInit)
	elf.RegisterStub("libSceImeDialog", "sceImeDialogTerm", libSceImeDialog_sceImeDialogTerm)
	elf.RegisterStub("libSceImeDialog", "sceImeDialogAbort", libSceImeDialog_sceImeDialogAbort)

	// State functions.
	elf.RegisterStub("libSceImeDialog", "sceImeDialogGetStatus", libSceImeDialog_sceImeDialogGetStatus)
	elf.RegisterStub("libSceImeDialog", "sceImeDialogGetResult", libSceImeDialog_sceImeDialogGetResult)
}
