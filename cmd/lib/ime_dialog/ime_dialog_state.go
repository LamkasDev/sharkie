package ime_dialog

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/ime"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000002370
// __int64 __fastcall sceImeDialogGetStatus(__m128 _XMM0)
func libSceImeDialog_sceImeDialogGetStatus() uintptr {
	GlobalImeDevice.Mutex.Lock()
	defer GlobalImeDevice.Mutex.Unlock()

	return uintptr(GlobalImeDevice.DialogStatus)
}

// 0x0000000000002680
// __int64 __fastcall sceImeDialogGetResult(_BYTE *)
func libSceImeDialog_sceImeDialogGetResult(result *ImeDialogResult) uintptr {
	if result == nil {
		logger.Printf("%-132s %s failed due to invalid result pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceImeDialogGetResult"),
		)
		return 0x80BC0031
	}
	GlobalImeDevice.Mutex.Lock()
	defer GlobalImeDevice.Mutex.Unlock()
	if GlobalImeDevice.DialogStatus == ImeDialogStatusNone {
		return 0x80BC0107 // DIALOG_NOT_IN_USE
	}
	if GlobalImeDevice.DialogStatus == ImeDialogStatusRunning {
		return 0x80BC0106 // DIALOG_NOT_FINISHED
	}
	if GlobalImeDevice.DialogStatus == ImeDialogStatusFinished {
		*result = GlobalImeDevice.DialogResult
	}

	return 0
}
