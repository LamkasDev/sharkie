package ime

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/ime"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x00000000000005B0
// __int64 __fastcall sceImeOpen(unsigned int *, unsigned int *, __m128 _XMM0, __m128 _XMM1)
func libSceIme_sceImeOpen(param *ImeParam) uintptr {
	GlobalImeDevice.Mutex.Lock()
	defer GlobalImeDevice.Mutex.Unlock()
	if GlobalImeDevice.IsOpen {
		logger.Printf("%-132s %s failed due to already open device.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceImeOpen"),
		)
		return 0x80BC0001
	}

	// Setup device.
	GlobalImeDevice.IsOpen = true
	GlobalImeDevice.Param = *param
	GlobalImeDevice.InputText = ""
	GlobalImeDevice.Events = nil

	// TODO: rest of the checks, add ime rect.
	event := ImeEvent{Id: ImeEventIdOpen}
	GlobalImeDevice.Events = append(GlobalImeDevice.Events, event)

	logger.Printf("%-132s %s opened ime device.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceImeOpen"),
	)
	return 0
}

// 0x0000000000000FB0
// __int64 __fastcall sceImeUpdate(__int64, __int64, __m128)
func libSceIme_sceImeUpdate(eventHandlerAddress uintptr) uintptr {
	if !GlobalImeDevice.IsOpen && !GlobalImeDevice.IsKeyboardOpen {
		logger.Printf("%-132s %s failed due to unopened device.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceImeUpdate"),
		)
		return 0x80BC0002
	}
	GlobalImeDevice.Mutex.Lock()
	defer GlobalImeDevice.Mutex.Unlock()
	for len(GlobalImeDevice.Events) > 0 {
		event := GlobalImeDevice.Events[0]
		GlobalImeDevice.Events = GlobalImeDevice.Events[1:]

		var arg uintptr
		isKeyboardEvent := event.Id == ImeEventIdKeyboardOpen || event.Id == ImeEventIdKeyboardKeycodeDown || event.Id == ImeEventIdKeyboardKeycodeUp
		handler := eventHandlerAddress
		if isKeyboardEvent {
			if handler == 0 {
				handler = GlobalImeDevice.KeyboardParam.EventHandlerAddress
			}
			arg = GlobalImeDevice.KeyboardParam.Arg
		} else {
			if handler == 0 {
				handler = GlobalImeDevice.Param.EventHandlerAddress
			}
			arg = GlobalImeDevice.Param.ArgPtr
		}

		if handler != 0 {
			thread := emu.GetCurrentThread()
			thread.CallAndWaitFromStub(handler, arg, uintptr(unsafe.Pointer(&event)))
		}
	}

	return 0
}

// 0x0000000000001AA0
// __int64 sceImeClose()
func libSceIme_sceImeClose() uintptr {
	GlobalImeDevice.Mutex.Lock()
	defer GlobalImeDevice.Mutex.Unlock()
	GlobalImeDevice.IsOpen = false

	logger.Printf("%-132s %s closed ime device.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceImeClose"),
	)
	return 0
}
