package lib

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/lib/app_content"
	"github.com/LamkasDev/sharkie/cmd/lib/audio_in"
	"github.com/LamkasDev/sharkie/cmd/lib/audio_out"
	"github.com/LamkasDev/sharkie/cmd/lib/common_dialog"
	"github.com/LamkasDev/sharkie/cmd/lib/error_dialog"
	"github.com/LamkasDev/sharkie/cmd/lib/game_custom_data_dialog"
	"github.com/LamkasDev/sharkie/cmd/lib/gnm_driver"
	"github.com/LamkasDev/sharkie/cmd/lib/http"
	"github.com/LamkasDev/sharkie/cmd/lib/ime"
	"github.com/LamkasDev/sharkie/cmd/lib/ime_dialog"
	"github.com/LamkasDev/sharkie/cmd/lib/invitation_dialog"
	"github.com/LamkasDev/sharkie/cmd/lib/kernel"
	"github.com/LamkasDev/sharkie/cmd/lib/libc"
	"github.com/LamkasDev/sharkie/cmd/lib/msg_dialog"
	"github.com/LamkasDev/sharkie/cmd/lib/net"
	"github.com/LamkasDev/sharkie/cmd/lib/net_ctl"
	"github.com/LamkasDev/sharkie/cmd/lib/np_auth"
	"github.com/LamkasDev/sharkie/cmd/lib/np_commerce"
	"github.com/LamkasDev/sharkie/cmd/lib/np_common"
	"github.com/LamkasDev/sharkie/cmd/lib/np_manager"
	"github.com/LamkasDev/sharkie/cmd/lib/np_matching2"
	"github.com/LamkasDev/sharkie/cmd/lib/np_score"
	"github.com/LamkasDev/sharkie/cmd/lib/np_signaling"
	"github.com/LamkasDev/sharkie/cmd/lib/np_sns"
	"github.com/LamkasDev/sharkie/cmd/lib/np_sns_facebook_dialog"
	"github.com/LamkasDev/sharkie/cmd/lib/np_trophy"
	"github.com/LamkasDev/sharkie/cmd/lib/np_tus"
	"github.com/LamkasDev/sharkie/cmd/lib/np_utility"
	"github.com/LamkasDev/sharkie/cmd/lib/np_web_api"
	"github.com/LamkasDev/sharkie/cmd/lib/pad"
	"github.com/LamkasDev/sharkie/cmd/lib/pngdec"
	"github.com/LamkasDev/sharkie/cmd/lib/posix"
	"github.com/LamkasDev/sharkie/cmd/lib/random"
	"github.com/LamkasDev/sharkie/cmd/lib/remote_play"
	"github.com/LamkasDev/sharkie/cmd/lib/rtc"
	"github.com/LamkasDev/sharkie/cmd/lib/rudp"
	"github.com/LamkasDev/sharkie/cmd/lib/save_data"
	"github.com/LamkasDev/sharkie/cmd/lib/save_data_dialog"
	"github.com/LamkasDev/sharkie/cmd/lib/ssl"
	"github.com/LamkasDev/sharkie/cmd/lib/sysmodule"
	"github.com/LamkasDev/sharkie/cmd/lib/system_service"
	"github.com/LamkasDev/sharkie/cmd/lib/user_service"
	"github.com/LamkasDev/sharkie/cmd/lib/video_out"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func RegisterStubs() {
	elf.RegisterStub("", "__sharkie_generic_stub", GenericStub)

	app_content.RegisterAppContentStubs()
	audio_in.RegisterAudioInStubs()
	audio_out.RegisterAudioOutStubs()
	common_dialog.RegisterCommonDialogStubs()
	error_dialog.RegisterErrorDialogStubs()
	game_custom_data_dialog.RegisterGameCustomDataDialogStubs()
	gnm_driver.RegisterGnmDriverStubs()
	http.RegisterHttpStubs()
	ime.RegisterImeStubs()
	ime_dialog.RegisterImeDialogStubs()
	invitation_dialog.RegisterInvitationDialogStubs()
	kernel.RegisterKernelStubs()
	libc.RegisterSceLibcInternalStubs()
	libc.RegisterLibcStubs()
	pngdec.RegisterPngDecStubs()
	msg_dialog.RegisterMsgDialogStubs()
	net.RegisterNetStubs()
	net_ctl.RegisterNetCtlStubs()
	np_auth.RegisterNpAuthStubs()
	np_commerce.RegisterNpCommerceStubs()
	np_common.RegisterNpCommonStubs()
	np_manager.RegisterNpManagerStubs()
	np_matching2.RegisterNpMatching2Stubs()
	np_score.RegisterNpScoreStubs()
	np_signaling.RegisterNpSignalingStubs()
	np_sns.RegisterNpSnsStubs()
	np_sns_facebook_dialog.RegisterNpSnsFacebookDialogStubs()
	np_trophy.RegisterNpTrophyStubs()
	np_tus.RegisterNpTusStubs()
	np_utility.RegisterNpUtilityStubs()
	np_web_api.RegisterNpWebApiStubs()
	pad.RegisterPadStubs()
	posix.RegisterPosixStubs()
	random.RegisterRandomStubs()
	remote_play.RegisterRemotePlayStubs()
	rtc.RegisterRtcStubs()
	rudp.RegisterRudpStubs()
	save_data.RegisterSaveDataStubs()
	save_data_dialog.RegisterSaveDataDialogStubs()
	ssl.RegisterSslStubs()
	sysmodule.RegisterSysmoduleStubs()
	system_service.RegisterSystemServiceStubs()
	user_service.RegisterUserServiceStubs()
	video_out.RegisterVideoOutStubs()
	// voice.RegisterVoiceStubs()

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
