package user_service

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/user"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000003400
// __int64 sceUserServiceGetInitialUser()
func libSceUserService_sceUserServiceGetInitialUser(userIdPtr uintptr) uintptr {
	if userIdPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid user id pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceUserServiceGetInitialUser"),
		)
		return 0x80960005
	}
	user := GlobalUserManager.GetInitialUser()
	userId := (*UserId)(unsafe.Pointer(userIdPtr))
	*userId = user.UserId

	logger.Printf("%-132s %s returned %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceUserServiceGetInitialUser"),
		color.Yellow.Sprintf("0x%X", user.UserId),
	)
	return 0
}

// 0x00000000000044F0
// __int64 __fastcall sceUserServiceGetUserName(__int64, __int64, unsigned __int64)
func libSceUserService_sceUserServiceGetUserName(userId int32, userNamePtr, size uintptr) uintptr {
	if userId == -1 {
		logger.Printf("%-132s %s failed due to invalid user id.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceUserServiceGetUserName"),
		)
		return 0x80960005
	}
	if userNamePtr == 0 {
		logger.Printf("%-132s %s failed due to invalid user name pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceUserServiceGetUserName"),
		)
		return 0x80960005
	}

	user := GlobalUserManager.GetUser(UserId(userId))
	if user == nil {
		logger.Printf("%-132s %s failed due to unknown user id %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceUserServiceGetUserName"),
			color.Yellow.Sprintf("0x%X", userId),
		)
		user = NewDefaultUser()
	}
	CString(Cstring(userNamePtr), user.UserName)
	// TODO: size check

	logger.Printf("%-132s %s returned %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceUserServiceGetUserName"),
		color.Blue.Sprint(user.UserName),
	)
	return 0
}

// 0x00000000000037B0
// __int64 sceUserServiceGetUserColor
func libSceUserService_sceUserServiceGetUserColor(userId int32, userColorPtr uintptr) uintptr {
	if userId == -1 {
		logger.Printf("%-132s %s failed due to invalid user id.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceUserServiceGetUserColor"),
		)
		return 0x80960005
	}
	if userColorPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid user name pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceUserServiceGetUserColor"),
		)
		return 0x80960005
	}

	user := GlobalUserManager.GetUser(UserId(userId))
	if user == nil {
		logger.Printf("%-132s %s failed due to unknown user id %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceUserServiceGetUserColor"),
			color.Yellow.Sprintf("0x%X", userId),
		)
		user = NewDefaultUser()
	}
	userColor := (*UserColor)(unsafe.Pointer(userColorPtr))
	*userColor = user.UserColor

	logger.Printf("%-132s %s returned %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceUserServiceGetUserColor"),
		color.Green.Sprint(user.UserColor),
	)
	return 0
}

// 0x0000000000002C00
// __int64 __fastcall sceUserServiceGetLoginUserIdList(_DWORD *)
func libSceUserService_sceUserServiceGetLoginUserIdList(userIdListPtr uintptr) uintptr {
	if userIdListPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid user id list pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceUserServiceGetLoginUserIdList"),
		)
		return 0x80960005
	}
	userIdList := (*LoginUserIdList)(unsafe.Pointer(userIdListPtr))
	for i := range userIdList.UserIds {
		userIdList.UserIds[i] = UserIdInvalid
	}
	loggedInUsers := GlobalUserManager.GetLoggedInUsers()
	for i, user := range loggedInUsers {
		userIdList.UserIds[i] = user.UserId
	}

	logger.Printf("%-132s %s returned %s users.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceUserServiceGetLoginUserIdList"),
		color.Green.Sprint(len(loggedInUsers)),
	)
	return 0
}
