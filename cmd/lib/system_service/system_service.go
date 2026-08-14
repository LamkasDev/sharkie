package system_service

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterSystemServiceStubs() {
	elf.RegisterStub("libSceSystemService", "sceSystemServiceHideSplashScreen", libSceSystemService_stub)
	elf.RegisterStub("libSceSystemService", "sceSystemServiceReportAbnormalTermination", libSceSystemService_stub)

	// Status functions.
	elf.RegisterStub("libSceSystemService", "sceSystemServiceGetStatus", libSceSystemService_sceSystemServiceGetStatus)
	elf.RegisterStub("libSceSystemService", "sceSystemServiceGetDisplaySafeAreaInfo", libSceSystemService_sceSystemServiceGetDisplaySafeAreaInfo)

	// Event functions.
	elf.RegisterStub("libSceSystemService", "sceSystemServiceReceiveEvent", libSceSystemService_sceSystemServiceReceiveEvent)

	// Parameter functions.
	elf.RegisterStub("libSceSystemService", "sceSystemServiceParamGetInt", libSceSystemService_sceSystemServiceParamGetInt)
}

func libSceSystemService_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
