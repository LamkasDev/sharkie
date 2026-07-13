package http

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterHttpStubs() {
	elf.RegisterStub("libSceHttp", "sceHttpTerm", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpUriEscape", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpAbortRequest", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpSetSendTimeOut", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpSetRecvTimeOut", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpSetConnectTimeOut", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpSetResolveRetry", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpSetResolveTimeOut", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpParseResponseHeader", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpGetAllResponseHeaders", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpAddRequestHeader", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpCreateRequest", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpDeleteTemplate", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpDeleteConnection", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpDeleteRequest", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpReadData", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpGetResponseContentLength", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpGetStatusCode", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpSendRequest", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpCreateRequestWithURL", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpCreateConnectionWithURL", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpsSetSslCallback", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpsDisableOption", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpCreateTemplate", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpInit", libSceHttp_stub)
}

func libSceHttp_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
