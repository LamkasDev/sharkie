package kernel

// 0x0000000000012640
// __int64 __fastcall sceCoredumpRegisterCoredumpHandler(__int64 (__fastcall *)(_QWORD), __int64, __int64)
func libKernel_sceCoredumpRegisterCoredumpHandler(handler uintptr, stackSize uint64, commonPtr uintptr) uintptr {
	return 0
}

// 0x00000000000129E0
// __int64 sceCoredumpUnregisterCoredumpHandler()
func libKernel_sceCoredumpUnregisterCoredumpHandler() uintptr {
	return 0
}
