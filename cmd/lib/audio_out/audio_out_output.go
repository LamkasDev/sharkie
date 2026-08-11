package audio_out

import (
	"encoding/binary"
	"math"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/audio_out"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func handleAudioOutput(handleId uint32, ptr uint64) (int32, int32) {
	handle := GlobalAudioOutputEngine.GetHandle(handleId)
	if handle == nil {
		logger.Printf("%-132s %s failed due to invalid handle.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("handleAudioOutput"),
		)
		return -1, 0
	}
	if handle.PipeWriter == nil || ptr == 0 {
		return 0, 0
	}
	info := handle.Format.ToAudioOutFormatInfo()
	bytesToRead := handle.Length * info.Channels * info.SampleSize
	pcmData := unsafe.Slice((*byte)(unsafe.Pointer(uintptr(ptr))), bytesToRead)

	// Convert any format to S16Stereo for oto.
	var outData []byte
	switch handle.Format {
	case AudioOutParamFormatS16Mono:
		outData = make([]byte, handle.Length*4) // 2 channels * 2 bytes.
		for i := range handle.Length {
			sample := pcmData[i*2 : i*2+2]
			copy(outData[i*4:], sample)
			copy(outData[i*4+2:], sample)
		}
	case AudioOutParamFormatS16Stereo:
		outData = pcmData // Already matches.
	case AudioOutParamFormatFloatMono:
		outData = make([]byte, handle.Length*4)
		for i := range handle.Length {
			fBits := binary.LittleEndian.Uint32(pcmData[i*4 : i*4+4])
			f := math.Float32frombits(fBits)
			// Clamp to [-1, 1]
			if f > 1.0 {
				f = 1.0
			} else if f < -1.0 {
				f = -1.0
			}
			s16 := int16(f * 32767.0)
			binary.LittleEndian.PutUint16(outData[i*4:], uint16(s16))
			binary.LittleEndian.PutUint16(outData[i*4+2:], uint16(s16)) // duplicate to stereo.
		}
	case AudioOutParamFormatFloatStereo:
		outData = make([]byte, handle.Length*4)
		for i := range handle.Length * 2 {
			fBits := binary.LittleEndian.Uint32(pcmData[i*4:])
			f := math.Float32frombits(fBits)
			// Clamp to [-1, 1].
			if f > 1.0 {
				f = 1.0
			} else if f < -1.0 {
				f = -1.0
			}
			s16 := int16(f * 32767.0)
			binary.LittleEndian.PutUint16(outData[i*2:], uint16(s16))
		}
	case AudioOutParamFormatFloat_8CH, AudioOutParamFormatFloat_8CH_Std:
		outData = make([]byte, handle.Length*4)
		for i := range handle.Length {
			fBitsL := binary.LittleEndian.Uint32(pcmData[i*32 : i*32+4])
			fBitsR := binary.LittleEndian.Uint32(pcmData[i*32+4 : i*32+8])
			fL := math.Float32frombits(fBitsL)
			fR := math.Float32frombits(fBitsR)
			if fL > 1.0 {
				fL = 1.0
			} else if fL < -1.0 {
				fL = -1.0
			}
			if fR > 1.0 {
				fR = 1.0
			} else if fR < -1.0 {
				fR = -1.0
			}
			binary.LittleEndian.PutUint16(outData[i*4:], uint16(int16(fL*32767.0)))
			binary.LittleEndian.PutUint16(outData[i*4+2:], uint16(int16(fR*32767.0)))
		}
	case AudioOutParamFormatS16_8CH, AudioOutParamFormatS16_8CH_Std:
		outData = make([]byte, handle.Length*4)
		for i := range handle.Length {
			copy(outData[i*4:], pcmData[i*16:i*16+2])
			copy(outData[i*4+2:], pcmData[i*16+2:i*16+4])
		}
	default:
		panic("unsupported audio output format")
	}

	handle.PipeWriter.Write(outData)
	return int32(handle.Length), int32(info.Channels)
}

// 0x0000000000000B80
// __int64 __fastcall sceAudioOutOutput(int a1, __int64 a2)
func libSceAudioOut_sceAudioOutOutput(handleId uint32, ptr uint64) uintptr {
	frames, channels := handleAudioOutput(handleId, ptr)
	if frames < 0 {
		return uintptr(frames)
	}

	return uintptr(frames * channels)
}

// 0x00000000000016D0
// __int64 __fastcall sceAudioOutOutputs(_DWORD *, unsigned int, double, double)
func libSceAudioOut_sceAudioOutOutputs(paramPtr uint64, num uint32) uintptr {
	if num == 0 || paramPtr == 0 {
		return 0
	}
	params := unsafe.Slice((*AudioOutOutputParam)(unsafe.Pointer(uintptr(paramPtr))), num)
	framesSent := int32(0)
	for i := range num {
		frames, _ := handleAudioOutput(uint32(params[i].Handle), params[i].Ptr)
		if frames > 0 && i == 0 {
			framesSent = frames
		}
	}

	return uintptr(framesSent)
}
