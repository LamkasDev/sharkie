package kernel

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000002C70
// void *_error()
func libKernel___error() uintptr {
	address := emu.GetErrnoAddress()
	if logger.LogErrorRet {
		logger.Printf("%-132s %s returned %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("_error"),
			color.Red.Sprintf("0x%X", address),
		)
	}

	return address
}

// 0x0000000000014E50
// __int64 __fastcall sceKernelError(int)
func libKernel_sceKernelError(err uintptr) uintptr {
	if err != 0 {
		err -= SonyErrorOffset
	}
	if logger.LogErrorRet {
		logger.Printf("%-132s %s returned %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelError"),
			color.Red.Sprintf("0x%X", err),
		)
	}

	return err
}
