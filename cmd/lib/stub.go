package lib

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterStubs() {
	elf.RegisterStub("", "__sharkie_generic_stub", GenericStub)

	RegisterKernelStubs()
	RegisterSceLibcInternalStubs()
	RegisterLibcStubs()
	RegisterVideoOutStubs()
}

func Abort() uintptr {
	logger.Printf(
		"%-132s aborted :c\n",
		emu.GlobalModuleManager.GetCallSiteText(),
	)
	logger.CleanupAndExit()

	return 0
}

func GenericStub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}

func StackChkFail() uintptr {
	color.Red.Sprint("Stack Corruption Detected!\n")
	logger.CleanupAndExit()

	return 0
}
