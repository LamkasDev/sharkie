package lib

import "github.com/LamkasDev/sharkie/cmd/elf"

func RegisterGnmDriverStubs() {
	// Command functions.
	elf.RegisterStub("libSceGnmDriver", "sceGnmSubmitCommandBuffers", libSceGnmDriver_sceGnmSubmitCommandBuffers)
	elf.RegisterStub("libSceGnmDriver", "sceGnmSubmitCommandBuffersForWorkload", libSceGnmDriver_sceGnmSubmitCommandBuffersForWorkload)
	elf.RegisterStub("libSceGnmDriver", "sceGnmSubmitAndFlipCommandBuffers", libSceGnmDriver_sceGnmSubmitAndFlipCommandBuffers)
	elf.RegisterStub("libSceGnmDriver", "sceGnmSubmitAndFlipCommandBuffersForWorkload", libSceGnmDriver_sceGnmSubmitAndFlipCommandBuffersForWorkload)
	elf.RegisterStub("libSceGnmDriver", "sceGnmRequestFlipAndSubmitDone", libSceGnmDriver_sceGnmRequestFlipAndSubmitDone)
	elf.RegisterStub("libSceGnmDriver", "sceGnmRequestFlipAndSubmitDoneForWorkload", libSceGnmDriver_sceGnmRequestFlipAndSubmitDoneForWorkload)

	// More commands.
	elf.RegisterStub("libSceGnmDriver", "sceGnmSubmitDone", libSceGnmDriver_sceGnmSubmitDone)
	elf.RegisterStub("libSceGnmDriver", "sceGnmDingDong", libSceGnmDriver_sceGnmDingDong)
	elf.RegisterStub("libSceGnmDriver", "sceGnmDingDongForWorkload", libSceGnmDriver_sceGnmDingDongForWorkload)
}
