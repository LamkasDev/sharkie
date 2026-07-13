package np_score

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterNpScoreStubs() {
	elf.RegisterStub("libSceNpScore", "sceNpScorePollAsync", libSceNpScore_stub)
	elf.RegisterStub("libSceNpScore", "sceNpScoreGetBoardInfo", libSceNpScore_stub)
	elf.RegisterStub("libSceNpScore", "sceNpScoreGetFriendsRankingAsync", libSceNpScore_stub)
	elf.RegisterStub("libSceNpScore", "sceNpScoreGetFriendsRanking", libSceNpScore_stub)
	elf.RegisterStub("libSceNpScore", "sceNpScoreGetRankingByRangeAsync", libSceNpScore_stub)
	elf.RegisterStub("libSceNpScore", "sceNpScoreGetRankingByNpIdAsync", libSceNpScore_stub)
	elf.RegisterStub("libSceNpScore", "sceNpScoreRecordScoreAsync", libSceNpScore_stub)
	elf.RegisterStub("libSceNpScore", "sceNpScoreAbortRequest", libSceNpScore_stub)
	elf.RegisterStub("libSceNpScore", "sceNpScoreRecordScore", libSceNpScore_stub)
	elf.RegisterStub("libSceNpScore", "sceNpScoreGetRankingByRange", libSceNpScore_stub)
	elf.RegisterStub("libSceNpScore", "sceNpScoreDeleteRequest", libSceNpScore_stub)
	elf.RegisterStub("libSceNpScore", "sceNpScoreGetRankingByNpId", libSceNpScore_stub)
	elf.RegisterStub("libSceNpScore", "sceNpScoreCreateRequest", libSceNpScore_stub)
	elf.RegisterStub("libSceNpScore", "sceNpScoreCreateNpTitleCtx", libSceNpScore_stub)
	elf.RegisterStub("libSceNpScore", "sceNpScoreDeleteNpTitleCtx", libSceNpScore_stub)
}

func libSceNpScore_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
