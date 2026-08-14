package av_player

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/av_player"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000002BF0
// __int64 __fastcall sceAvPlayerAddSource(_QWORD *, __int64)
func libSceAvPlayer_sceAvPlayerAddSource(handleId uint32) uintptr {
	handle := GlobalAvPlayerEngine.GetHandle(handleId)
	if handle == nil {
		logger.Printf("%-132s %s failed due to invalid handle.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceAvPlayerAddSource"),
		)
		return 0x806A0001
	}

	return 0
}
