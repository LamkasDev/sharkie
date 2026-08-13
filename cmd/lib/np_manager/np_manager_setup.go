package np_manager

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/np_manager"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/user"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000017550
// __int64 __fastcall sceNpGetState(unsigned int, _DWORD *)
func libSceNpManager_sceNpGetState(userId UserId, statePtr *NpState) uintptr {
	if statePtr == nil {
		logger.Printf("%-132s %s failed due to invalid state pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceNpGetState"),
		)
		return 0x80550003
	}
	if userId == UserIdInvalid {
		logger.Printf("%-132s %s failed due to invalid user id.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceNpGetState"),
		)
		return 0x80550003
	}
	*statePtr = NpStateSignedOut

	if logger.LogMisc {
		logger.Printf("%-132s %s returned login state.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceNpGetState"),
		)
	}
	return 0
}
