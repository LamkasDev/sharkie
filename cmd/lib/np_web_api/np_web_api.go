package np_web_api

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterNpWebApiStubs() {
	// Setup functions.
	elf.RegisterStub("libSceNpWebApi", "sceNpWebApiInitialize", libSceNpWebApi_sceNpWebApiInitialize)

	elf.RegisterStub("libSceNpWebApi", "sceNpWebApiUtilityParseNpId", libSceNpWebApi_stub)
	elf.RegisterStub("libSceNpWebApi", "sceNpWebApiDeleteContext", libSceNpWebApi_stub)
	elf.RegisterStub("libSceNpWebApi", "sceNpWebApiUnregisterPushEventCallback", libSceNpWebApi_stub)
	elf.RegisterStub("libSceNpWebApi", "sceNpWebApiUnregisterServicePushEventCallback", libSceNpWebApi_stub)
	elf.RegisterStub("libSceNpWebApi", "sceNpWebApiAbortRequest", libSceNpWebApi_stub)
	elf.RegisterStub("libSceNpWebApi", "sceNpWebApiRegisterServicePushEventCallback", libSceNpWebApi_stub)
	elf.RegisterStub("libSceNpWebApi", "sceNpWebApiRegisterPushEventCallback", libSceNpWebApi_stub)
	elf.RegisterStub("libSceNpWebApi", "sceNpWebApiCreateContext", libSceNpWebApi_stub)
	elf.RegisterStub("libSceNpWebApi", "sceNpWebApiGetHttpResponseHeaderValue", libSceNpWebApi_stub)
	elf.RegisterStub("libSceNpWebApi", "sceNpWebApiGetHttpResponseHeaderValueLength", libSceNpWebApi_stub)
	elf.RegisterStub("libSceNpWebApi", "sceNpWebApiReadData", libSceNpWebApi_stub)
	elf.RegisterStub("libSceNpWebApi", "sceNpWebApiGetHttpStatusCode", libSceNpWebApi_stub)
	elf.RegisterStub("libSceNpWebApi", "sceNpWebApiSendRequest", libSceNpWebApi_stub)
	elf.RegisterStub("libSceNpWebApi", "sceNpWebApiCreateRequest", libSceNpWebApi_stub)
	elf.RegisterStub("libSceNpWebApi", "sceNpWebApiDeleteRequest", libSceNpWebApi_stub)
	elf.RegisterStub("libSceNpWebApi", "sceNpWebApiTerminate", libSceNpWebApi_stub)
	elf.RegisterStub("libSceNpWebApi", "sceNpWebApiDeleteHandle", libSceNpWebApi_stub)
	elf.RegisterStub("libSceNpWebApi", "sceNpWebApiDeletePushEventFilter", libSceNpWebApi_stub)
	elf.RegisterStub("libSceNpWebApi", "sceNpWebApiDeleteServicePushEventFilter", libSceNpWebApi_stub)
	elf.RegisterStub("libSceNpWebApi", "sceNpWebApiCreateServicePushEventFilter", libSceNpWebApi_stub)
	elf.RegisterStub("libSceNpWebApi", "sceNpWebApiCreatePushEventFilter", libSceNpWebApi_stub)
	elf.RegisterStub("libSceNpWebApi", "sceNpWebApiCreateHandle", libSceNpWebApi_stub)
}

func libSceNpWebApi_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
