package lib

import (
	"unsafe"

	. "github.com/LamkasDev/sharkie/cmd/structs/net"
)

// 0x0000000000003380
// __int64 __fastcall sceNetCtlGetInfo(unsigned int, _BYTE *)
func libSceNetCtl_sceNetCtlGetInfo() uintptr {
	return 0
}

// 0x00000000000031A0
// __int64 __fastcall sceNetCtlGetResult(unsigned int, __int64)
func libSceNetCtl_sceNetCtlGetResult(eventType, errorCodePtr uintptr) uintptr {
	if errorCodePtr == 0 {
		return SCE_NET_CTL_ERROR_INVALID_ADDR
	}
	errorCodeSlice := unsafe.Slice((*int)(unsafe.Pointer(errorCodePtr)), 1)
	errorCodeSlice[0] = 0

	return 0
}

// 0x0000000000003200
// __int64 __fastcall sceNetCtlGetState(__int64)
func libSceNetCtl_sceNetCtlGetState(statePtr uintptr) uintptr {
	if statePtr == 0 {
		return SCE_NET_CTL_ERROR_INVALID_ADDR
	}
	stateSlice := unsafe.Slice((*int)(unsafe.Pointer(statePtr)), 1)
	stateSlice[0] = 0

	return 0
}

// 0x0000000000001DF0
// __int64 __fastcall sceNetCtlRegisterCallback(__int64, __int64, unsigned int *)
func libSceNetCtl_sceNetCtlRegisterCallback() uintptr {
	return 0
}

// 0x0000000000002430
// __int64 sceNetCtlCheckCallback()
func libSceNetCtl_sceNetCtlCheckCallback() uintptr {
	return 0
}
