package ime

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/ime"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/user"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000004FD0
// __int64 __fastcall sceImeKeyboardOpen(int, int *, __m128 _XMM0)
func libSceIme_sceImeKeyboardOpen(userId UserId, param *ImeKeyboardParam) uintptr {
	if param == nil {
		logger.Printf("%-132s %s failed due to invalid param pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceImeKeyboardOpen"),
		)
		return 0x80BC0031
	}
	if param.EventHandlerAddress == 0 {
		logger.Printf("%-132s %s failed due to invalid event handler.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceImeKeyboardOpen"),
		)
		return 0x80BC0022
	}
	GlobalImeDevice.Mutex.Lock()
	defer GlobalImeDevice.Mutex.Unlock()
	if GlobalImeDevice.IsKeyboardOpen {
		logger.Printf("%-132s %s failed due already open keyboard.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceImeKeyboardOpen"),
		)
		return 0x80BC0001
	}

	// Setup device.
	GlobalImeDevice.IsKeyboardOpen = true
	GlobalImeDevice.KeyboardParam = *param
	if userId == 254 || userId == 255 || userId < 0 {
		userId = 1000
	}
	GlobalImeDevice.KeyboardUserId = int32(userId)

	// The game expects resource ID array in the param.
	// We can manually set the UserId and ResourceId[0].
	// Offset 0: UserId (4 bytes), Offset 4: ResourceId[0] (4 bytes).
	event := ImeEvent{Id: ImeEventIdKeyboardOpen}
	event.Param.Data[0] = byte(userId)
	event.Param.Data[1] = byte(userId >> 8)
	event.Param.Data[2] = byte(userId >> 16)
	event.Param.Data[3] = byte(userId >> 24)
	event.Param.Data[4] = 1
	GlobalImeDevice.Events = append(GlobalImeDevice.Events, event)

	logger.Printf("%-132s %s opened ime keyboard (userId=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceImeKeyboardOpen"),
		color.Yellow.Sprintf("0x%X", userId),
	)
	return 0
}

// 0x0000000000005B40
// __int64 __fastcall sceImeKeyboardClose(int, __m128 _XMM0, __m128 _XMM1)
func libSceIme_sceImeKeyboardClose(userId UserId) uintptr {
	GlobalImeDevice.Mutex.Lock()
	defer GlobalImeDevice.Mutex.Unlock()
	GlobalImeDevice.IsKeyboardOpen = false

	logger.Printf("%-132s %s closed ime keyboard (userId=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceImeKeyboardClose"),
		color.Yellow.Sprintf("0x%X", userId),
	)
	return 0
}

// 0x0000000000006100
// __int64 __fastcall sceImeKeyboardUpdate(__int64, __int64, __int64, __m128 _XMM0)
func libSceIme_sceImeKeyboardUpdate() uintptr {
	return libSceIme_sceImeUpdate(0)
}

// 0x0000000000005E40
// __int64 __fastcall sceImeKeyboardGetResourceId(int, __int64, __m128 _XMM0)
func libSceIme_sceImeKeyboardGetResourceId(userId UserId, resourceIdArray *ImeKeyboardResourceIdArray) uintptr {
	if resourceIdArray == nil {
		logger.Printf("%-132s %s failed due to invalid resource id array pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceImeKeyboardGetResourceId"),
		)
		return 0x80BC0031
	}

	resourceIdArray.UserId = userId
	for i := range resourceIdArray.ResourceIds {
		resourceIdArray.ResourceIds[i] = 0
	}

	logger.Printf("%-132s %s returned resource ids (userId=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceImeKeyboardGetResourceId"),
		color.Yellow.Sprintf("0x%X", userId),
	)
	return 0
}
