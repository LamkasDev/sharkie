package user_service

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
)

func RegisterUserServiceStubs() {
	elf.RegisterStub("libSceUserService", "sceUserServiceInitialize", SceUserServiceInitialize)
	elf.RegisterStub("libSceUserService", "sceUserServiceGetInitialUser", SceUserServiceGetInitialUser)
	elf.RegisterStub("libSceUserService", "sceUserServiceGetUserName", SceUserServiceGetUserName)
	// elf.RegisterStub("libSceUserService", "sceUserServiceGetEvent", SceUserServiceGetEvent)
}
