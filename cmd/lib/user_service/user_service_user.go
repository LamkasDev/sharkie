package user_service

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/user"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000003400
// __int64 sceUserServiceGetInitialUser()
func libSceUserService_sceUserServiceGetInitialUser(userId *UserId) uintptr {
	if userId == nil {
		logger.Printf("%-132s %s failed due to invalid user id pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceUserServiceGetInitialUser"),
		)
		return 0x80960005
	}
	user := GlobalUserManager.GetInitialUser()
	*userId = user.UserId

	logger.Printf("%-132s %s returned %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceUserServiceGetInitialUser"),
		color.Green.Sprint(user.UserId),
	)
	return 0
}

// 0x00000000000044F0
// __int64 __fastcall sceUserServiceGetUserName(__int64, __int64, unsigned __int64)
func libSceUserService_sceUserServiceGetUserName(userId UserId, userNamePtr Cstring, size uintptr) uintptr {
	if userId == UserIdInvalid {
		logger.Printf("%-132s %s failed due to invalid user id.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceUserServiceGetUserName"),
		)
		return 0x80960005
	}
	if userNamePtr == nil {
		logger.Printf("%-132s %s failed due to invalid user name pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceUserServiceGetUserName"),
		)
		return 0x80960005
	}

	user := GlobalUserManager.GetUser(userId)
	if user == nil {
		logger.Printf("%-132s %s failed due to unknown user id %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceUserServiceGetUserName"),
			color.Yellow.Sprintf("0x%X", userId),
		)
		user = NewDefaultUser()
	}
	CString(userNamePtr, user.UserName)
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
func libSceUserService_sceUserServiceGetUserColor(userId UserId, userColor *UserColor) uintptr {
	if userId == UserIdInvalid {
		logger.Printf("%-132s %s failed due to invalid user id.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceUserServiceGetUserColor"),
		)
		return 0x80960005
	}
	if userColor == nil {
		logger.Printf("%-132s %s failed due to invalid user name pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceUserServiceGetUserColor"),
		)
		return 0x80960005
	}

	user := GlobalUserManager.GetUser(userId)
	if user == nil {
		logger.Printf("%-132s %s failed due to unknown user id %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceUserServiceGetUserColor"),
			color.Green.Sprint(userId),
		)
		user = NewDefaultUser()
	}
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
func libSceUserService_sceUserServiceGetLoginUserIdList(userIdList *LoginUserIdList) uintptr {
	if userIdList == nil {
		logger.Printf("%-132s %s failed due to invalid user id list pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceUserServiceGetLoginUserIdList"),
		)
		return 0x80960005
	}
	for i := range userIdList.UserIds {
		userIdList.UserIds[i] = UserIdInvalid
	}
	loggedInUsers := GlobalUserManager.GetLoggedInUsers()
	for i, user := range loggedInUsers {
		userIdList.UserIds[i] = user.UserId
	}

	if logger.LogMisc {
		logger.Printf("%-132s %s returned %s users.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceUserServiceGetLoginUserIdList"),
			color.Green.Sprint(len(loggedInUsers)),
		)
	}
	return 0
}
