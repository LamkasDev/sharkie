package ime_dialog

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/ime"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000000D00
// __int64 __fastcall sceImeDialogInit(int *, __int64, __int64)
func libSceImeDialog_sceImeDialogInit(param *ImeDialogParam, extended uintptr) uintptr {
	GlobalImeDevice.Mutex.Lock()
	defer GlobalImeDevice.Mutex.Unlock()
	if GlobalImeDevice.IsDialogOpen {
		logger.Printf("%-132s %s failed due to already open dialog.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceImeOpen"),
		)
		return 0x80BC0001
	}
	GlobalImeDevice.IsDialogOpen = true
	GlobalImeDevice.DialogParam = *param
	GlobalImeDevice.DialogStatus = ImeDialogStatusRunning
	GlobalImeDevice.InputText = ""

	logger.Printf("%-132s %s opened ime dialog.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceImeDialogInit"),
	)
	return 0
}

// 0x0000000000002700
// __int64 sceImeDialogTerm()
func libSceImeDialog_sceImeDialogTerm() uintptr {
	GlobalImeDevice.Mutex.Lock()
	defer GlobalImeDevice.Mutex.Unlock()
	GlobalImeDevice.IsDialogOpen = false
	GlobalImeDevice.DialogStatus = ImeDialogStatusNone

	logger.Printf("%-132s %s closed ime dialog.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceImeDialogTerm"),
	)
	return 0
}

// 0x0000000000002660
// __int64 sceImeDialogAbort()
func libSceImeDialog_sceImeDialogAbort() uintptr {
	GlobalImeDevice.Mutex.Lock()
	defer GlobalImeDevice.Mutex.Unlock()
	GlobalImeDevice.DialogStatus = ImeDialogStatusFinished
	GlobalImeDevice.DialogResult = ImeDialogResult{EndStatus: ImeDialogEndStatusAborted}

	logger.Printf("%-132s %s aborted ime dialog.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceImeDialogAbort"),
	)
	return 0
}
