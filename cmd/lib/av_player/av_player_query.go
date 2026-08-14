package av_player

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/av_player"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x00000000000037F0
// char __fastcall sceAvPlayerIsActive(__int64)
func libSceAvPlayer_sceAvPlayerIsActive(handleId uint32) uintptr {
	handle := GlobalAvPlayerEngine.GetHandle(handleId)
	if handle == nil {
		logger.Printf("%-132s %s failed due to invalid handle.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceAvPlayerIsActive"),
		)
		return 0
	}

	return 1
}

// 0x0000000000003930
// __int64 __fastcall sceAvPlayerGetAudioData(__int64, __int64)
func libSceAvPlayer_sceAvPlayerGetAudioData() uintptr {
	return 0
}

// 0x0000000000003950
// __int64 __fastcall sceAvPlayerGetVideoDataEx(int *, __int64)
func libSceAvPlayer_sceAvPlayerGetVideoDataEx() uintptr {
	return 0
}
