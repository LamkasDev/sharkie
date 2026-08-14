package np_common

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/np_manager"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x000000000001EF30
// __int64 __fastcall sceNpCmpNpId(__int64, __int64)
func libSceNpCommon_sceNpCmpNpId(npId1, npId2 *NpId) uintptr {
	if npId1 == nil || npId2 == nil {
		logger.Printf("%-132s %s failed due to invalid pointers.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceNpCmpNpId"),
		)
		return 0x80550003
	}
	dataMatch := GoString(Cstring(&npId1.Handle.Data[0])) == GoString(Cstring(&npId2.Handle.Data[0]))
	if !dataMatch || npId1.Opt != npId2.Opt || npId1.Reserved != npId2.Reserved {
		logger.Printf("%-132s %s failed matching ids.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceNpCmpNpId"),
		)
		return 0x80550609
	}

	return 0
}
