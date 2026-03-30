package lib

import "github.com/LamkasDev/sharkie/cmd/elf"

func RegisterGnmDriverStubs() {
	// Command functions.
	elf.RegisterStub("libSceGnmDriver", "sceGnmSubmitCommandBuffersForWorkload", libSceGnmDriver_sceGnmSubmitCommandBuffersForWorkload)
	elf.RegisterStub("libSceGnmDriver", "sceGnmSubmitAndFlipCommandBuffersForWorkload", libSceGnmDriver_sceGnmSubmitAndFlipCommandBuffersForWorkload)
}
