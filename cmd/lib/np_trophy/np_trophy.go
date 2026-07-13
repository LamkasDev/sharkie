package np_trophy

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterNpTrophyStubs() {
	elf.RegisterStub("libSceNpTrophy", "sceNpTrophyGetGroupInfo", libSceNpTrophy_stub)
	elf.RegisterStub("libSceNpTrophy", "sceNpTrophyGetTrophyInfo", libSceNpTrophy_stub)
	elf.RegisterStub("libSceNpTrophy", "sceNpTrophyGetGameInfo", libSceNpTrophy_stub)
	elf.RegisterStub("libSceNpTrophy", "sceNpTrophyUnlockTrophy", libSceNpTrophy_stub)
	elf.RegisterStub("libSceNpTrophy", "sceNpTrophyDestroyContext", libSceNpTrophy_stub)
	elf.RegisterStub("libSceNpTrophy", "sceNpTrophyDestroyHandle", libSceNpTrophy_stub)
	elf.RegisterStub("libSceNpTrophy", "sceNpTrophyRegisterContext", libSceNpTrophy_stub)
	elf.RegisterStub("libSceNpTrophy", "sceNpTrophyCreateHandle", libSceNpTrophy_stub)
	elf.RegisterStub("libSceNpTrophy", "sceNpTrophyCreateContext", libSceNpTrophy_stub)
	elf.RegisterStub("libSceNpTrophy", "sceNpTrophyGetGameIcon", libSceNpTrophy_stub)
	elf.RegisterStub("libSceNpTrophy", "sceNpTrophyGetGroupIcon", libSceNpTrophy_stub)
	elf.RegisterStub("libSceNpTrophy", "sceNpTrophyGetTrophyIcon", libSceNpTrophy_stub)
}

func libSceNpTrophy_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
