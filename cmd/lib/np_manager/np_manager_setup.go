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

// 0x0000000000018530
// __int64 __fastcall sceNpHasSignedUp(unsigned int, bool *)
func libSceNpManager_sceNpHasSignedUp(userId UserId, signedPtr *bool) uintptr {
	if signedPtr == nil {
		logger.Printf("%-132s %s failed due to invalid state pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceNpHasSignedUp"),
		)
		return 0x80550003
	}
	*signedPtr = false
	if userId == UserIdInvalid {
		logger.Printf("%-132s %s failed due to invalid user id.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceNpHasSignedUp"),
		)
		return 0x80550003
	}
	signed := false
	*signedPtr = signed

	if logger.LogMisc {
		logger.Printf("%-132s %s returned signed up state.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceNpHasSignedUp"),
		)
	}
	return 0
}

// 0x0000000000018720
// __int64 __fastcall sceNpGetOnlineId(unsigned int, __int64, __m128 _XMM0)
func libSceNpManager_sceNpGetOnlineId(userId UserId, onlineIdPtr *NpOnlineId) uintptr {
	if userId == UserIdInvalid || onlineIdPtr == nil {
		logger.Printf("%-132s %s failed due to invalid online id pointer or user id.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceNpGetOnlineId"),
		)
		return 0x80550003
	}
	if true {
		/* logger.Printf("%-132s %s failed due to being signed out.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceNpGetOnlineId"),
		) */
		return 0x80550006
	}
	*onlineIdPtr = NpOnlineId{}

	if logger.LogMisc {
		logger.Printf("%-132s %s returned online id.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceNpGetOnlineId"),
		)
	}
	return 0
}

// 0x0000000000018850
// __int64 __fastcall sceNpGetNpId(unsigned int, __int64, double)
func libSceNpManager_sceNpGetNpId(userId UserId, idPtr *NpId) uintptr {
	if userId == UserIdInvalid || idPtr == nil {
		logger.Printf("%-132s %s failed due to invalid id pointer or user id.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceNpGetNpId"),
		)
		return 0x80550003
	}
	if true {
		logger.Printf("%-132s %s failed due to being signed out.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceNpGetNpId"),
		)
		return 0x80550006
	}
	*idPtr = NpId{}

	if logger.LogMisc {
		logger.Printf("%-132s %s returned id.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceNpGetNpId"),
		)
	}
	return 0
}
