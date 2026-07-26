package system_service

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000004140                                                                                                                                                                                                                                                                                  │
// __int64 __fastcall sceSystemServiceParamGetInt(int, int *)                                                                                                                                                                                                                                          │
func libSceSystemService_sceSystemServiceParamGetInt(paramId, valuePtr uintptr) uintptr {
	if valuePtr == 0 {
		logger.Printf("%-132s %s failed due to invalid event pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSystemServiceReceiveEvent"),
		)
		return 0x80A10003
	}

	var value uint32
	switch paramId {
	case 1: // Language.
		value = 1 // English.
	case 2: // Date format.
		value = 1 // DDMMYYYY.
	case 3: // Time format.
		value = 1 // 24 hour.
	case 4: // Timezone.
		value = 120
	case 5: // Summertime.
		value = 1
	case 7: // Game parental level.
		value = 0 // Off.
	case 1000: // Enter button assignment.
		value = 0 // Circle.
	default:
		value = 0
	}
	WriteResult(valuePtr, value)

	return 0
}
