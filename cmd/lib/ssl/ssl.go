package ssl

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterSslStubs() {
	elf.RegisterStub("libSceSsl", "sceSslTerm", libSceSsl_stub)
	elf.RegisterStub("libSceSsl", "sceSslGetSerialNumber", libSceSsl_stub)
	elf.RegisterStub("libSceSsl", "sceSslInit", libSceSsl_stub)
}

func libSceSsl_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
