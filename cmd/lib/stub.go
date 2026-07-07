package lib

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/lib/audio_out"
	"github.com/LamkasDev/sharkie/cmd/lib/gnm_driver"
	"github.com/LamkasDev/sharkie/cmd/lib/kernel"
	"github.com/LamkasDev/sharkie/cmd/lib/libc"
	"github.com/LamkasDev/sharkie/cmd/lib/net_ctl"
	"github.com/LamkasDev/sharkie/cmd/lib/system_service"
	"github.com/LamkasDev/sharkie/cmd/lib/video_out"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterStubs() {
	elf.RegisterStub("", "__sharkie_generic_stub", GenericStub)

	kernel.RegisterKernelStubs()
	libc.RegisterSceLibcInternalStubs()
	libc.RegisterLibcStubs()
	gnm_driver.RegisterGnmDriverStubs()
	video_out.RegisterVideoOutStubs()
	audio_out.RegisterAudioOutStubs()
	net_ctl.RegisterNetCtlStubs()
	system_service.RegisterSystemServiceStubs()

	RegisterMinecraftStubs()
}

func GenericStub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
