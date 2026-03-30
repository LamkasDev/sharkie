package lib

import "github.com/LamkasDev/sharkie/cmd/elf"

func RegisterVideoOutStubs() {
	// Setup functions.
	elf.RegisterStub("libSceVideoOut", "sceVideoOutOpen", libSceVideoOut_sceVideoOutOpen)
	elf.RegisterStub("libSceVideoOut", "sceVideoOutSetFlipRate", libSceVideoOut_sceVideoOutSetFlipRate)
	elf.RegisterStub("libSceVideoOut", "sceVideoOutGetBufferLabelAddress", libSceVideoOut_sceVideoOutGetBufferLabelAddress)

	// Command functions.
	elf.RegisterStub("libSceVideoOut", "sceVideoOutAddFlipEvent", libSceVideoOut_sceVideoOutAddFlipEvent)
	elf.RegisterStub("libSceVideoOut", "sceVideoOutSubmitEopFlip", libSceVideoOut_sceVideoOutSubmitEopFlip)

	// Buffer functions.
	elf.RegisterStub("libSceVideoOut", "sceVideoOutRegisterBuffers", libSceVideoOut_sceVideoOutRegisterBuffers)
	elf.RegisterStub("libSceVideoOut", "sceVideoOutRegisterBufferAttribute", libSceVideoOut_sceVideoOutRegisterBufferAttribute)
	elf.RegisterStub("libSceVideoOut", "sceVideoOutSetBufferAttribute", libSceVideoOut_sceVideoOutSetBufferAttribute)
}
