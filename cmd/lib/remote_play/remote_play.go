package remote_play

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterRemotePlayStubs() {
	elf.RegisterStub("libSceRemoteplay", "sceRemoteplayProhibit", libSceRemoteplay_stub)
	elf.RegisterStub("libSceRemoteplay", "sceRemoteplayApprove", libSceRemoteplay_stub)
	elf.RegisterStub("libSceRemoteplay", "sceRemoteplayInitialize", libSceRemoteplay_stub)
}

func libSceRemoteplay_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
