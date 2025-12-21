package lib

import (
	"fmt"
	"os"

	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/gookit/color"
)

func RegisterStubs() {
	elf.RegisterStub("", "__sharkie_generic_stub", GenericStub)

	RegisterKernelStubs()
	RegisterSceLibcInternalStubs()
	RegisterLibcStubs()
}

func Abort() uintptr {
	fmt.Printf(
		"%-120s aborted :c\n",
		emu.GlobalModuleManager.GetCallSiteText(),
	)
	os.Exit(0)

	return 0
}

func GenericStub() uintptr {
	fmt.Printf(
		"%-120s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}

func StackChkFail() uintptr {
	color.Red.Sprint("Stack Corruption Detected!\n")
	os.Exit(1)

	return 0
}
