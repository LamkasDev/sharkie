package np_trophy

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/np_trophy"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/user"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x00000000000012A0
// __int64 __fastcall sceNpTrophyCreateContext(unsigned int *, unsigned int, unsigned int, __int64)
func libSceNpTrophy_sceNpTrophyCreateContext(contextId *uint32, userId UserId, serviceLabel uint32, options uint64) uintptr {
	if contextId == nil || options != 0 {
		logger.Printf("%-132s %s failed due to invalid context id pointer or options.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceNpTrophyCreateContext"),
		)
		return 0x80551604
	}
	context := GlobalTrophyManager.CreateContext()
	context.ServiceLabel = serviceLabel
	context.UserId = userId
	*contextId = context.Id

	logger.Printf("%-132s %s created context %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceNpTrophyCreateContext"),
		color.Yellow.Sprintf("0x%X", context.Id),
	)
	return 0
}

// 0x0000000000001590
// __int64 __fastcall sceNpTrophyCreateHandle(__int64)
func libSceNpTrophy_sceNpTrophyCreateHandle(handleId *uint32) uintptr {
	if handleId == nil {
		logger.Printf("%-132s %s failed due to invalid handle id pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceNpTrophyCreateHandle"),
		)
		return 0x80551604
	}
	handle := GlobalTrophyManager.CreateHandle()
	*handleId = handle.Id

	logger.Printf("%-132s %s created handle %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceNpTrophyCreateHandle"),
		color.Yellow.Sprintf("0x%X", handle.Id),
	)
	return 0
}

// 0x00000000000028F0
// __int64 __fastcall sceNpTrophyRegisterContext(unsigned int, unsigned int, __int64, __m128 _XMM0)
func libSceNpTrophy_sceNpTrophyRegisterContext(contextId, handleId uint32, options uint64) uintptr {
	if options != 0 {
		logger.Printf("%-132s %s failed due to invalid options.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceNpTrophyRegisterContext"),
		)
		return 0x80551604
	}
	context := GlobalTrophyManager.GetContext(contextId)
	if context == nil {
		logger.Printf("%-132s %s failed due to invalid context id.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceNpTrophyRegisterContext"),
		)
		return 0x80551609
	}
	handle := GlobalTrophyManager.GetHandle(handleId)
	if handle == nil {
		logger.Printf("%-132s %s failed due to invalid handle id.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceNpTrophyRegisterContext"),
		)
		return 0x80551608
	}

	return 0
}
