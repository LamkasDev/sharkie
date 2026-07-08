// Package audio contains structs to emulate audio devices.
package audio

import (
	"sync"

	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
)

var GlobalAudioEngine *AudioEngine

// AudioEngine keeps state of audio devices.
type AudioEngine struct {
	Handles    map[uint32]*AudioOutHandle
	NextHandle uint32
	Lock       sync.Mutex
}

func NewAudioEngine() *AudioEngine {
	return &AudioEngine{
		Handles:    map[uint32]*AudioOutHandle{},
		NextHandle: 0x20000001,
		Lock:       sync.Mutex{},
	}
}

func (ae *AudioEngine) CreateHandle() *AudioOutHandle {
	ae.Lock.Lock()
	defer ae.Lock.Unlock()
	handle := &AudioOutHandle{
		Id: ae.NextHandle,
	}
	ae.Handles[handle.Id] = handle
	ae.NextHandle++

	return handle
}

func SetupAudioEngine() {
	GlobalAudioEngine = NewAudioEngine()
	if _, err := fs.GlobalFilesystem.Write(fs.GetUsablePath(AudioInBufferName), make([]byte, AudioInBufferDefault)); err != nil {
		panic(err)
	}
	if _, err := fs.GlobalFilesystem.Write(fs.GetUsablePath(AudioVideoSettingsName), make([]byte, AudioVideoSettingsDefault)); err != nil {
		panic(err)
	}
	CreateDefaultEventFlags([]string{
		AudioInEventFlagName,
	})
}
