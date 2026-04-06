package lib

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs/audio"
	"github.com/gookit/color"
)

// 0x0000000000000420
// __int64 __fastcall sceAudioOutOpen(unsigned int, int, unsigned int, unsigned int, unsigned int, unsigned int)
func libSceAudioOut_sceAudioOutOpen() uintptr {
	handle := &AudioOutHandle{
		Id: GlobalAudioEngine.NextHandle,
	}
	GlobalAudioEngine.Handles[handle.Id] = handle
	GlobalAudioEngine.NextHandle++

	logger.Printf("%-132s %s returned %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceAudioOutOpen"),
		color.Yellow.Sprintf("0x%X", handle.Id),
	)
	return uintptr(handle.Id)
}

// 0x0000000000000B80
// __int64 __fastcall sceAudioOutOutput(int a1, __int64 a2)
func libSceAudioOut_sceAudioOutOutput() uintptr {
	return 0
}

// 0x0000000000002D90
// __int64 __fastcall sceAudioOutGetPortState(int a1, unsigned __int8 *a2)
func libSceAudioOut_sceAudioOutGetPortState() uintptr {
	return 0
}

// 0x00000000000010B0
// __int64 __fastcall sceAudioOutSetVolume(int, int _ESI, __int64, __m128 _XMM0, __m128 _XMM1, __m128 _XMM2, __m128 _XMM3, __m128 _XMM4, __m128 _XMM5, __m128 _XMM6)
func libSceAudioOut_sceAudioOutSetVolume() uintptr {
	return 0
}
