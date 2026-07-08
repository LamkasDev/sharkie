package user_service

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/user"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x00000000000027F0
// __int64 __fastcall sceUserServiceInitialize(unsigned int *)
func SceUserServiceInitialize(param uintptr) uintptr {
	logger.Printf("sceUserServiceInitialize called (param=%x)\n", param)
	return 0
}

// 0x0000000000003400
// __int64 sceUserServiceGetInitialUser()
func SceUserServiceGetInitialUser(userIdPtr uintptr) uintptr {
	if userIdPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid user id pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceUserServiceGetInitialUser"),
		)
		return 0x80960005
	}

	user := GlobalUserManager.GetInitialUser()
	userIdSlice := unsafe.Slice((*int32)(unsafe.Pointer(userIdPtr)), 1)
	userIdSlice[0] = user.UserId

	logger.Printf("%-132s %s returned %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceUserServiceGetInitialUser"),
		color.Yellow.Sprintf("0x%X", user.UserId),
	)
	return 0
}

// 0x00000000000044F0
// __int64 __fastcall sceUserServiceGetUserName(__int64, __int64, unsigned __int64)
func SceUserServiceGetUserName(userId int32, userNamePtr, size uintptr) uintptr {
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

	user := GlobalUserManager.GetUser(userId)
	if user == nil {
		logger.Printf("%-132s %s failed due to unknown user id %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceUserServiceGetUserName"),
			color.Yellow.Sprintf("0x%X", userId),
		)
		user = NewDefaultUser()
	}
	CString(Cstring(userNamePtr), user.UserName)

	logger.Printf("%-132s %s returned %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceUserServiceGetUserName"),
		color.Green.Sprint(user.UserName),
	)
	return 0
}

// 0x0000000000003150
// __int64 __fastcall sceUserServiceGetEvent(_QWORD *)
func SceUserServiceGetEvent(eventPtr uintptr) uintptr {
	logger.Printf("sceUserServiceGetEvent called (eventPtr=%x)\n", eventPtr)
	return 0
}
