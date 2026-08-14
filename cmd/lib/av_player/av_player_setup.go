package av_player

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/av_player"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000001620
// __int64 __fastcall sceAvPlayerInit(__int64)
func libSceAvPlayer_sceAvPlayerInit(initData *AvPlayerInitData) uintptr {
	if initData == nil {
		logger.Printf("%-132s %s failed due to invalid init data.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceAvPlayerInit"),
		)
		return 0x806A0001
	}
	handle := GlobalAvPlayerEngine.CreateHandle()

	logger.Printf("%-132s %s returned %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceAvPlayerInit"),
		color.Yellow.Sprintf("0x%X", handle.Id),
	)
	return uintptr(handle.Id)
}

// 0x00000000000025A0
// __int64 __fastcall sceAvPlayerPostInit(__int64, unsigned int *_RSI)
func libSceAvPlayer_sceAvPlayerPostInit(handleId uint32, postInitData *AvPlayerPostInitData) uintptr {
	handle := GlobalAvPlayerEngine.GetHandle(handleId)
	if handle == nil || postInitData == nil {
		logger.Printf("%-132s %s failed due to invalid handle or post init data pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceAvPlayerPostInit"),
		)
		return 0x806A0001
	}

	return 0
}
