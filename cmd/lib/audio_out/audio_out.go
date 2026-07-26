package audio_out

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterAudioOutStubs() {
	// Setup functions.
	// elf.RegisterStub("libSceAudioOut", "sceAudioOutInit", libSceAudioOut_sceAudioOutInit)
	elf.RegisterStub("libSceAudioOut", "sceAudioOutOpen", libSceAudioOut_sceAudioOutOpen)
	elf.RegisterStub("libSceAudioOut", "sceAudioOutOutput", libSceAudioOut_sceAudioOutOutput)
	elf.RegisterStub("libSceAudioOut", "sceAudioOutGetPortState", libSceAudioOut_sceAudioOutGetPortState)
	elf.RegisterStub("libSceAudioOut", "sceAudioOutSetVolume", libSceAudioOut_sceAudioOutSetVolume)
	elf.RegisterStub("libSceAudioOut", "sceAudioOutOutput", libSceAudioOut_stub)
	elf.RegisterStub("libSceAudioOut", "sceAudioOutOutputs", libSceAudioOut_stub)
}

func libSceAudioOut_stub() uintptr {
	return 0
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
