package np_web_api

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/np_web_api"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000002F40
// __int64 __fastcall sceNpWebApiInitialize(__int64, __int64)
func libSceNpWebApi_sceNpWebApiInitialize(httpContextId uint32, poolSize uint64) uintptr {
	context := GlobalWebApiManager.CreateContext()
	context.HttpContextId = httpContextId

	logger.Printf("%-132s %s created context %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceNpWebApiInitialize"),
		color.Yellow.Sprintf("0x%X", context.Id),
	)
	return uintptr(context.Id)
}
