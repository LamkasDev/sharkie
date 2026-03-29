package lib

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x000000000000B950
// __int64 __fastcall sceVideoOutSubmitEopFlip(int a1, unsigned int a2, unsigned int a3, __int64 a4, __int64 a5)
func libSceVideoOut_sceVideoOutSubmitEopFlip(handle, bufId, mode, flipArg, _ uintptr) uintptr {
	logger.Printf("%-132s %s called with %s (bufId=%s, mode=%s, flipArg=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceVideoOutSubmitEopFlip"),
		color.Yellow.Sprintf("0x%X", handle),
		color.Yellow.Sprintf("0x%X", bufId),
		color.Yellow.Sprintf("0x%X", mode),
		color.Yellow.Sprintf("0x%X", flipArg),
	)

	return 0
}
