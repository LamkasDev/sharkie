package system_service

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterSystemServiceStubs() {
	elf.RegisterStub("libSceSystemService", "sceSystemServiceHideSplashScreen", libSceSystemService_stub)
	// elf.RegisterStub("libSceSystemService", "sceSystemServiceReceiveEvent", libSceSystemService_stub)
	elf.RegisterStub("libSceSystemService", "sceSystemServiceParamGetInt", libSceSystemService_sceSystemServiceParamGetInt)
	// elf.RegisterStub("libSceSystemService", "sceSystemServiceGetStatus", libSceSystemService_stub)
}

func libSceSystemService_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}

// 0x0000000000004140                                                                                                                                                                                                                                                                                  │
// __int64 __fastcall sceSystemServiceParamGetInt(int, int *)                                                                                                                                                                                                                                          │
func libSceSystemService_sceSystemServiceParamGetInt(paramId, valuePtr uintptr) uintptr {
	if valuePtr == 0 {
		return 0x80A10006 // ORBIS_SYSTEM_SERVICE_ERROR_PARAMETER
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
		value = 1 // Sun is the devil.
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
