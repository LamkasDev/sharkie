package pad

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterPadStubs() {
	// Setup functions.
	elf.RegisterStub("libScePad", "scePadInit", libScePad_scePadInit)
	elf.RegisterStub("libScePad", "scePadOpen", libScePad_scePadOpen)
	elf.RegisterStub("libScePad", "scePadGetControllerInformation", libScePad_scePadGetControllerInformation)
	elf.RegisterStub("libScePad", "scePadRead", libScePad_scePadRead)
	elf.RegisterStub("libScePad", "scePadReadState", libScePad_scePadReadState)
	elf.RegisterStub("libScePad", "scePadClose", libScePad_stub)
	elf.RegisterStub("libScePad", "scePadResetLightBar", libScePad_stub)
	elf.RegisterStub("libScePad", "scePadSetLightBar", libScePad_stub)
	elf.RegisterStub("libScePad", "scePadSetMotionSensorState", libScePad_stub)
	elf.RegisterStub("libScePad", "scePadSetVibration", libScePad_stub)
}

func libScePad_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
