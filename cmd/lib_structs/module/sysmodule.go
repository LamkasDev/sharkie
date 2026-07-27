package module

type SysmoduleId uint32

const (
	SysmoduleInvalid                   = SysmoduleId(0x0000)
	SysmoduleFiber                     = SysmoduleId(0x0006) // libSceFiber.sprx
	SysmoduleUlt                       = SysmoduleId(0x0007) // libSceUlt.sprx
	SysmoduleNgs2                      = SysmoduleId(0x000B) // libSceNgs2.sprx
	SysmoduleXml                       = SysmoduleId(0x0017) // libSceXml.sprx
	SysmoduleNpUtility                 = SysmoduleId(0x0019) // libSceNpUtility.sprx
	SysmoduleVoice                     = SysmoduleId(0x001A) // libSceVoice.sprx
	SysmoduleVoiceqos                  = SysmoduleId(0x001B) // libSceVoiceQos.sprx
	SysmoduleNpMatching2               = SysmoduleId(0x001C) // libSceNpMatching2.sprx
	SysmoduleNpScoreRanking            = SysmoduleId(0x001E) // libSceNpScoreRanking.sprx
	SysmoduleRudp                      = SysmoduleId(0x0021) // libSceRudp.sprx
	SysmoduleNpTus                     = SysmoduleId(0x002C) // libSceNpTus.sprx
	SysmoduleFace                      = SysmoduleId(0x0038) // libSceFace.sprx
	SysmoduleSmart                     = SysmoduleId(0x0039) // libSceSmart.sprx
	SysmoduleJson                      = SysmoduleId(0x0080) // libSceJson.sprx
	SysmoduleGameLiveStreaming         = SysmoduleId(0x0081) // libSceGameLiveStreaming.sprx
	SysmoduleCompanionUtil             = SysmoduleId(0x0082) // libSceCompanionUtil.sprx
	SysmodulePlaygo                    = SysmoduleId(0x0083) // libScePlayGo.sprx
	SysmoduleFont                      = SysmoduleId(0x0084) // libSceFont.sprx
	SysmoduleVideoRecording            = SysmoduleId(0x0085) // libSceVideoRecording.sprx
	SysmoduleS3dconversion             = SysmoduleId(0x0086) // libSceS3DConversion.sprx
	SysmoduleAudiodec                  = SysmoduleId(0x0088) // libSceAudiodec.sprx
	SysmoduleJpegDec                   = SysmoduleId(0x008A) // libSceJpegDec.sprx
	SysmoduleJpegEnc                   = SysmoduleId(0x008B) // libSceJpegEnc.sprx
	SysmodulePngDec                    = SysmoduleId(0x008C) // libScePngDec.sprx
	SysmodulePngEnc                    = SysmoduleId(0x008D) // libScePngEnc.sprx
	SysmoduleVideodec                  = SysmoduleId(0x008E) // libSceVideodec.sprx
	SysmoduleMove                      = SysmoduleId(0x008F) // libSceMove.sprx
	SysmodulePadTracker                = SysmoduleId(0x0091) // libScePadTracker.sprx
	SysmoduleDepth                     = SysmoduleId(0x0092) // libSceDepth.sprx
	SysmoduleHand                      = SysmoduleId(0x0093) // libSceHand.sprx
	SysmoduleLibime                    = SysmoduleId(0x0095) // libSceIme.sprx
	SysmoduleImeDialog                 = SysmoduleId(0x0096) // libSceImeDialog.sprx
	SysmoduleNpParty                   = SysmoduleId(0x0097) // libSceNpParty.sprx
	SysmoduleFontFt                    = SysmoduleId(0x0098) // libSceFontFt.sprx
	SysmoduleFreetypeOt                = SysmoduleId(0x0099) // libSceFreeTypeOt.sprx
	SysmoduleFreetypeOl                = SysmoduleId(0x009A) // libSceFreeTypeOl.sprx
	SysmoduleFreetypeOptOl             = SysmoduleId(0x009B) // libSceFreeTypeOptOl.sprx
	SysmoduleScreenShot                = SysmoduleId(0x009C) // libSceScreenShot.sprx
	SysmoduleNpAuth                    = SysmoduleId(0x009D) // libSceNpAuth.sprx
	SysmoduleSulpha                    = SysmoduleId(0x009F)
	SysmoduleSaveDataDialog            = SysmoduleId(0x00A0) // libSceSaveDataDialog.sprx
	SysmoduleInvitationDialog          = SysmoduleId(0x00A2) // libSceInvitationDialog.sprx
	SysmoduleDebugKeyboard             = SysmoduleId(0x00A3)
	SysmoduleMessageDialog             = SysmoduleId(0x00A4) // libSceMsgDialog.sprx
	SysmoduleAvPlayer                  = SysmoduleId(0x00A5) // libSceAvPlayer.sprx
	SysmoduleContentExport             = SysmoduleId(0x00A6) // libSceContentExport.sprx
	SysmoduleAudio3d                   = SysmoduleId(0x00A7) // libSceAudio3d.sprx
	SysmoduleNpCommerce                = SysmoduleId(0x00A8) // libSceNpCommerce.sprx
	SysmoduleMouse                     = SysmoduleId(0x00A9) // libSceMouse.sprx
	SysmoduleCompanionHttpd            = SysmoduleId(0x00AA) // libSceCompanionHttpd.sprx
	SysmoduleWebBrowserDialog          = SysmoduleId(0x00AB) // libSceWebBrowserDialog.sprx
	SysmoduleErrorDialog               = SysmoduleId(0x00AC) // libSceErrorDialog.sprx
	SysmoduleNpTrophy                  = SysmoduleId(0x00AD) // libSceNpTrophy.sprx
	SysmoduleVideoCoreIf               = SysmoduleId(0x00AE) // libSceVideoCoreInterface.sprx
	SysmoduleVideoCoreServerIf         = SysmoduleId(0x00AF) // libSceVideoCoreServerInterface.sprx
	SysmoduleNpSnsFacebook             = SysmoduleId(0x00B0) // libSceNpSnsFacebookDialog.sprx
	SysmoduleMoveTracker               = SysmoduleId(0x00B1) // libSceMoveTracker.sprx
	SysmoduleNpProfileDialog           = SysmoduleId(0x00B2) // libSceNpProfileDialog.sprx
	SysmoduleNpFriendListDialog        = SysmoduleId(0x00B3) // libSceNpFriendListDialog.sprx
	SysmoduleAppContent                = SysmoduleId(0x00B4) // libSceAppContent.sprx
	SysmoduleNpSignaling               = SysmoduleId(0x00B5) // libSceNpSignaling.sprx
	SysmoduleRemotePlay                = SysmoduleId(0x00B6) // libSceRemoteplay.sprx
	SysmoduleUsbd                      = SysmoduleId(0x00B7) // libSceUsbd.sprx
	SysmoduleGameCustomDataDialog      = SysmoduleId(0x00B8) // libSceGameCustomDataDialog.sprx
	SysmoduleNpEulaDialog              = SysmoduleId(0x00B9) // libSceNpEulaDialog.sprx
	SysmoduleRandom                    = SysmoduleId(0x00BA) // libSceRandom.sprx
	SysmoduleReserved2                 = SysmoduleId(0x00BB)
	SysmoduleM4aacEnc                  = SysmoduleId(0x00BC) // libSceM4aacEnc.sprx
	SysmoduleAudiodecCpu               = SysmoduleId(0x00BD) // libSceAudiodecCpu.sprx
	SysmoduleAudiodecCpuDdp            = SysmoduleId(0x00BE) // libSceAudiodecCpuDdp.sprx
	SysmoduleAudiodecCpuM4aac          = SysmoduleId(0x00C0) // libSceAudiodecCpuM4aac.sprx
	SysmoduleBemp2Sys                  = SysmoduleId(0x00C1) // libSceBemp2sys.sprx
	SysmoduleBeisobmf                  = SysmoduleId(0x00C2) // libSceBeisobmf.sprx
	SysmodulePlayReady                 = SysmoduleId(0x00C3) // libScePlayReady.sprx
	SysmoduleVideoNativeExtEssential   = SysmoduleId(0x00C4) // libSceVideoNativeExtEssential.sprx
	SysmoduleZlib                      = SysmoduleId(0x00C5) // libSceZlib.sprx
	SysmoduleDtcpIp                    = SysmoduleId(0x00C6) // libSceDtcpIp.sprx
	SysmoduleContentSearch             = SysmoduleId(0x00C7) // libSceContentSearch.sprx
	SysmoduleShareUtility              = SysmoduleId(0x00C8) // libSceShareUtility.sprx
	SysmoduleAudiodecCpuDtsHdLbr       = SysmoduleId(0x00C9) // libSceAudiodecCpuDtsHdLbr.sprx
	SysmoduleDeci4h                    = SysmoduleId(0x00CA)
	SysmoduleHeadTracker               = SysmoduleId(0x00CB) // libSceHeadTracker.sprx
	SysmoduleGameUpdate                = SysmoduleId(0x00CC) // libSceGameUpdate.sprx
	SysmoduleAutoMounterClient         = SysmoduleId(0x00CD) // libSceAutoMounterClient.sprx
	SysmoduleSystemGesture             = SysmoduleId(0x00CE) // libSceSystemGesture.sprx
	SysmoduleVideodec2                 = SysmoduleId(0x00CF) // libSceVideodec2.sprx
	SysmoduleVdecwrap                  = SysmoduleId(0x00D0) // libSceVdecwrap.sprx
	SysmoduleAt9Enc                    = SysmoduleId(0x00D1) // libSceAt9Enc.sprx
	SysmoduleConvertKeycode            = SysmoduleId(0x00D2) // libSceConvertKeycode.sprx
	SysmoduleSharePlay                 = SysmoduleId(0x00D3) // libSceSharePlay.sprx
	SysmoduleHmd                       = SysmoduleId(0x00D4) // libSceHmd.sprx
	SysmoduleUsbStorage                = SysmoduleId(0x00D5) // libSceUsbStorage.sprx
	SysmoduleUsbStorageDialog          = SysmoduleId(0x00D6) // libSceUsbStorageDialog.sprx
	SysmoduleDiscMap                   = SysmoduleId(0x00D7) // libSceDiscMap.sprx
	SysmoduleFaceTracker               = SysmoduleId(0x00D8) // libSceFaceTracker.sprx
	SysmoduleHandTracker               = SysmoduleId(0x00D9) // libSceHandTracker.sprx
	SysmoduleNpSnsYoutubeDialog        = SysmoduleId(0x00DA) // libSceNpSnsYouTubeDialog.sprx
	SysmoduleProfileCacheExternal      = SysmoduleId(0x00DC) // libSceProfileCacheExternal.sprx
	SysmoduleMusicPlayerService        = SysmoduleId(0x00DD) // libSceMusicPlayerService.sprx
	SysmoduleSpSysCallWrapper          = SysmoduleId(0x00DE) // libSceSpSysCallWrapper.sprx
	SysmodulePs2EmuMenuDialog          = SysmoduleId(0x00DF) // libScePs2EmuMenuDialog.sprx
	SysmoduleNpSnsDailymotionDialog    = SysmoduleId(0x00E0) // libSceNpSnsDailyMotionDialog.sprx
	SysmoduleAudiodecCpuHevag          = SysmoduleId(0x00E1) // libSceAudiodecCpuHevag.sprx
	SysmoduleLoginDialog               = SysmoduleId(0x00E2) // libSceLoginDialog.sprx
	SysmoduleLoginService              = SysmoduleId(0x00E3) // libSceLoginService.sprx
	SysmoduleSigninDialog              = SysmoduleId(0x00E4) // libSceSigninDialog.sprx
	SysmoduleVdecsw                    = SysmoduleId(0x00E5) // libSceVdecsw.sprx
	SysmoduleCustomMusicCore           = SysmoduleId(0x00E6) // libSceCustomMusicCore.sprx
	SysmoduleJson2                     = SysmoduleId(0x00E7) // libSceJson2.sprx
	SysmoduleAudioLatencyEstimation    = SysmoduleId(0x00E8) // libSceAudioLatencyEstimation.sprx
	SysmoduleWkFontConfig              = SysmoduleId(0x00E9) // libSceWkFontConfig.sprx
	SysmoduleVorbisDec                 = SysmoduleId(0x00EA) // libSceVorbisDec.sprx
	SysmoduleHmdSetupDialog            = SysmoduleId(0x00EB) // libSceHmdSetupDialog.sprx
	SysmoduleReserved28                = SysmoduleId(0x00EC)
	SysmoduleVrTracker                 = SysmoduleId(0x00ED) // libSceVrTracker.sprx
	SysmoduleContentDelete             = SysmoduleId(0x00EE) // libSceContentDelete.sprx
	SysmoduleImeBackend                = SysmoduleId(0x00EF) // libSceImeBackend.sprx
	SysmoduleNetCtlApDialog            = SysmoduleId(0x00F0) // libSceNetCtlApDialog.sprx
	SysmodulePlaygoDialog              = SysmoduleId(0x00F1) // libScePlayGoDialog.sprx
	SysmoduleSocialScreen              = SysmoduleId(0x00F2) // libSceSocialScreen.sprx
	SysmoduleEditMp4                   = SysmoduleId(0x00F3) // libSceEditMp4.sprx
	SysmodulePsmKitSystem              = SysmoduleId(0x00F5) // libScePsmKitSystem.sprx
	SysmoduleTextToSpeech              = SysmoduleId(0x00F6) // libSceTextToSpeech.sprx
	SysmoduleNpToolkit                 = SysmoduleId(0x00F7) // libSceNpToolkit.sprx
	SysmoduleCustomMusicService        = SysmoduleId(0x00F8) // libSceCustomMusicService.sprx
	SysmoduleClSysCallWrapper          = SysmoduleId(0x00F9) // libSceClSysCallWrapper.sprx
	SysmoduleSystemLogger              = SysmoduleId(0x00FA) // libSceSystemLogger.sprx
	SysmoduleBluetoothHid              = SysmoduleId(0x00FB) // libSceBluetoothHid.sprx
	SysmoduleVideoDecoderArbitration   = SysmoduleId(0x00FC) // libSceVideoDecoderArbitration.sprx
	SysmoduleVrServiceDialog           = SysmoduleId(0x00FD) // libSceVrServiceDialog.sprx
	SysmoduleJobManager                = SysmoduleId(0x00FE) // libSceJobManager.sprx
	SysmoduleShareFactoryUtil          = SysmoduleId(0x00FF) // libSceShareFactoryUtil.sprx
	SysmoduleSocialScreenDialog        = SysmoduleId(0x0100) // libSceSocialScreenDialog.sprx
	SysmoduleNpSnsDialog               = SysmoduleId(0x0101) // libSceNpSnsDialog.sprx
	SysmoduleNpToolkit2                = SysmoduleId(0x0102) // libSceNpToolkit2.sprx
	SysmoduleSrcUtl                    = SysmoduleId(0x0103) // libSceSrcUtl.sprx
	SysmoduleDiscId                    = SysmoduleId(0x0104) // libSceDiscId.sprx
	SysmoduleNpUniversalDataSystem     = SysmoduleId(0x0105) // libSceNpUniversalDataSystem.sprx
	SysmoduleKeyboard                  = SysmoduleId(0x0106) // libSceKeyboard.sprx
	SysmoduleGic                       = SysmoduleId(0x0107) // libSceGic.sprx
	SysmodulePlayReady2                = SysmoduleId(0x0108) // libScePlayReady2.sprx
	SysmoduleCesCs                     = SysmoduleId(0x010C) // libSceCesCs.sprx
	SysmodulePlayerInvitationDialog    = SysmoduleId(0x010D) // libScePlayerInvitationDialog.sprx
	SysmoduleNpSessionSignaling        = SysmoduleId(0x0112) // libSceNpSessionSignaling.sprx
	SysmoduleNpEntitlementAccess       = SysmoduleId(0x0113) // libSceNpEntitlementAccess.sprx
	SysmoduleNpCppWebApi               = SysmoduleId(0x0115) // libSceNpCppWebApi.sprx
	SysmoduleHubAppUtil                = SysmoduleId(0x0116) // libSceHubAppUtil.sprx
	SysmoduleNpPartner001              = SysmoduleId(0x011A) // libSceNpPartner001.sprx
	SysmoduleFontGs                    = SysmoduleId(0x012F) // libSceFontGs.sprx
	SysmoduleFontGsm                   = SysmoduleId(0x0135) // libSceFontGsm.sprx
	SysmoduleNpPartnerSubscription     = SysmoduleId(0x0138) // libSceNpPartnerSubscription.sprx
	SysmoduleNpAuthAuthorizedAppDialog = SysmoduleId(0x0139) // libSceNpAuthAuthorizedAppDialog.sprx
)

const (
	SysmoduleInternalRazorCpu = SysmoduleId(0x80000019) // libSceRazorCpu.sprx
)

var SysmoduleMap = map[SysmoduleId]string{
	SysmoduleFiber:                     "libSceFiber.sprx",
	SysmoduleUlt:                       "libSceUlt.sprx",
	SysmoduleNgs2:                      "libSceNgs2.sprx",
	SysmoduleXml:                       "libSceXml.sprx",
	SysmoduleNpUtility:                 "libSceNpUtility.sprx",
	SysmoduleVoice:                     "libSceVoice.sprx",
	SysmoduleVoiceqos:                  "libSceVoiceQos.sprx",
	SysmoduleNpMatching2:               "libSceNpMatching2.sprx",
	SysmoduleNpScoreRanking:            "libSceNpScoreRanking.sprx",
	SysmoduleRudp:                      "libSceRudp.sprx",
	SysmoduleNpTus:                     "libSceNpTus.sprx",
	SysmoduleFace:                      "libSceFace.sprx",
	SysmoduleSmart:                     "libSceSmart.sprx",
	SysmoduleJson:                      "libSceJson.sprx",
	SysmoduleGameLiveStreaming:         "libSceGameLiveStreaming.sprx",
	SysmoduleCompanionUtil:             "libSceCompanionUtil.sprx",
	SysmodulePlaygo:                    "libScePlayGo.sprx",
	SysmoduleFont:                      "libSceFont.sprx",
	SysmoduleVideoRecording:            "libSceVideoRecording.sprx",
	SysmoduleS3dconversion:             "libSceS3DConversion.sprx",
	SysmoduleAudiodec:                  "libSceAudiodec.sprx",
	SysmoduleJpegDec:                   "libSceJpegDec.sprx",
	SysmoduleJpegEnc:                   "libSceJpegEnc.sprx",
	SysmodulePngDec:                    "libScePngDec.sprx",
	SysmodulePngEnc:                    "libScePngEnc.sprx",
	SysmoduleVideodec:                  "libSceVideodec.sprx",
	SysmoduleMove:                      "libSceMove.sprx",
	SysmodulePadTracker:                "libScePadTracker.sprx",
	SysmoduleDepth:                     "libSceDepth.sprx",
	SysmoduleHand:                      "libSceHand.sprx",
	SysmoduleLibime:                    "libSceIme.sprx",
	SysmoduleImeDialog:                 "libSceImeDialog.sprx",
	SysmoduleNpParty:                   "libSceNpParty.sprx",
	SysmoduleFontFt:                    "libSceFontFt.sprx",
	SysmoduleFreetypeOt:                "libSceFreeTypeOt.sprx",
	SysmoduleFreetypeOl:                "libSceFreeTypeOl.sprx",
	SysmoduleFreetypeOptOl:             "libSceFreeTypeOptOl.sprx",
	SysmoduleScreenShot:                "libSceScreenShot.sprx",
	SysmoduleNpAuth:                    "libSceNpAuth.sprx",
	SysmoduleSaveDataDialog:            "libSceSaveDataDialog.sprx",
	SysmoduleInvitationDialog:          "libSceInvitationDialog.sprx",
	SysmoduleMessageDialog:             "libSceMsgDialog.sprx",
	SysmoduleAvPlayer:                  "libSceAvPlayer.sprx",
	SysmoduleContentExport:             "libSceContentExport.sprx",
	SysmoduleAudio3d:                   "libSceAudio3d.sprx",
	SysmoduleNpCommerce:                "libSceNpCommerce.sprx",
	SysmoduleMouse:                     "libSceMouse.sprx",
	SysmoduleCompanionHttpd:            "libSceCompanionHttpd.sprx",
	SysmoduleWebBrowserDialog:          "libSceWebBrowserDialog.sprx",
	SysmoduleErrorDialog:               "libSceErrorDialog.sprx",
	SysmoduleNpTrophy:                  "libSceNpTrophy.sprx",
	SysmoduleVideoCoreIf:               "libSceVideoCoreInterface.sprx",
	SysmoduleVideoCoreServerIf:         "libSceVideoCoreServerInterface.sprx",
	SysmoduleNpSnsFacebook:             "libSceNpSnsFacebookDialog.sprx",
	SysmoduleMoveTracker:               "libSceMoveTracker.sprx",
	SysmoduleNpProfileDialog:           "libSceNpProfileDialog.sprx",
	SysmoduleNpFriendListDialog:        "libSceNpFriendListDialog.sprx",
	SysmoduleAppContent:                "libSceAppContent.sprx",
	SysmoduleNpSignaling:               "libSceNpSignaling.sprx",
	SysmoduleRemotePlay:                "libSceRemoteplay.sprx",
	SysmoduleUsbd:                      "libSceUsbd.sprx",
	SysmoduleGameCustomDataDialog:      "libSceGameCustomDataDialog.sprx",
	SysmoduleNpEulaDialog:              "libSceNpEulaDialog.sprx",
	SysmoduleRandom:                    "libSceRandom.sprx",
	SysmoduleM4aacEnc:                  "libSceM4aacEnc.sprx",
	SysmoduleAudiodecCpu:               "libSceAudiodecCpu.sprx",
	SysmoduleAudiodecCpuDdp:            "libSceAudiodecCpuDdp.sprx",
	SysmoduleAudiodecCpuM4aac:          "libSceAudiodecCpuM4aac.sprx",
	SysmoduleBemp2Sys:                  "libSceBemp2sys.sprx",
	SysmoduleBeisobmf:                  "libSceBeisobmf.sprx",
	SysmodulePlayReady:                 "libScePlayReady.sprx",
	SysmoduleVideoNativeExtEssential:   "libSceVideoNativeExtEssential.sprx",
	SysmoduleZlib:                      "libSceZlib.sprx",
	SysmoduleDtcpIp:                    "libSceDtcpIp.sprx",
	SysmoduleContentSearch:             "libSceContentSearch.sprx",
	SysmoduleShareUtility:              "libSceShareUtility.sprx",
	SysmoduleAudiodecCpuDtsHdLbr:       "libSceAudiodecCpuDtsHdLbr.sprx",
	SysmoduleHeadTracker:               "libSceHeadTracker.sprx",
	SysmoduleGameUpdate:                "libSceGameUpdate.sprx",
	SysmoduleAutoMounterClient:         "libSceAutoMounterClient.sprx",
	SysmoduleSystemGesture:             "libSceSystemGesture.sprx",
	SysmoduleVideodec2:                 "libSceVideodec2.sprx",
	SysmoduleVdecwrap:                  "libSceVdecwrap.sprx",
	SysmoduleAt9Enc:                    "libSceAt9Enc.sprx",
	SysmoduleConvertKeycode:            "libSceConvertKeycode.sprx",
	SysmoduleSharePlay:                 "libSceSharePlay.sprx",
	SysmoduleHmd:                       "libSceHmd.sprx",
	SysmoduleUsbStorage:                "libSceUsbStorage.sprx",
	SysmoduleUsbStorageDialog:          "libSceUsbStorageDialog.sprx",
	SysmoduleDiscMap:                   "libSceDiscMap.sprx",
	SysmoduleFaceTracker:               "libSceFaceTracker.sprx",
	SysmoduleHandTracker:               "libSceHandTracker.sprx",
	SysmoduleNpSnsYoutubeDialog:        "libSceNpSnsYouTubeDialog.sprx",
	SysmoduleProfileCacheExternal:      "libSceProfileCacheExternal.sprx",
	SysmoduleMusicPlayerService:        "libSceMusicPlayerService.sprx",
	SysmoduleSpSysCallWrapper:          "libSceSpSysCallWrapper.sprx",
	SysmodulePs2EmuMenuDialog:          "libScePs2EmuMenuDialog.sprx",
	SysmoduleNpSnsDailymotionDialog:    "libSceNpSnsDailyMotionDialog.sprx",
	SysmoduleAudiodecCpuHevag:          "libSceAudiodecCpuHevag.sprx",
	SysmoduleLoginDialog:               "libSceLoginDialog.sprx",
	SysmoduleLoginService:              "libSceLoginService.sprx",
	SysmoduleSigninDialog:              "libSceSigninDialog.sprx",
	SysmoduleVdecsw:                    "libSceVdecsw.sprx",
	SysmoduleCustomMusicCore:           "libSceCustomMusicCore.sprx",
	SysmoduleJson2:                     "libSceJson2.sprx",
	SysmoduleAudioLatencyEstimation:    "libSceAudioLatencyEstimation.sprx",
	SysmoduleWkFontConfig:              "libSceWkFontConfig.sprx",
	SysmoduleVorbisDec:                 "libSceVorbisDec.sprx",
	SysmoduleHmdSetupDialog:            "libSceHmdSetupDialog.sprx",
	SysmoduleVrTracker:                 "libSceVrTracker.sprx",
	SysmoduleContentDelete:             "libSceContentDelete.sprx",
	SysmoduleImeBackend:                "libSceImeBackend.sprx",
	SysmoduleNetCtlApDialog:            "libSceNetCtlApDialog.sprx",
	SysmodulePlaygoDialog:              "libScePlayGoDialog.sprx",
	SysmoduleSocialScreen:              "libSceSocialScreen.sprx",
	SysmoduleEditMp4:                   "libSceEditMp4.sprx",
	SysmodulePsmKitSystem:              "libScePsmKitSystem.sprx",
	SysmoduleTextToSpeech:              "libSceTextToSpeech.sprx",
	SysmoduleNpToolkit:                 "libSceNpToolkit.sprx",
	SysmoduleCustomMusicService:        "libSceCustomMusicService.sprx",
	SysmoduleClSysCallWrapper:          "libSceClSysCallWrapper.sprx",
	SysmoduleSystemLogger:              "libSceSystemLogger.sprx",
	SysmoduleBluetoothHid:              "libSceBluetoothHid.sprx",
	SysmoduleVideoDecoderArbitration:   "libSceVideoDecoderArbitration.sprx",
	SysmoduleVrServiceDialog:           "libSceVrServiceDialog.sprx",
	SysmoduleJobManager:                "libSceJobManager.sprx",
	SysmoduleShareFactoryUtil:          "libSceShareFactoryUtil.sprx",
	SysmoduleSocialScreenDialog:        "libSceSocialScreenDialog.sprx",
	SysmoduleNpSnsDialog:               "libSceNpSnsDialog.sprx",
	SysmoduleNpToolkit2:                "libSceNpToolkit2.sprx",
	SysmoduleSrcUtl:                    "libSceSrcUtl.sprx",
	SysmoduleDiscId:                    "libSceDiscId.sprx",
	SysmoduleNpUniversalDataSystem:     "libSceNpUniversalDataSystem.sprx",
	SysmoduleKeyboard:                  "libSceKeyboard.sprx",
	SysmoduleGic:                       "libSceGic.sprx",
	SysmodulePlayReady2:                "libScePlayReady2.sprx",
	SysmoduleCesCs:                     "libSceCesCs.sprx",
	SysmodulePlayerInvitationDialog:    "libScePlayerInvitationDialog.sprx",
	SysmoduleNpSessionSignaling:        "libSceNpSessionSignaling.sprx",
	SysmoduleNpEntitlementAccess:       "libSceNpEntitlementAccess.sprx",
	SysmoduleNpCppWebApi:               "libSceNpCppWebApi.sprx",
	SysmoduleHubAppUtil:                "libSceHubAppUtil.sprx",
	SysmoduleNpPartner001:              "libSceNpPartner001.sprx",
	SysmoduleFontGs:                    "libSceFontGs.sprx",
	SysmoduleFontGsm:                   "libSceFontGsm.sprx",
	SysmoduleNpPartnerSubscription:     "libSceNpPartnerSubscription.sprx",
	SysmoduleNpAuthAuthorizedAppDialog: "libSceNpAuthAuthorizedAppDialog.sprx",
}

func IsBootModule(name string) bool {
	return BootModulesMap[name]
}

var BootModulesMap = map[string]bool{
	"I18N.CJK.dll":                              true,
	"I18N.MidEast.dll":                          true,
	"I18N.Other.dll":                            true,
	"I18N.Rare.dll":                             true,
	"I18N.West.dll":                             true,
	"I18N.dll":                                  true,
	"JSC.Net.dll":                               true,
	"LoginMgrUIProcess":                         true,
	"LoginMgrWebProcess":                        true,
	"Microsoft.CSharp.dll":                      true,
	"Mono.CSharp.dll":                           true,
	"Mono.Data.Sqlite.dll":                      true,
	"Mono.Data.Tds.dll":                         true,
	"Mono.Security.dll":                         true,
	"NKNetworkProcess":                          true,
	"NKNetworkProcessWebApp":                    true,
	"NKUIProcess":                               true,
	"NKWebProcess":                              true,
	"NKWebProcessHeapLimited":                   true,
	"Newtonsoft.Json.PlayStation.dll":           true,
	"ReactNative.Components.Vsh.dll":            true,
	"ReactNative.Debug.DevSupport.dll":          true,
	"ReactNative.Modules.Vsh.Gct.Telemetry.dll": true,
	"ReactNative.Modules.Vsh.dll":               true,
	"ReactNative.PUI.dll":                       true,
	"ReactNative.Vsh.Common.dll":                true,
	"Sce.Facebook.CSSLayout.dll":                true,
	"Sce.PlayStation.BclExtensions.dll":         true,
	"Sce.PlayStation.Core.dll":                  true,
	"Sce.PlayStation.Ime.dll":                   true,
	"Sce.PlayStation.Json.dll":                  true,
	"Sce.PlayStation.Orbis.Speech.dll":          true,
	"Sce.PlayStation.Orbis.dll":                 true,
	"Sce.PlayStation.PUI.dll":                   true,
	"Sce.PlayStation.PUIPlatform.dll":           true,
	"Sce.Vsh.Accessor.Db.Notify.dll":            true,
	"Sce.Vsh.Accessor.Db.dll":                   true,
	"Sce.Vsh.Accessor.dll":                      true,
	"Sce.Vsh.AppDbWrapper.dll":                  true,
	"Sce.Vsh.AppInstUtilWrapper.dll":            true,
	"Sce.Vsh.AutoMounterWrapper.dll":            true,
	"Sce.Vsh.BackupRestoreUtil.dll":             true,
	"Sce.Vsh.CloudClient.dll":                   true,
	"Sce.Vsh.DataTransfer.dll":                  true,
	"Sce.Vsh.Db.Shared.dll":                     true,
	"Sce.Vsh.DbPreparationWrapper.dll":          true,
	"Sce.Vsh.DbRecoveryUtilityWrapper.dll":      true,
	"Sce.Vsh.ErrorDialogUtilWrapper.dll":        true,
	"Sce.Vsh.EventServiceWrapper.dll":           true,
	"Sce.Vsh.FileSelector.dll":                  true,
	"Sce.Vsh.FileSelectorAdvance.dll":           true,
	"Sce.Vsh.Friend.dll":                        true,
	"Sce.Vsh.GameListRetrieverWrapper.dll":      true,
	"Sce.Vsh.Gls.GlsSharedMediaView.dll":        true,
	"Sce.Vsh.Gls.NativeCall.dll":                true,
	"Sce.Vsh.GriefReportStorage.dll":            true,
	"Sce.Vsh.JsExtension.dll":                   true,
	"Sce.Vsh.KernelSysWrapper.dll":              true,
	"Sce.Vsh.LncUtilWrapper.dll":                true,
	"Sce.Vsh.Lx.dll":                            true,
	"Sce.Vsh.MarlinDownloaderWrapper.dll":       true,
	"Sce.Vsh.Messages.DbAccessLib.dll":          true,
	"Sce.Vsh.MimeType.dll":                      true,
	"Sce.Vsh.MorpheusUpdWrapper.dll":            true,
	"Sce.Vsh.MyGameList.dll":                    true,
	"Sce.Vsh.Np.AppInfo.dll":                    true,
	"Sce.Vsh.Np.Asm.dll":                        true,
	"Sce.Vsh.Np.AuCheck.dll":                    true,
	"Sce.Vsh.Np.Common.dll":                     true,
	"Sce.Vsh.Np.GameIntent.dll":                 true,
	"Sce.Vsh.Np.IdMapper.dll":                   true,
	"Sce.Vsh.Np.Manager.dll":                    true,
	"Sce.Vsh.Np.Papc.dll":                       true,
	"Sce.Vsh.Np.Pbtc.dll":                       true,
	"Sce.Vsh.Np.RifManager.dll":                 true,
	"Sce.Vsh.Np.ServiceChecker.dll":             true,
	"Sce.Vsh.Np.ServiceChecker2.dll":            true,
	"Sce.Vsh.Np.Sns.dll":                        true,
	"Sce.Vsh.Np.Tmdb.dll":                       true,
	"Sce.Vsh.Np.Trophy.dll":                     true,
	"Sce.Vsh.Np.Uds.dll":                        true,
	"Sce.Vsh.Np.Webapi.dll":                     true,
	"Sce.Vsh.Np.Webapi2.dll":                    true,
	"Sce.Vsh.Orbis.AbstractStorage.dll":         true,
	"Sce.Vsh.Orbis.Bgft.dll":                    true,
	"Sce.Vsh.Orbis.ContentManager.dll":          true,
	"Sce.Vsh.PartyCommon.dll":                   true,
	"Sce.Vsh.Passcode.dll":                      true,
	"Sce.Vsh.PatchCheckerClientWrapper.dll":     true,
	"Sce.Vsh.ProfileCache.dll":                  true,
	"Sce.Vsh.PsnMessageUtil.dll":                true,
	"Sce.Vsh.PsnUtil.dll":                       true,
	"Sce.Vsh.Registry.dll":                      true,
	"Sce.Vsh.RequestShareScreen.dll":            true,
	"Sce.Vsh.RequestShareStorageWrapper.dll":    true,
	"Sce.Vsh.RnpsAppMgrWrapper.dll":             true,
	"Sce.Vsh.SQLite.dll":                        true,
	"Sce.Vsh.SQLiteAux.dll":                     true,
	"Sce.Vsh.SessionInvitation.dll":             true,
	"Sce.Vsh.ShareGuideScene.dll":               true,
	"Sce.Vsh.ShareServerPostWrapper.dll":        true,
	"Sce.Vsh.ShellCoreUtilWrapper.dll":          true,
	"Sce.Vsh.Sticker.StickerLibAccessor.dll":    true,
	"Sce.Vsh.SysUtilWrapper.dll":                true,
	"Sce.Vsh.SyscallWrapper.dll":                true,
	"Sce.Vsh.SysfileUtilWrapper.dll":            true,
	"Sce.Vsh.SystemLoggerUtilWrapper.dll":       true,
	"Sce.Vsh.SystemLoggerWrapper.dll":           true,
	"Sce.Vsh.TeamChat.TeamChatAccessor.dll":     true,
	"Sce.Vsh.Theme.dll":                         true,
	"Sce.Vsh.UpdateServiceWrapper.dll":          true,
	"Sce.Vsh.UsbStorageScene.dll":               true,
	"Sce.Vsh.UserServiceWrapper.dll":            true,
	"Sce.Vsh.VideoPlayer.dll":                   true,
	"Sce.Vsh.VideoServiceWrapper.dll":           true,
	"Sce.Vsh.VoiceAndAgent.dll":                 true,
	"Sce.Vsh.VoiceMsg.VoiceMsgWrapper.dll":      true,
	"Sce.Vsh.VrEnvironment.dll":                 true,
	"Sce.Vsh.WebBrowser.dll":                    true,
	"Sce.Vsh.WebViewDialog.dll":                 true,
	"Sce.Vsh.Webbrowser.XdbWrapper.dll":         true,
	"Sce.Vsh.Webbrowser.XutilWrapper.dll":       true,
	"Sce.Vsh.dll":                               true,
	"ScePlayReady":                              true,
	"ScePlayReady2":                             true,
	"SecureUIProcess":                           true,
	"SecureWebProcess":                          true,
	"SlimGLServerProcess":                       true,
	"System.Collections.dll":                    true,
	"System.ComponentModel.Composition.dll":     true,
	"System.ComponentModel.DataAnnotations.dll": true,
	"System.Core.dll":                           true,
	"System.Data.Services.Client.dll":           true,
	"System.Data.dll":                           true,
	"System.IO.Compression.FileSystem.dll":      true,
	"System.IO.Compression.dll":                 true,
	"System.Net.Http.WebRequest.dll":            true,
	"System.Net.Http.dll":                       true,
	"System.Net.dll":                            true,
	"System.Numerics.dll":                       true,
	"System.Reactive.Core.dll":                  true,
	"System.Reactive.Interfaces.dll":            true,
	"System.Reactive.Linq.dll":                  true,
	"System.Resources.ResourceManager.dll":      true,
	"System.Runtime.Extensions.dll":             true,
	"System.Runtime.Serialization.dll":          true,
	"System.Runtime.dll":                        true,
	"System.ServiceModel.Internals.dll":         true,
	"System.ServiceModel.Web.dll":               true,
	"System.ServiceModel.dll":                   true,
	"System.Threading.Tasks.dll":                true,
	"System.Transactions.dll":                   true,
	"System.Web.Services.dll":                   true,
	"System.Windows.dll":                        true,
	"System.Xml.Linq.dll":                       true,
	"System.Xml.Serialization.dll":              true,
	"System.Xml.dll":                            true,
	"System.dll":                                true,
	"UIProcess":                                 true,
	"WebAppBundle":                              true,
	"WebBrowserUIProcess":                       true,
	"WebProcess":                                true,
	"WebProcessHTMLTile":                        true,
	"WebProcessHeapLimited":                     true,
	"WebProcessWebApp":                          true,
	"custom_video_core":                         true,
	"libNativeExtensions":                       true,
	"libReactNative.Modules.Vsh":                true,
	"libSceAbstractDailymotion":                 true,
	"libSceAbstractFacebook":                    true,
	"libSceAbstractLocal":                       true,
	"libSceAbstractStorage":                     true,
	"libSceAbstractTwitter":                     true,
	"libSceAbstractYoutube":                     true,
	"libSceAjm":                                 true,
	"libSceAppInstUtil":                         true,
	"libSceAsyncStorageInternal":                true,
	"libSceAudioIn":                             true,
	"libSceAudioOut":                            true,
	"libSceAvPlayerStreaming":                   true,
	"libSceAvSetting":                           true,
	"libSceAvcap":                               true,
	"libSceBackupRestoreUtil":                   true,
	"libSceBgft":                                true,
	"libSceCamera":                              true,
	"libSceCdlgUtilServer":                      true,
	"libSceCommonDialog":                        true,
	"libSceCompositeExt":                        true,
	"libSceDataTransfer":                        true,
	"libSceFacebook.Yoga":                       true,
	"libSceFios2":                               true,
	"libSceFreeTypeHinter":                      true,
	"libSceFreeTypeSubFunc":                     true,
	"libSceGLSlimClientVSH":                     true,
	"libSceGLSlimServerVSH":                     true,
	"libSceGifParser":                           true,
	"libSceGnmDriver":                           true,
	"libSceGnmDriverForNeoMode":                 true,
	"libSceGvMp4Parser":                         true,
	"libSceHidControl":                          true,
	"libSceHttp":                                true,
	"libSceHttp2":                               true,
	"libSceHttpCache":                           true,
	"libSceIduUtil":                             true,
	"libSceImageUtil":                           true,
	"libSceInjectedBundle":                      true,
	"libSceIpmi":                                true,
	"libSceJitBridge":                           true,
	"libSceJpegParser":                          true,
	"libSceJsc":                                 true,
	"libSceJscCompiler":                         true,
	"libSceKbEmulate":                           true,
	"libSceLibcInternal":                        true,
	"libSceLibreSsl":                            true,
	"libSceLibreSsl3":                           true,
	"libSceManxWtf":                             true,
	"libSceMbus":                                true,
	"libSceMetadataReaderWriter":                true,
	"libSceMusicCoreServerClient":               true,
	"libSceMusicCoreServerClientJsEx":           true,
	"libSceNKWeb":                               true,
	"libSceNKWebCdlgInjectedBundle":             true,
	"libSceNKWebKit":                            true,
	"libSceNKWebKitRequirements":                true,
	"libSceNet":                                 true,
	"libSceNetCtl":                              true,
	"libSceNpCommon":                            true,
	"libSceNpGameIntent":                        true,
	"libSceNpGriefReport":                       true,
	"libSceNpManager":                           true,
	"libSceNpSns":                               true,
	"libSceNpSnsDailymotionDialog":              true,
	"libSceNpWebApi":                            true,
	"libSceNpWebApi2":                           true,
	"libSceOpusCeltDec":                         true,
	"libSceOpusCeltEnc":                         true,
	"libSceOpusDec":                             true,
	"libSceOpusSilkEnc":                         true,
	"libSceOrbisCompat":                         true,
	"libSceOrbisCompatForVideoService":          true,
	"libScePad":                                 true,
	"libScePatchCheckerClient":                  true,
	"libScePigletv2VSH":                         true,
	"libScePngParser":                           true,
	"libScePosixForWebKit":                      true,
	"libScePrecompiledShaders":                  true,
	"libScePsm":                                 true,
	"libScePsmUtil":                             true,
	"libSceRazorCpu":                            true,
	"libSceRegMgr":                              true,
	"libSceRnpsAppMgr":                          true,
	"libSceRnpsInjectedBundle":                  true,
	"libSceRnpsNKInjectedBundle":                true,
	"libSceRtc":                                 true,
	"libSceSaveData":                            true,
	"libSceScm":                                 true,
	"libSceShellUIUtil":                         true,
	"libSceSsl":                                 true,
	"libSceSsl2":                                true,
	"libSceSysCore":                             true,
	"libSceSysUtil":                             true,
	"libSceSysmodule":                           true,
	"libSceSystemService":                       true,
	"libSceTtsCoreEnUs":                         true,
	"libSceTtsCoreJp":                           true,
	"libSceUpdateService":                       true,
	"libSceUserService":                         true,
	"libSceVdecCore":                            true,
	"libSceVdecSavc":                            true,
	"libSceVdecSavc2":                           true,
	"libSceVdecShevc":                           true,
	"libSceVideoOut":                            true,
	"libSceVideoOutSecondary":                   true,
	"libSceVnaInternal":                         true,
	"libSceVnaWebsocket":                        true,
	"libSceVoiceQoS":                            true,
	"libSceWeb":                                 true,
	"libSceWebBrowserInjectedBundle":            true,
	"libSceWebCdlgInjectedBundle":               true,
	"libSceWebForVideoService":                  true,
	"libSceWebKit2":                             true,
	"libSceWebKit2ForVideoService":              true,
	"libSceWebKit2Secure":                       true,
	"libc":                                      true,
	"libcairo":                                  true,
	"libcurl":                                   true,
	"libfontconfig":                             true,
	"libfreetype":                               true,
	"libharfbuzz":                               true,
	"libicu":                                    true,
	"libkernel":                                 true,
	"libkernel_sys":                             true,
	"libkernel_web":                             true,
	"libmono-btls-shared":                       true,
	"libmono-profiler-log":                      true,
	"libmonosgen-2.0":                           true,
	"libpng16":                                  true,
	"libswctrl":                                 true,
	"libswreset":                                true,
	"mscorlib.dll":                              true,
	"orbis-jsc-compiler":                        true,
	"swagner":                                   true,
	"swreset":                                   true,
	"ulobjmgr":                                  true,
	"webapp":                                    true,
	"websocket-sharp.dll":                       true,
}
