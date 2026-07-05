package lib

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	. "github.com/LamkasDev/sharkie/cmd/structs"
)

func RegisterSystemServiceStubs() {
	elf.RegisterStub("libSceSystemService", "sceSystemServiceParamGetInt", libSceSystemService_sceSystemServiceParamGetInt)
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
		value = 16 // English.
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
