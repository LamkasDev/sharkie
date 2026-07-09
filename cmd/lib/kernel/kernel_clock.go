package kernel

import (
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
)

func SceKernelClockGettime(clockId uint32, timestampPtr uintptr) uintptr {
	return libKernel_sceKernelClockGettime(clockId, timestampPtr)
}

// 0x0000000000014CB0
// __int64 __fastcall sceKernelClockGettime(__int64, __int64)
func libKernel_sceKernelClockGettime(clockId uint32, timestampPtr uintptr) uintptr {
	err := libKernel_clock_gettime(clockId, timestampPtr)
	if err != 0 {
		return GetErrno() - SonyErrorOffset
	}

	return 0
}

func SceKernelClockGettimezone(timezonePtr uintptr) uintptr {
	return libKernel_sceKernelClockGettimezone(timezonePtr)
}

// 0x0000000000014D20
// __int64 __fastcall sceKernelGettimezone(__int64)
func libKernel_sceKernelClockGettimezone(timezonePtr uintptr) uintptr {
	err := libKernel_clock_gettimeofday(0, timezonePtr)
	if err != 0 {
		return GetErrno() - SonyErrorOffset
	}

	return 0
}
