package ime

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
)

func RegisterImeStubs() {
	elf.RegisterStub("libSceIme", "sceImeKeyboardOpen", libSceIme_stub)
	elf.RegisterStub("libSceIme", "sceImeKeyboardClose", libSceIme_stub)
	elf.RegisterStub("libSceIme", "sceImeUpdate", libSceIme_stub)
}

func libSceIme_stub() uintptr {
	return 0
}
