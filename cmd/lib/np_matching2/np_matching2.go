package np_matching2

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterNpMatching2Stubs() {
	elf.RegisterStub("libSceNpMatching2", "sceNpMatching2SendRoomChatMessage", libSceNpMatching2_stub)
	elf.RegisterStub("libSceNpMatching2", "sceNpMatching2GetRoomDataInternal", libSceNpMatching2_stub)
	elf.RegisterStub("libSceNpMatching2", "sceNpMatching2SendRoomMessage", libSceNpMatching2_stub)
	elf.RegisterStub("libSceNpMatching2", "sceNpMatching2GetUserInfoList", libSceNpMatching2_stub)
	elf.RegisterStub("libSceNpMatching2", "sceNpMatching2SearchRoom", libSceNpMatching2_stub)
	elf.RegisterStub("libSceNpMatching2", "sceNpMatching2RegisterRoomMessageCallback", libSceNpMatching2_stub)
	elf.RegisterStub("libSceNpMatching2", "sceNpMatching2SetUserInfo", libSceNpMatching2_stub)
	elf.RegisterStub("libSceNpMatching2", "sceNpMatching2GetRoomDataExternalList", libSceNpMatching2_stub)
	elf.RegisterStub("libSceNpMatching2", "sceNpMatching2GetRoomMemberDataInternal", libSceNpMatching2_stub)
	elf.RegisterStub("libSceNpMatching2", "sceNpMatching2KickoutRoomMember", libSceNpMatching2_stub)
	elf.RegisterStub("libSceNpMatching2", "sceNpMatching2ContextStop", libSceNpMatching2_stub)
	elf.RegisterStub("libSceNpMatching2", "sceNpMatching2DestroyContext", libSceNpMatching2_stub)
	elf.RegisterStub("libSceNpMatching2", "sceNpMatching2SignalingGetLocalNetInfo", libSceNpMatching2_stub)
	elf.RegisterStub("libSceNpMatching2", "sceNpMatching2GetServerId", libSceNpMatching2_stub)
	elf.RegisterStub("libSceNpMatching2", "sceNpMatching2SignalingGetConnectionStatus", libSceNpMatching2_stub)
	elf.RegisterStub("libSceNpMatching2", "sceNpMatching2SetRoomDataInternal", libSceNpMatching2_stub)
	elf.RegisterStub("libSceNpMatching2", "sceNpMatching2SetRoomMemberDataInternal", libSceNpMatching2_stub)
	elf.RegisterStub("libSceNpMatching2", "sceNpMatching2LeaveRoom", libSceNpMatching2_stub)
	elf.RegisterStub("libSceNpMatching2", "sceNpMatching2SetRoomDataExternal", libSceNpMatching2_stub)
	elf.RegisterStub("libSceNpMatching2", "sceNpMatching2CreateJoinRoom", libSceNpMatching2_stub)
	elf.RegisterStub("libSceNpMatching2", "sceNpMatching2JoinRoom", libSceNpMatching2_stub)
	elf.RegisterStub("libSceNpMatching2", "sceNpMatching2GetWorldInfoList", libSceNpMatching2_stub)
	elf.RegisterStub("libSceNpMatching2", "sceNpMatching2RegisterRoomEventCallback", libSceNpMatching2_stub)
	elf.RegisterStub("libSceNpMatching2", "sceNpMatching2RegisterSignalingCallback", libSceNpMatching2_stub)
	elf.RegisterStub("libSceNpMatching2", "sceNpMatching2SetDefaultRequestOptParam", libSceNpMatching2_stub)
	elf.RegisterStub("libSceNpMatching2", "sceNpMatching2RegisterContextCallback", libSceNpMatching2_stub)
	elf.RegisterStub("libSceNpMatching2", "sceNpMatching2Terminate", libSceNpMatching2_stub)
	elf.RegisterStub("libSceNpMatching2", "sceNpMatching2ContextStart", libSceNpMatching2_stub)
	elf.RegisterStub("libSceNpMatching2", "sceNpMatching2CreateContext", libSceNpMatching2_stub)
	elf.RegisterStub("libSceNpMatching2", "sceNpMatching2Initialize", libSceNpMatching2_stub)
}

func libSceNpMatching2_stub() uintptr {
	logger.Printf(
		"%-132s hi from %s :3\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprintf("generic stub"),
	)

	return 0
}
