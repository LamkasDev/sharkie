package user_service

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/user_service"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000003150
// __int64 __fastcall sceUserServiceGetEvent(_QWORD *)
func libSceUserService_sceUserServiceGetEvent(event *UserServiceEvent) uintptr {
	if event == nil {
		logger.Printf("%-132s %s failed due to invalid event pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceUserServiceGetEvent"),
		)
		return 0x80960005
	}
	nextEvent := GlobalUserService.EventQueue.Next()
	if nextEvent == nil {
		return 0x80960007
	}
	*event = nextEvent.(UserServiceEvent)

	return 0
}
