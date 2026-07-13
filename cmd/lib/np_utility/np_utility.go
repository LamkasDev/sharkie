package np_utility

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterNpUtilityStubs() {
	elf.RegisterStub("libSceNpUtility", "sceNpWordFilterPollAsync", libSceNpUtility_stub)
	elf.RegisterStub("libSceNpUtility", "sceNpWordFilterAbortRequest", libSceNpUtility_stub)
	elf.RegisterStub("libSceNpUtility", "sceNpWordFilterDeleteRequest", libSceNpUtility_stub)
	elf.RegisterStub("libSceNpUtility", "sceNpWordFilterCreateAsyncRequest", libSceNpUtility_stub)
	elf.RegisterStub("libSceNpUtility", "sceNpWordFilterCreateRequest", libSceNpUtility_stub)
	elf.RegisterStub("libSceNpUtility", "sceNpWordFilterSanitizeComment", libSceNpUtility_stub)
	elf.RegisterStub("libSceNpUtility", "sceNpWordFilterCensorComment", libSceNpUtility_stub)
	elf.RegisterStub("libSceNpUtility", "sceNpWordFilterCreateTitleCtx", libSceNpUtility_stub)
	elf.RegisterStub("libSceNpUtility", "sceNpWordFilterDeleteTitleCtx", libSceNpUtility_stub)
	elf.RegisterStub("libSceNpUtility", "sceNpBandwidthTestShutdown", libSceNpUtility_stub)
	elf.RegisterStub("libSceNpUtility", "sceNpBandwidthTestGetStatus", libSceNpUtility_stub)
	elf.RegisterStub("libSceNpUtility", "sceNpBandwidthTestInitStart", libSceNpUtility_stub)
	elf.RegisterStub("libSceNpUtility", "sceNpLookupDeleteRequest", libSceNpUtility_stub)
	elf.RegisterStub("libSceNpUtility", "sceNpLookupNpId", libSceNpUtility_stub)
	elf.RegisterStub("libSceNpUtility", "sceNpLookupCreateRequest", libSceNpUtility_stub)
	elf.RegisterStub("libSceNpUtility", "sceNpLookupCreateTitleCtx", libSceNpUtility_stub)
	elf.RegisterStub("libSceNpUtility", "sceNpLookupDeleteTitleCtx", libSceNpUtility_stub)
}

func libSceNpUtility_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
