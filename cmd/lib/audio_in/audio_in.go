package audio_in

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
)

func RegisterAudioInStubs() {
	// Setup functions.
	elf.RegisterStub("libSceAudioIn", "sceAudioInOpen", libSceAudioIn_sceAudioInOpen)
	elf.RegisterStub("libSceAudioIn", "sceAudioInOpenEx", libSceAudioIn_sceAudioInOpenEx)
	elf.RegisterStub("libSceAudioIn", "sceAudioInGetHandleStatusInfo", libSceAudioIn_sceAudioInGetHandleStatusInfo)
	elf.RegisterStub("libSceAudioIn", "sceAudioInGetSilentState", libSceAudioIn_sceAudioInGetSilentState)

	// Input functions.
	elf.RegisterStub("libSceAudioIn", "sceAudioInInput", libSceAudioIn_sceAudioInInput)
}
