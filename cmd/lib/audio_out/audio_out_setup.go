package audio_out

import (
	"io"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/audio_out"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000000D40
// __int64 __fastcall sceAudioOutInit(__m128 _XMM0, __int64, __int64)
func libSceAudioOut_sceAudioOutInit() uintptr {
	return 0
}

// 0x0000000000000420
// __int64 __fastcall sceAudioOutOpen(unsigned int, int, unsigned int, unsigned int, unsigned int, unsigned int)
func libSceAudioOut_sceAudioOutOpen(userId int32, portType AudioOutPortType, index int32, length uint32, sampleFreq uint32, param uint32) uintptr {
	handle := GlobalAudioOutputEngine.CreateHandle()
	handle.PortId = int(index)
	handle.PortType = portType
	handle.Length = length
	handle.SampleRate = sampleFreq

	// Parse format and channels.
	handle.Format = AudioOutParamFormat(param & 0xFF)
	channels := 2
	switch handle.Format {
	case AudioOutParamFormatS16Mono, AudioOutParamFormatFloatMono:
		channels = 1
	case AudioOutParamFormatS16Stereo, AudioOutParamFormatFloatStereo:
		channels = 2
	case AudioOutParamFormatS16_8CH, AudioOutParamFormatFloat_8CH, AudioOutParamFormatS16_8CH_Std, AudioOutParamFormatFloat_8CH_Std:
		channels = 8
	}
	handle.Channels = channels

	// Create oto player.
	if GlobalAudioOutputEngine.Context != nil {
		pr, pw := io.Pipe()
		handle.PipeWriter = pw
		handle.Player = GlobalAudioOutputEngine.Context.NewPlayer(pr)
		handle.Player.Play()
	}

	logger.Printf("%-132s %s returned %s (userId=%s, portType=%s, index=%s, length=%s, sampleFreq=%s, param=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceAudioOutOpen"),
		color.Yellow.Sprintf("0x%X", handle.Id),
		color.Yellow.Sprintf("0x%X", userId),
		color.Yellow.Sprintf("0x%X", portType),
		color.Yellow.Sprintf("0x%X", index),
		color.Yellow.Sprintf("0x%X", length),
		color.Yellow.Sprintf("0x%X", sampleFreq),
		color.Yellow.Sprintf("0x%X", param),
	)
	return uintptr(handle.Id)
}

// 0x0000000000002D90
// __int64 __fastcall sceAudioOutGetPortState(int a1, unsigned __int8 *a2)
func libSceAudioOut_sceAudioOutGetPortState(handleId uint32, state *AudioOutPortState) uintptr {
	handle := GlobalAudioOutputEngine.GetHandle(handleId)
	if handle == nil || state == nil {
		logger.Printf("%-132s %s failed due to invalid handle or state pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceAudioOutGetPortState"),
		)
		return 0x802F0001
	}
	switch handle.PortType {
	case AudioOutPortMain, AudioOutPortBgm, AudioOutPortAudio3d:
		state.Output = AudioOutStateOutputConnectedPrimary
		if handle.Channels > 2 {
			state.Channel = 2
		} else {
			state.Channel = uint8(handle.Channels)
		}
	case AudioOutPortVoice, AudioOutPortPersonal:
		state.Output = AudioOutStateOutputConnectedHeadphone
		state.Channel = 1
	case AudioOutPortPadSpk:
		state.Output = AudioOutStateOutputConnectedTertiary
		state.Channel = 1
		state.Volume = 127
	case AudioOutPortAux:
		state.Output = AudioOutStateOutputConnectedExternal
		state.Channel = 0
	default:
		state.Output = AudioOutStateOutputUnknown
		state.Channel = 0
	}
	if handle.PortType != AudioOutPortPadSpk {
		state.Volume = -1
	}
	state.RerouteCounter = 0
	state.Flag = 0

	return 0
}

// 0x00000000000010B0
// __int64 __fastcall sceAudioOutSetVolume(int, int _ESI, __int64, __m128 _XMM0, __m128 _XMM1, __m128 _XMM2, __m128 _XMM3, __m128 _XMM4, __m128 _XMM5, __m128 _XMM6)
func libSceAudioOut_sceAudioOutSetVolume(handleId uint32, flag uint32, volPtr *uint64) uintptr {
	handle := GlobalAudioOutputEngine.GetHandle(handleId)
	if handle == nil || volPtr == nil {
		logger.Printf("%-132s %s failed due to invalid handle or volume pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceAudioOutSetVolume"),
		)
		return 0x802F0001
	}
	// TODO: finish this.

	return 0
}
