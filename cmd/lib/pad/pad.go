package pad

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
)

func RegisterPadStubs() {
	// Setup functions.
	elf.RegisterStub("libScePad", "scePadInit", libScePad_scePadInit)
	elf.RegisterStub("libScePad", "scePadOpen", libScePad_scePadOpen)
	elf.RegisterStub("libScePad", "scePadGetControllerInformation", libScePad_scePadGetControllerInformation)
	elf.RegisterStub("libScePad", "scePadRead", libScePad_scePadRead)
}
