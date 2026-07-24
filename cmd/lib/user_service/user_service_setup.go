package user_service

import (
	"github.com/LamkasDev/sharkie/cmd/logger"
)

// 0x00000000000027F0
// __int64 __fastcall sceUserServiceInitialize(unsigned int *)
func libSceUserService_sceUserServiceInitialize(param uintptr) uintptr {
	logger.Printf("sceUserServiceInitialize called (param=%x)\n", param)
	return 0
}
