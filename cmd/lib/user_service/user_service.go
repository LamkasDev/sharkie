package user_service

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/logger"
)

func RegisterUserServiceStubs() {
	elf.RegisterStub("libSceUserService", "sceUserServiceInitialize", SceUserServiceInitialize)
	elf.RegisterStub("libSceUserService", "sceUserServiceGetInitialUser", SceUserServiceGetInitialUser)
	elf.RegisterStub("libSceUserService", "sceUserServiceGetUserName", SceUserServiceGetUserName)
	// elf.RegisterStub("libSceUserService", "sceUserServiceGetEvent", SceUserServiceGetEvent)
}

func SceUserServiceInitialize(param uintptr) uintptr {
	logger.Printf("sceUserServiceInitialize called (param=%x)\n", param)
	return 0
}

func SceUserServiceGetInitialUser(userIdPtr uintptr) uintptr {
	logger.Printf("sceUserServiceGetInitialUser called (userIdPtr=%x)\n", userIdPtr)
	return 0
}

func SceUserServiceGetUserName(userId uintptr, userNamePtr uintptr, size uintptr) uintptr {
	logger.Printf("sceUserServiceGetUserName called (userId=%x, userNamePtr=%x, size=%d)\n", userId, userNamePtr, size)
	return 0
}

func SceUserServiceGetEvent(eventPtr uintptr) uintptr {
	logger.Printf("sceUserServiceGetEvent called (eventPtr=%x)\n", eventPtr)
	return 0
}
