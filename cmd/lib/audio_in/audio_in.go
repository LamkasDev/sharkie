package audio_in

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
)

func RegisterAudioInStubs() {
	// Setup functions.
	elf.RegisterStub("libSceAudioIn", "sceAudioInOpen", libSceAudioIn_sceAudioInOpen)

	// Input functions.
	elf.RegisterStub("libSceAudioIn", "sceAudioInInput", libSceAudioIn_sceAudioInInput)
}
