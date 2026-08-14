package system_service

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/system_service"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000000E40
// __int64 __fastcall sceSystemServiceGetStatus(__int64 _RDI, __m128 _XMM0)
func libSceSystemService_sceSystemServiceGetStatus(status *SystemServiceStatus) uintptr {
	if status == nil {
		logger.Printf("%-132s %s failed due to invalid status pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSystemServiceGetStatus"),
		)
		return 0x80A10003
	}
	status.EventNumber = int32(GlobalSystemService.EventQueue.Len())
	status.IsSystemUiOverlaid = false
	status.IsInBackgroundExecution = false
	status.IsCpuMode7CpuNormal = true
	status.IsGameLiveStreamingOnAir = false
	status.IsOutOfVrPlayArea = false

	return 0
}

// 0x0000000000001FC0
// __int64 __fastcall sceSystemServiceGetDisplaySafeAreaInfo(_DWORD *, __m128 _XMM0)
func libSceSystemService_sceSystemServiceGetDisplaySafeAreaInfo(info *SystemServiceDisplaySafeAreaInfo) uintptr {
	if info == nil {
		logger.Printf("%-132s %s failed due to invalid info pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceSystemServiceGetDisplaySafeAreaInfo"),
		)
		return 0x80A10003
	}
	info.Ratio = 1.0

	return 0
}
