package ime

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
)

func RegisterImeStubs() {
	// Setup funtions.
	elf.RegisterStub("libSceIme", "sceImeOpen", libSceIme_sceImeOpen)
	elf.RegisterStub("libSceIme", "sceImeClose", libSceIme_sceImeClose)
	elf.RegisterStub("libSceIme", "sceImeUpdate", libSceIme_sceImeUpdate)

	// Keyboard functions.
	elf.RegisterStub("libSceIme", "sceImeKeyboardOpen", libSceIme_sceImeKeyboardOpen)
	elf.RegisterStub("libSceIme", "sceImeKeyboardClose", libSceIme_sceImeKeyboardClose)
	elf.RegisterStub("libSceIme", "sceImeKeyboardUpdate", libSceIme_sceImeKeyboardUpdate)
	elf.RegisterStub("libSceIme", "sceImeKeyboardGetResourceId", libSceIme_sceImeKeyboardGetResourceId)
}
