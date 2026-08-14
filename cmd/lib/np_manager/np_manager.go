package np_manager

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterNpManagerStubs() {
	// Setup functions.
	elf.RegisterStub("libSceNpManager", "sceNpGetState", libSceNpManager_sceNpGetState)
	elf.RegisterStub("libSceNpManager", "sceNpHasSignedUp", libSceNpManager_sceNpHasSignedUp)
	elf.RegisterStub("libSceNpManager", "sceNpGetOnlineId", libSceNpManager_sceNpGetOnlineId)
	elf.RegisterStub("libSceNpManager", "sceNpGetNpId", libSceNpManager_sceNpGetNpId)

	elf.RegisterStub("libSceNpManager", "sceNpUnregisterStateCallback", libSceNpManager_stub)
	elf.RegisterStub("libSceNpManager", "sceNpInGameMessageSendData", libSceNpManager_stub)
	elf.RegisterStub("libSceNpManager", "sceNpInGameMessagePrepare", libSceNpManager_stub)
	elf.RegisterStub("libSceNpManager", "sceNpInGameMessageCreateHandle", libSceNpManager_stub)
	elf.RegisterStub("libSceNpManager", "sceNpInGameMessageDeleteHandle", libSceNpManager_stub)
	elf.RegisterStub("libSceNpManager", "sceNpInGameMessageTerminate", libSceNpManager_stub)
	elf.RegisterStub("libSceNpManager", "sceNpAbortRequest", libSceNpManager_stub)
	elf.RegisterStub("libSceNpManager", "sceNpPollAsync", libSceNpManager_stub)
	elf.RegisterStub("libSceNpManager", "sceNpCheckCallback", libSceNpManager_stub)
	elf.RegisterStub("libSceNpManager", "sceNpRegisterGamePresenceCallback", libSceNpManager_stub)
	elf.RegisterStub("libSceNpManager", "sceNpRegisterStateCallback", libSceNpManager_stub)
	elf.RegisterStub("libSceNpManager", "sceNpGetGamePresenceStatus", libSceNpManager_stub)
	elf.RegisterStub("libSceNpManager", "sceNpGetParentalControlInfo", libSceNpManager_stub)
	elf.RegisterStub("libSceNpManager", "sceNpCheckNpAvailability", libSceNpManager_stub)
	elf.RegisterStub("libSceNpManager", "sceNpDeleteRequest", libSceNpManager_stub)
	elf.RegisterStub("libSceNpManager", "sceNpCheckPlus", libSceNpManager_stub)
	elf.RegisterStub("libSceNpManager", "sceNpCreateAsyncRequest", libSceNpManager_stub)
	elf.RegisterStub("libSceNpManager", "sceNpSetContentRestriction", libSceNpManager_stub)
	elf.RegisterStub("libSceNpManager", "sceNpNotifyPlusFeature", libSceNpManager_stub)
	elf.RegisterStub("libSceNpManager", "sceNpSetNpTitleId", libSceNpManager_stub)
	elf.RegisterStub("libSceNpManager", "sceNpInGameMessageInitialize", libSceNpManager_stub)
}

func libSceNpManager_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
