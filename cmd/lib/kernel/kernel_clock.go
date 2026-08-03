package kernel

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/lib/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/time"
)

func SceKernelClockGettime(clockId ClockId, timestamp *Timestamp) uintptr {
	return libKernel_sceKernelClockGettime(clockId, timestamp)
}

// 0x0000000000014CB0
// __int64 __fastcall sceKernelClockGettime(__int64, __int64)
func libKernel_sceKernelClockGettime(clockId ClockId, timestamp *Timestamp) uintptr {
	err := posix.Clock_gettime(clockId, timestamp)
	if err != 0 {
		return emu.GetErrno() - SonyErrorOffset
	}

	return 0
}

func SceKernelClockGettimezone(timezone *Timezone) uintptr {
	return libKernel_sceKernelClockGettimezone(timezone)
}

// 0x0000000000014D20
// __int64 __fastcall sceKernelGettimezone(__int64)
func libKernel_sceKernelClockGettimezone(timezone *Timezone) uintptr {
	err := posix.Gettimeofday(nil, timezone)
	if err != 0 {
		return emu.GetErrno() - SonyErrorOffset
	}

	return 0
}
