package av_player

import "github.com/LamkasDev/sharkie/cmd/elf"

func RegisterAvPlayerStubs() {
	return
	// Setup functions.
	elf.RegisterStub("libSceAvPlayer", "sceAvPlayerInit", libSceAvPlayer_sceAvPlayerInit)
	elf.RegisterStub("libSceAvPlayer", "sceAvPlayerPostInit", libSceAvPlayer_sceAvPlayerPostInit)

	// Query functions.
	elf.RegisterStub("libSceAvPlayer", "sceAvPlayerIsActive", libSceAvPlayer_sceAvPlayerIsActive)
	elf.RegisterStub("libSceAvPlayer", "sceAvPlayerGetAudioData", libSceAvPlayer_sceAvPlayerGetAudioData)
	elf.RegisterStub("libSceAvPlayer", "sceAvPlayerGetVideoDataEx", libSceAvPlayer_sceAvPlayerGetVideoDataEx)

	// State functions.
	elf.RegisterStub("libSceAvPlayer", "sceAvPlayerAddSource", libSceAvPlayer_sceAvPlayerAddSource)
}
