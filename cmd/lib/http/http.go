package http

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterHttpStubs() {
	// Setup functions.
	elf.RegisterStub("libSceHttp", "sceHttpInit", libSceHttp_sceHttpInit)
	elf.RegisterStub("libSceHttp", "sceHttpCreateTemplate", libSceHttp_sceHttpCreateTemplate)
	elf.RegisterStub("libSceHttp", "sceHttpDeleteTemplate", libSceHttp_sceHttpDeleteTemplate)

	// URI functions.
	elf.RegisterStub("libSceHttp", "sceHttpUriParse", libSceHttp_sceHttpUriParse)
	elf.RegisterStub("libSceHttp", "sceHttpUriSweepPath", libSceHttp_sceHttpUriSweepPath)

	// Connection functions.
	elf.RegisterStub("libSceHttp", "sceHttpCreateConnection", libSceHttp_sceHttpCreateConnection)
	elf.RegisterStub("libSceHttp", "sceHttpCreateConnectionWithURL", libSceHttp_sceHttpCreateConnectionWithURL)
	elf.RegisterStub("libSceHttp", "sceHttpDeleteConnection", libSceHttp_sceHttpDeleteConnection)

	// Header functions.
	elf.RegisterStub("libSceHttp", "sceHttpAddRequestHeader", libSceHttp_sceHttpAddRequestHeader)

	// Request functions.
	elf.RegisterStub("libSceHttp", "sceHttpCreateRequestWithURL", libSceHttp_sceHttpCreateRequestWithURL)
	elf.RegisterStub("libSceHttp", "sceHttpDeleteRequest", libSceHttp_sceHttpDeleteRequest)
	elf.RegisterStub("libSceHttp", "sceHttpSendRequest", libSceHttp_sceHttpSendRequest)
	elf.RegisterStub("libSceHttp", "sceHttpGetStatusCode", libSceHttp_sceHttpGetStatusCode)
	elf.RegisterStub("libSceHttp", "sceHttpGetAllResponseHeaders", libSceHttp_sceHttpGetAllResponseHeaders)
	elf.RegisterStub("libSceHttp", "sceHttpGetResponseContentLength", libSceHttp_sceHttpGetResponseContentLength)
	elf.RegisterStub("libSceHttp", "sceHttpReadData", libSceHttp_sceHttpReadData)

	// Epoll functions.
	elf.RegisterStub("libSceHttp", "sceHttpCreateEpoll", libSceHttp_sceHttpCreateEpoll)

	elf.RegisterStub("libSceHttp", "sceHttpTerm", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpUriEscape", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpAbortRequest", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpSetSendTimeOut", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpSetRecvTimeOut", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpSetConnectTimeOut", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpSetResolveRetry", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpSetResolveTimeOut", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpParseResponseHeader", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpCreateRequest", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpsSetSslCallback", libSceHttp_stub)
	elf.RegisterStub("libSceHttp", "sceHttpsDisableOption", libSceHttp_stub)
}

func libSceHttp_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
