package system_service

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/system_service"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000000A90
// __int64 __fastcall sceSystemServiceReceiveEvent(_DWORD *)
func libSceSystemService_sceSystemServiceReceiveEvent(event *SystemServiceEvent) uintptr {
	if event == nil {
		logger.Printf("%-132s %s failed due to invalid event pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSystemServiceReceiveEvent"),
		)
		return 0x80A10003
	}
	nextEvent := GlobalSystemService.EventQueue.Next()
	if nextEvent == nil {
		logger.Printf("%-132s %s failed due to no event.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSystemServiceReceiveEvent"),
		)
		return 0x80A10004
	}
	*event = nextEvent.(SystemServiceEvent)

	return 0
}
