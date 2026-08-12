// Package audio_out contains structs to emulate audio output devices.
package audio_out

import (
	"sync"

	"github.com/ebitengine/oto/v3"
)

var GlobalAudioOutputEngine *AudioOutputEngine

// AudioOutputEngine keeps state of audio output devices.
type AudioOutputEngine struct {
	Context    *oto.Context
	Handles    map[uint32]*AudioOutHandle
	NextHandle uint32
	Lock       sync.RWMutex
}

func NewAudioOutputEngine() *AudioOutputEngine {
	ctx, ready, err := oto.NewContext(&oto.NewContextOptions{
		SampleRate:   48000,
		ChannelCount: 2,
		Format:       oto.FormatSignedInt16LE,
	})
	if err != nil {
		panic("Audio init failed: " + err.Error())
	}
	<-ready // Wait for audio context to be ready.

	return &AudioOutputEngine{
		Context:    ctx,
		Handles:    map[uint32]*AudioOutHandle{},
		NextHandle: 0x1001,
		Lock:       sync.RWMutex{},
	}
}

func (ae *AudioOutputEngine) CreateHandle() *AudioOutHandle {
	ae.Lock.Lock()
	defer ae.Lock.Unlock()
	handle := &AudioOutHandle{
		Id: ae.NextHandle,
	}
	ae.Handles[handle.Id] = handle
	ae.NextHandle++

	return handle
}

func (ae *AudioOutputEngine) GetHandle(id uint32) *AudioOutHandle {
	ae.Lock.RLock()
	defer ae.Lock.RUnlock()
	return ae.Handles[id]
}

func SetupAudioOutputEngine() {
	SetupLegacyAudioStuff()
	GlobalAudioOutputEngine = NewAudioOutputEngine()
}
