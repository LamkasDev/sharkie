package audio_in

// 0x00000000000003E0
// __int64 __fastcall sceAudioInOpen(unsigned int, int, int, int, int, unsigned int)
func libSceAudioIn_sceAudioInOpen() uintptr {
	return 0
}

// 0x00000000000007F0
// __int64 __fastcall sceAudioInOpenEx(unsigned int, int, int, unsigned int, int, int, unsigned int)
func libSceAudioIn_sceAudioInOpenEx() uintptr {
	return 0
}

// 0x00000000000016D0
// __int64 __fastcall sceAudioInGetHandleStatusInfo(int, __int64)
func libSceAudioIn_sceAudioInGetHandleStatusInfo() uintptr {
	return 0
}

// 0x0000000000001940
// __int64 __fastcall sceAudioInGetSilentState(int)
func libSceAudioIn_sceAudioInGetSilentState() uintptr {
	return 1
}
