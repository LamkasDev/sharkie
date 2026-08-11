package audio_out

import (
	"github.com/LamkasDev/sharkie/cmd/lib_structs"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
)

const AudioInBufferDefault = 4096
const AudioInBufferName = "/vmicDdShmAin"
const AudioInEventFlagName = "/vmicDdEvfAin"

const AudioVideoSettingsDefault = 4096
const AudioVideoSettingsName = "/SceAvSetting"

func SetupLegacyAudioStuff() {
	if _, err := fs.GlobalFilesystem.Write(fs.GlobalFilesystem.GetUsablePath(AudioInBufferName), make([]byte, AudioInBufferDefault)); err != nil {
		panic(err)
	}
	if _, err := fs.GlobalFilesystem.Write(fs.GlobalFilesystem.GetUsablePath(AudioVideoSettingsName), make([]byte, AudioVideoSettingsDefault)); err != nil {
		panic(err)
	}
	lib_structs.CreateDefaultEventFlags([]string{
		AudioInEventFlagName,
	})
}
