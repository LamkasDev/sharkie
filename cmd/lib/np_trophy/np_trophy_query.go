package np_trophy

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/np_trophy"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/user"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000001DC0
// __int64 __fastcall sceNpTrophyGetGameInfo(unsigned int, unsigned int, _QWORD *, _QWORD *, double)
func libSceNpTrophy_sceNpTrophyGetGameInfo(contextId, handleId uint32, userId UserId, details *TrophyGameDetails, data *TrophyGameData) uintptr {
	context := GlobalTrophyManager.GetContext(contextId)
	if context == nil {
		logger.Printf("%-132s %s failed due to invalid context id.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceNpTrophyGetGameInfo"),
		)
		return 0x80551609
	}
	handle := GlobalTrophyManager.GetHandle(handleId)
	if handle == nil {
		logger.Printf("%-132s %s failed due to invalid handle id.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceNpTrophyGetGameInfo"),
		)
		return 0x80551608
	}
	*details = TrophyGameDetails{}
	*data = TrophyGameData{}

	logger.Printf("%-132s %s returned trophy game info (contextId=%s, handleId=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceNpTrophyGetGameInfo"),
		color.Yellow.Sprintf("0x%X", context.Id),
		color.Yellow.Sprintf("0x%X", handle.Id),
	)
	return 0
}

// 0x0000000000002530
// __int64 __fastcall sceNpTrophyGetTrophyInfo(unsigned int, unsigned int, unsigned int, _QWORD *, _QWORD *, double)
func libSceNpTrophy_sceNpTrophyGetTrophyInfo(contextId, handleId, trophyId uint32, userId UserId, details *TrophyDetails, data *TrophyData) uintptr {
	context := GlobalTrophyManager.GetContext(contextId)
	if context == nil {
		logger.Printf("%-132s %s failed due to invalid context id.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceNpTrophyGetTrophyInfo"),
		)
		return 0x80551609
	}
	handle := GlobalTrophyManager.GetHandle(handleId)
	if handle == nil {
		logger.Printf("%-132s %s failed due to invalid handle id.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceNpTrophyGetTrophyInfo"),
		)
		return 0x80551608
	}
	*details = TrophyDetails{}
	*data = TrophyData{}

	logger.Printf("%-132s %s returned trophy info for %s (contextId=%s, handleId=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceNpTrophyGetTrophyInfo"),
		color.Yellow.Sprintf("0x%X", trophyId),
		color.Yellow.Sprintf("0x%X", context.Id),
		color.Yellow.Sprintf("0x%X", handle.Id),
	)
	return 0
}
