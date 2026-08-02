package net_ctl

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/net"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000003380
// __int64 __fastcall sceNetCtlGetInfo(unsigned int, _BYTE *)
func libSceNetCtl_sceNetCtlGetInfo() uintptr {
	return 0
}

// 0x00000000000031A0
// __int64 __fastcall sceNetCtlGetResult(unsigned int, __int64)
func libSceNetCtl_sceNetCtlGetResult(eventType uint32, errorCode *uint32) uintptr {
	if errorCode == nil {
		logger.Printf("%-132s %s failed due to invalid error code pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceNetCtlGetResult"),
		)
		return SCE_NET_CTL_ERROR_INVALID_ADDR
	}
	*errorCode = 0
	return 0
}

// 0x0000000000003200
// __int64 __fastcall sceNetCtlGetState(__int64)
func libSceNetCtl_sceNetCtlGetState(state *uint32) uintptr {
	if state == nil {
		logger.Printf("%-132s %s failed due to invalid state pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceNetCtlGetState"),
		)
		return SCE_NET_CTL_ERROR_INVALID_ADDR
	}
	*state = 0
	return 0
}

// 0x0000000000001DF0
// __int64 __fastcall sceNetCtlRegisterCallback(__int64, __int64, unsigned int *)
func libSceNetCtl_sceNetCtlRegisterCallback() uintptr {
	return 0
}

// 0x0000000000002430
// __int64 sceNetCtlCheckCallback()
func libSceNetCtl_sceNetCtlCheckCallback() uintptr {
	return 0
}
