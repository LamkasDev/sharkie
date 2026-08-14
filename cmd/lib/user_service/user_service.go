package user_service

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
)

func RegisterUserServiceStubs() {
	// Setup functions.
	elf.RegisterStub("libSceUserService", "sceUserServiceInitialize", libSceUserService_sceUserServiceInitialize)
	elf.RegisterStub("libSceUserService", "sceUserServiceInitialize2", libSceUserService_sceUserServiceInitialize)

	// Event functions.
	elf.RegisterStub("libSceUserService", "sceUserServiceGetEvent", libSceUserService_sceUserServiceGetEvent)

	// User functions.
	elf.RegisterStub("libSceUserService", "sceUserServiceGetInitialUser", libSceUserService_sceUserServiceGetInitialUser)
	elf.RegisterStub("libSceUserService", "sceUserServiceGetUserName", libSceUserService_sceUserServiceGetUserName)
	elf.RegisterStub("libSceUserService", "sceUserServiceGetUserColor", libSceUserService_sceUserServiceGetUserColor)
	elf.RegisterStub("libSceUserService", "sceUserServiceGetLoginUserIdList", libSceUserService_sceUserServiceGetLoginUserIdList)
}
