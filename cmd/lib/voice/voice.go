package voice

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterVoiceStubs() {
	elf.RegisterStub("libSceVoice", "sceVoiceDisconnectIPortFromOPort", libSceVoice_stub)
	elf.RegisterStub("libSceVoice", "sceVoiceConnectIPortToOPort", libSceVoice_stub)
	elf.RegisterStub("libSceVoice", "sceVoiceReadFromOPort", libSceVoice_stub)
	elf.RegisterStub("libSceVoice", "sceVoiceGetPortAttr", libSceVoice_stub)
	elf.RegisterStub("libSceVoice", "sceVoiceWriteToIPort", libSceVoice_stub)
	elf.RegisterStub("libSceVoice", "sceVoiceGetPortInfo", libSceVoice_stub)
	elf.RegisterStub("libSceVoice", "sceVoiceDeletePort", libSceVoice_stub)
	elf.RegisterStub("libSceVoice", "sceVoiceEnd", libSceVoice_stub)
	elf.RegisterStub("libSceVoice", "sceVoiceStop", libSceVoice_stub)
	elf.RegisterStub("libSceVoice", "sceVoiceStart", libSceVoice_stub)
	elf.RegisterStub("libSceVoice", "sceVoiceCreatePort", libSceVoice_stub)
	elf.RegisterStub("libSceVoice", "sceVoiceInit", libSceVoice_stub)
}

func libSceVoice_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
