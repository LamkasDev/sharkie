package lib

import "github.com/LamkasDev/sharkie/cmd/elf"

func RegisterVideoOutStubs() {
	// Command functions.
	elf.RegisterStub("libSceVideoOut", "sceVideoOutSubmitEopFlip", libSceVideoOut_sceVideoOutSubmitEopFlip)
}
