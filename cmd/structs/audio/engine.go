// Package audio contains structs to emulate audio devices.
package audio

import (
	"github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/LamkasDev/sharkie/cmd/structs/fs"
)

var GlobalAudioEngine *AudioEngine

// AudioEngine keeps state of audio devices.
type AudioEngine struct {
	Handles    map[uint32]*AudioOutHandle
	NextHandle uint32
}

func NewAudioEngine() *AudioEngine {
	return &AudioEngine{
		Handles:    map[uint32]*AudioOutHandle{},
		NextHandle: 0x20000001,
	}
}

func SetupAudioEngine() {
	GlobalAudioEngine = NewAudioEngine()
	if _, err := fs.GlobalFilesystem.Write(fs.GetUsablePath(AudioInBufferName), make([]byte, AudioInBufferDefault)); err != nil {
		panic(err)
	}
	if _, err := fs.GlobalFilesystem.Write(fs.GetUsablePath(AudioVideoSettingsName), make([]byte, AudioVideoSettingsDefault)); err != nil {
		panic(err)
	}
	structs.CreateDefaultEventFlags([]string{
		AudioInEventFlagName,
	})
}
