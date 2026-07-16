package module

const (
	SysmoduleInvalid                   = 0x0000
	SysmoduleFiber                     = 0x0006 // libSceFiber.sprx
	SysmoduleUlt                       = 0x0007 // libSceUlt.sprx
	SysmoduleNgs2                      = 0x000B // libSceNgs2.sprx
	SysmoduleXml                       = 0x0017 // libSceXml.sprx
	SysmoduleNpUtility                 = 0x0019 // libSceNpUtility.sprx
	SysmoduleVoice                     = 0x001A // libSceVoice.sprx
	SysmoduleVoiceqos                  = 0x001B // libSceVoiceQos.sprx
	SysmoduleNpMatching2               = 0x001C // libSceNpMatching2.sprx
	SysmoduleNpScoreRanking            = 0x001E // libSceNpScoreRanking.sprx
	SysmoduleRudp                      = 0x0021 // libSceRudp.sprx
	SysmoduleNpTus                     = 0x002C // libSceNpTus.sprx
	SysmoduleFace                      = 0x0038 // libSceFace.sprx
	SysmoduleSmart                     = 0x0039 // libSceSmart.sprx
	SysmoduleJson                      = 0x0080 // libSceJson.sprx
	SysmoduleGameLiveStreaming         = 0x0081 // libSceGameLiveStreaming.sprx
	SysmoduleCompanionUtil             = 0x0082 // libSceCompanionUtil.sprx
	SysmodulePlaygo                    = 0x0083 // libScePlayGo.sprx
	SysmoduleFont                      = 0x0084 // libSceFont.sprx
	SysmoduleVideoRecording            = 0x0085 // libSceVideoRecording.sprx
	SysmoduleS3dconversion             = 0x0086 // libSceS3DConversion.sprx
	SysmoduleAudiodec                  = 0x0088 // libSceAudiodec.sprx
	SysmoduleJpegDec                   = 0x008A // libSceJpegDec.sprx
	SysmoduleJpegEnc                   = 0x008B // libSceJpegEnc.sprx
	SysmodulePngDec                    = 0x008C // libScePngDec.sprx
	SysmodulePngEnc                    = 0x008D // libScePngEnc.sprx
	SysmoduleVideodec                  = 0x008E // libSceVideodec.sprx
	SysmoduleMove                      = 0x008F // libSceMove.sprx
	SysmodulePadTracker                = 0x0091 // libScePadTracker.sprx
	SysmoduleDepth                     = 0x0092 // libSceDepth.sprx
	SysmoduleHand                      = 0x0093 // libSceHand.sprx
	SysmoduleLibime                    = 0x0095 // libSceIme.sprx
	SysmoduleImeDialog                 = 0x0096 // libSceImeDialog.sprx
	SysmoduleNpParty                   = 0x0097 // libSceNpParty.sprx
	SysmoduleFontFt                    = 0x0098 // libSceFontFt.sprx
	SysmoduleFreetypeOt                = 0x0099 // libSceFreeTypeOt.sprx
	SysmoduleFreetypeOl                = 0x009A // libSceFreeTypeOl.sprx
	SysmoduleFreetypeOptOl             = 0x009B // libSceFreeTypeOptOl.sprx
	SysmoduleScreenShot                = 0x009C // libSceScreenShot.sprx
	SysmoduleNpAuth                    = 0x009D // libSceNpAuth.sprx
	SysmoduleSulpha                    = 0x009F
	SysmoduleSaveDataDialog            = 0x00A0 // libSceSaveDataDialog.sprx
	SysmoduleInvitationDialog          = 0x00A2 // libSceInvitationDialog.sprx
	SysmoduleDebugKeyboard             = 0x00A3
	SysmoduleMessageDialog             = 0x00A4 // libSceMsgDialog.sprx
	SysmoduleAvPlayer                  = 0x00A5 // libSceAvPlayer.sprx
	SysmoduleContentExport             = 0x00A6 // libSceContentExport.sprx
	SysmoduleAudio3d                   = 0x00A7 // libSceAudio3d.sprx
	SysmoduleNpCommerce                = 0x00A8 // libSceNpCommerce.sprx
	SysmoduleMouse                     = 0x00A9 // libSceMouse.sprx
	SysmoduleCompanionHttpd            = 0x00AA // libSceCompanionHttpd.sprx
	SysmoduleWebBrowserDialog          = 0x00AB // libSceWebBrowserDialog.sprx
	SysmoduleErrorDialog               = 0x00AC // libSceErrorDialog.sprx
	SysmoduleNpTrophy                  = 0x00AD // libSceNpTrophy.sprx
	SysmoduleVideoCoreIf               = 0x00AE // libSceVideoCoreInterface.sprx
	SysmoduleVideoCoreServerIf         = 0x00AF // libSceVideoCoreServerInterface.sprx
	SysmoduleNpSnsFacebook             = 0x00B0 // libSceNpSnsFacebookDialog.sprx
	SysmoduleMoveTracker               = 0x00B1 // libSceMoveTracker.sprx
	SysmoduleNpProfileDialog           = 0x00B2 // libSceNpProfileDialog.sprx
	SysmoduleNpFriendListDialog        = 0x00B3 // libSceNpFriendListDialog.sprx
	SysmoduleAppContent                = 0x00B4 // libSceAppContent.sprx
	SysmoduleNpSignaling               = 0x00B5 // libSceNpSignaling.sprx
	SysmoduleRemotePlay                = 0x00B6 // libSceRemoteplay.sprx
	SysmoduleUsbd                      = 0x00B7 // libSceUsbd.sprx
	SysmoduleGameCustomDataDialog      = 0x00B8 // libSceGameCustomDataDialog.sprx
	SysmoduleNpEulaDialog              = 0x00B9 // libSceNpEulaDialog.sprx
	SysmoduleRandom                    = 0x00BA // libSceRandom.sprx
	SysmoduleReserved2                 = 0x00BB
	SysmoduleM4aacEnc                  = 0x00BC // libSceM4aacEnc.sprx
	SysmoduleAudiodecCpu               = 0x00BD // libSceAudiodecCpu.sprx
	SysmoduleAudiodecCpuDdp            = 0x00BE // libSceAudiodecCpuDdp.sprx
	SysmoduleAudiodecCpuM4aac          = 0x00C0 // libSceAudiodecCpuM4aac.sprx
	SysmoduleBemp2Sys                  = 0x00C1 // libSceBemp2sys.sprx
	SysmoduleBeisobmf                  = 0x00C2 // libSceBeisobmf.sprx
	SysmodulePlayReady                 = 0x00C3 // libScePlayReady.sprx
	SysmoduleVideoNativeExtEssential   = 0x00C4 // libSceVideoNativeExtEssential.sprx
	SysmoduleZlib                      = 0x00C5 // libSceZlib.sprx
	SysmoduleDtcpIp                    = 0x00C6 // libSceDtcpIp.sprx
	SysmoduleContentSearch             = 0x00C7 // libSceContentSearch.sprx
	SysmoduleShareUtility              = 0x00C8 // libSceShareUtility.sprx
	SysmoduleAudiodecCpuDtsHdLbr       = 0x00C9 // libSceAudiodecCpuDtsHdLbr.sprx
	SysmoduleDeci4h                    = 0x00CA
	SysmoduleHeadTracker               = 0x00CB // libSceHeadTracker.sprx
	SysmoduleGameUpdate                = 0x00CC // libSceGameUpdate.sprx
	SysmoduleAutoMounterClient         = 0x00CD // libSceAutoMounterClient.sprx
	SysmoduleSystemGesture             = 0x00CE // libSceSystemGesture.sprx
	SysmoduleVideodec2                 = 0x00CF // libSceVideodec2.sprx
	SysmoduleVdecwrap                  = 0x00D0 // libSceVdecwrap.sprx
	SysmoduleAt9Enc                    = 0x00D1 // libSceAt9Enc.sprx
	SysmoduleConvertKeycode            = 0x00D2 // libSceConvertKeycode.sprx
	SysmoduleSharePlay                 = 0x00D3 // libSceSharePlay.sprx
	SysmoduleHmd                       = 0x00D4 // libSceHmd.sprx
	SysmoduleUsbStorage                = 0x00D5 // libSceUsbStorage.sprx
	SysmoduleUsbStorageDialog          = 0x00D6 // libSceUsbStorageDialog.sprx
	SysmoduleDiscMap                   = 0x00D7 // libSceDiscMap.sprx
	SysmoduleFaceTracker               = 0x00D8 // libSceFaceTracker.sprx
	SysmoduleHandTracker               = 0x00D9 // libSceHandTracker.sprx
	SysmoduleNpSnsYoutubeDialog        = 0x00DA // libSceNpSnsYouTubeDialog.sprx
	SysmoduleProfileCacheExternal      = 0x00DC // libSceProfileCacheExternal.sprx
	SysmoduleMusicPlayerService        = 0x00DD // libSceMusicPlayerService.sprx
	SysmoduleSpSysCallWrapper          = 0x00DE // libSceSpSysCallWrapper.sprx
	SysmodulePs2EmuMenuDialog          = 0x00DF // libScePs2EmuMenuDialog.sprx
	SysmoduleNpSnsDailymotionDialog    = 0x00E0 // libSceNpSnsDailyMotionDialog.sprx
	SysmoduleAudiodecCpuHevag          = 0x00E1 // libSceAudiodecCpuHevag.sprx
	SysmoduleLoginDialog               = 0x00E2 // libSceLoginDialog.sprx
	SysmoduleLoginService              = 0x00E3 // libSceLoginService.sprx
	SysmoduleSigninDialog              = 0x00E4 // libSceSigninDialog.sprx
	SysmoduleVdecsw                    = 0x00E5 // libSceVdecsw.sprx
	SysmoduleCustomMusicCore           = 0x00E6 // libSceCustomMusicCore.sprx
	SysmoduleJson2                     = 0x00E7 // libSceJson2.sprx
	SysmoduleAudioLatencyEstimation    = 0x00E8 // libSceAudioLatencyEstimation.sprx
	SysmoduleWkFontConfig              = 0x00E9 // libSceWkFontConfig.sprx
	SysmoduleVorbisDec                 = 0x00EA // libSceVorbisDec.sprx
	SysmoduleHmdSetupDialog            = 0x00EB // libSceHmdSetupDialog.sprx
	SysmoduleReserved28                = 0x00EC
	SysmoduleVrTracker                 = 0x00ED // libSceVrTracker.sprx
	SysmoduleContentDelete             = 0x00EE // libSceContentDelete.sprx
	SysmoduleImeBackend                = 0x00EF // libSceImeBackend.sprx
	SysmoduleNetCtlApDialog            = 0x00F0 // libSceNetCtlApDialog.sprx
	SysmodulePlaygoDialog              = 0x00F1 // libScePlayGoDialog.sprx
	SysmoduleSocialScreen              = 0x00F2 // libSceSocialScreen.sprx
	SysmoduleEditMp4                   = 0x00F3 // libSceEditMp4.sprx
	SysmodulePsmKitSystem              = 0x00F5 // libScePsmKitSystem.sprx
	SysmoduleTextToSpeech              = 0x00F6 // libSceTextToSpeech.sprx
	SysmoduleNpToolkit                 = 0x00F7 // libSceNpToolkit.sprx
	SysmoduleCustomMusicService        = 0x00F8 // libSceCustomMusicService.sprx
	SysmoduleClSysCallWrapper          = 0x00F9 // libSceClSysCallWrapper.sprx
	SysmoduleSystemLogger              = 0x00FA // libSceSystemLogger.sprx
	SysmoduleBluetoothHid              = 0x00FB // libSceBluetoothHid.sprx
	SysmoduleVideoDecoderArbitration   = 0x00FC // libSceVideoDecoderArbitration.sprx
	SysmoduleVrServiceDialog           = 0x00FD // libSceVrServiceDialog.sprx
	SysmoduleJobManager                = 0x00FE // libSceJobManager.sprx
	SysmoduleShareFactoryUtil          = 0x00FF // libSceShareFactoryUtil.sprx
	SysmoduleSocialScreenDialog        = 0x0100 // libSceSocialScreenDialog.sprx
	SysmoduleNpSnsDialog               = 0x0101 // libSceNpSnsDialog.sprx
	SysmoduleNpToolkit2                = 0x0102 // libSceNpToolkit2.sprx
	SysmoduleSrcUtl                    = 0x0103 // libSceSrcUtl.sprx
	SysmoduleDiscId                    = 0x0104 // libSceDiscId.sprx
	SysmoduleNpUniversalDataSystem     = 0x0105 // libSceNpUniversalDataSystem.sprx
	SysmoduleKeyboard                  = 0x0106 // libSceKeyboard.sprx
	SysmoduleGic                       = 0x0107 // libSceGic.sprx
	SysmodulePlayReady2                = 0x0108 // libScePlayReady2.sprx
	SysmoduleCesCs                     = 0x010C // libSceCesCs.sprx
	SysmodulePlayerInvitationDialog    = 0x010D // libScePlayerInvitationDialog.sprx
	SysmoduleNpSessionSignaling        = 0x0112 // libSceNpSessionSignaling.sprx
	SysmoduleNpEntitlementAccess       = 0x0113 // libSceNpEntitlementAccess.sprx
	SysmoduleNpCppWebApi               = 0x0115 // libSceNpCppWebApi.sprx
	SysmoduleHubAppUtil                = 0x0116 // libSceHubAppUtil.sprx
	SysmoduleNpPartner001              = 0x011A // libSceNpPartner001.sprx
	SysmoduleFontGs                    = 0x012F // libSceFontGs.sprx
	SysmoduleFontGsm                   = 0x0135 // libSceFontGsm.sprx
	SysmoduleNpPartnerSubscription     = 0x0138 // libSceNpPartnerSubscription.sprx
	SysmoduleNpAuthAuthorizedAppDialog = 0x0139 // libSceNpAuthAuthorizedAppDialog.sprx
)

const (
	SysmoduleInternalRazorCpu = 0x80000019 // libSceRazorCpu.sprx
)

var SysmoduleMap = map[uint16]string{
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
