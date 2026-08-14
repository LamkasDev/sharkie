package mouse

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/mouse"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000001130
// __int64 __fastcall sceMouseRead(int, __int64, int, __m128 _XMM0)
func libSceMouse_sceMouseRead(handleId uint32, data *MouseData, count uintptr) uintptr {
	if data == nil {
		logger.Printf("%-132s %s failed due to invalid data pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceMouseRead"),
		)
		return 0x80DF0001
	}

	*data = MouseData{Connected: false}
	return 1
}
