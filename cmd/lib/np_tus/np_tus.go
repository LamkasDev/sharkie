package np_tus

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterNpTusStubs() {
	elf.RegisterStub("libSceNpTus", "sceNpTusGetData", libSceNpTus_stub)
	elf.RegisterStub("libSceNpTus", "sceNpTusGetDataVUser", libSceNpTus_stub)
	elf.RegisterStub("libSceNpTus", "sceNpTusSetData", libSceNpTus_stub)
	elf.RegisterStub("libSceNpTus", "sceNpTusSetDataVUser", libSceNpTus_stub)
	elf.RegisterStub("libSceNpTus", "sceNpTusGetMultiSlotVariable", libSceNpTus_stub)
	elf.RegisterStub("libSceNpTus", "sceNpTusGetMultiSlotVariableVUser", libSceNpTus_stub)
	elf.RegisterStub("libSceNpTus", "sceNpTusSetMultiSlotVariable", libSceNpTus_stub)
	elf.RegisterStub("libSceNpTus", "sceNpTusSetMultiSlotVariableVUser", libSceNpTus_stub)
	elf.RegisterStub("libSceNpTus", "sceNpTusDeleteNpTitleCtx", libSceNpTus_stub)
	elf.RegisterStub("libSceNpTus", "sceNpTusDeleteRequest", libSceNpTus_stub)
	elf.RegisterStub("libSceNpTus", "sceNpTssGetData", libSceNpTus_stub)
	elf.RegisterStub("libSceNpTus", "sceNpTusCreateRequest", libSceNpTus_stub)
	elf.RegisterStub("libSceNpTus", "sceNpTusCreateNpTitleCtx", libSceNpTus_stub)
}

func libSceNpTus_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
