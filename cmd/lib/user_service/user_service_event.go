package user_service

import (
	"github.com/LamkasDev/sharkie/cmd/logger"
)

// 0x0000000000003150
// __int64 __fastcall sceUserServiceGetEvent(_QWORD *)
func libSceUserService_sceUserServiceGetEvent(eventPtr uintptr) uintptr {
	logger.Printf("sceUserServiceGetEvent called (eventPtr=%x)\n", eventPtr)
	return 0
}
