package audio_out

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
)

func RegisterAudioOutStubs() {
	// Setup functions.
	elf.RegisterStub("libSceAudioOut", "sceAudioOutInit", libSceAudioOut_sceAudioOutInit)
	elf.RegisterStub("libSceAudioOut", "sceAudioOutOpen", libSceAudioOut_sceAudioOutOpen)
	elf.RegisterStub("libSceAudioOut", "sceAudioOutGetPortState", libSceAudioOut_sceAudioOutGetPortState)
	elf.RegisterStub("libSceAudioOut", "sceAudioOutSetVolume", libSceAudioOut_sceAudioOutSetVolume)

	// Output functions.
	elf.RegisterStub("libSceAudioOut", "sceAudioOutOutput", libSceAudioOut_sceAudioOutOutput)
	elf.RegisterStub("libSceAudioOut", "sceAudioOutOutputs", libSceAudioOut_sceAudioOutOutputs)
}
