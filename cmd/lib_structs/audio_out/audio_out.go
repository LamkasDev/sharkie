package audio_out

import (
	"io"

	"github.com/ebitengine/oto/v3"
)

type AudioOutPortType int32

const (
	AudioOutPortMain     = AudioOutPortType(0)
	AudioOutPortBgm      = AudioOutPortType(1)
	AudioOutPortVoice    = AudioOutPortType(2)
	AudioOutPortPersonal = AudioOutPortType(3)
	AudioOutPortPadSpk   = AudioOutPortType(4)
	AudioOutPortAudio3d  = AudioOutPortType(126)
	AudioOutPortAux      = AudioOutPortType(127)
)

type AudioOutPortOutput uint16

const (
	AudioOutStateOutputUnknown            = AudioOutPortOutput(0x00)
	AudioOutStateOutputConnectedPrimary   = AudioOutPortOutput(0x01)
	AudioOutStateOutputConnectedSecondary = AudioOutPortOutput(0x02)
	AudioOutStateOutputConnectedTertiary  = AudioOutPortOutput(0x04)
	AudioOutStateOutputConnectedHeadphone = AudioOutPortOutput(0x40)
	AudioOutStateOutputConnectedExternal  = AudioOutPortOutput(0x80)
)

type AudioOutParamFormat int32

const (
	AudioOutParamFormatS16Mono       = AudioOutParamFormat(0)
	AudioOutParamFormatS16Stereo     = AudioOutParamFormat(1)
	AudioOutParamFormatS16_8CH       = AudioOutParamFormat(2)
	AudioOutParamFormatFloatMono     = AudioOutParamFormat(3)
	AudioOutParamFormatFloatStereo   = AudioOutParamFormat(4)
	AudioOutParamFormatFloat_8CH     = AudioOutParamFormat(5)
	AudioOutParamFormatS16_8CH_Std   = AudioOutParamFormat(6)
	AudioOutParamFormatFloat_8CH_Std = AudioOutParamFormat(7)
)

type AudioOutFormatInfo struct {
	IsFloat    bool
	SampleSize uint32
	Channels   uint32
}

func (format AudioOutParamFormat) ToAudioOutFormatInfo() AudioOutFormatInfo {
	switch format {
	case AudioOutParamFormatS16Mono:
		return AudioOutFormatInfo{IsFloat: false, SampleSize: 2, Channels: 1}
	case AudioOutParamFormatS16Stereo:
		return AudioOutFormatInfo{IsFloat: false, SampleSize: 2, Channels: 2}
	case AudioOutParamFormatS16_8CH, AudioOutParamFormatS16_8CH_Std:
		return AudioOutFormatInfo{IsFloat: false, SampleSize: 2, Channels: 8}
	case AudioOutParamFormatFloatMono:
		return AudioOutFormatInfo{IsFloat: true, SampleSize: 4, Channels: 1}
	case AudioOutParamFormatFloatStereo:
		return AudioOutFormatInfo{IsFloat: true, SampleSize: 4, Channels: 2}
	case AudioOutParamFormatFloat_8CH, AudioOutParamFormatFloat_8CH_Std:
		return AudioOutFormatInfo{IsFloat: true, SampleSize: 4, Channels: 8}
	}
	return AudioOutFormatInfo{IsFloat: false, SampleSize: 2, Channels: 2} // Fallback
}

type AudioOutOutputParam struct {
	Handle int32
	Ptr    uint64
}

type AudioOutPortState struct {
	Output         AudioOutPortOutput
	Channel        uint8
	Reserved1      [1]byte
	Volume         int16
	RerouteCounter uint16
	Flag           uint64
	Reserved2      [16]byte
}

type AudioOutHandle struct {
	Id         uint32
	PortId     int
	PortType   AudioOutPortType
	Length     uint32
	SampleRate uint32
	Format     AudioOutParamFormat
	Channels   int

	PipeWriter *io.PipeWriter
	Player     *oto.Player
}
