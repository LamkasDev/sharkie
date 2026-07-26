package kernel

import (
	"time"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/lib/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/time"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func SceKernelGetProcessTime() uintptr {
	return libKernel_sceKernelGetProcessTime()
}

// 0x0000000000014D50
// __int64 sceKernelGetProcessTime()
func libKernel_sceKernelGetProcessTime() uintptr {
	elapsed := time.Since(TscStartTime)
	micros := uintptr(elapsed.Microseconds())

	if logger.LogMisc {
		logger.Printf("%-132s %s returned %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelGetProcessTime"),
			color.Yellow.Sprintf("0x%X", micros),
		)
	}
	return micros
}

// 0x0000000000014CE0
// __int64 __fastcall sceKernelGettimeofday(__int64)
func libKernel_sceKernelGettimeofday(timevalue *Timevalue) uintptr {
	err := posix.Clock_gettimeofday(timevalue, nil)
	if err != 0 {
		return emu.GetErrno() - SonyErrorOffset
	}

	return 0
}

func SceKernelConvertUtcToLocaltime(utcTime int64, localTimePtr *int64, timesec *Timesec, dstSecPtr *uint64) uintptr {
	return libKernel_sceKernelConvertUtcToLocaltime(utcTime, localTimePtr, timesec, dstSecPtr)
}

// 0x00000000000151D0
// __int64 __fastcall sceKernelConvertUtcToLocaltime(__int64, _QWORD *, __int64, _DWORD *)
func libKernel_sceKernelConvertUtcToLocaltime(utcTime int64, localTimePtr *int64, timesec *Timesec, dstSecPtr *uint64) uintptr {
	var timezone Timezone
	err := posix.Clock_gettimeofday(nil, &timezone)
	if err != 0 {
		return emu.GetErrno() - SonyErrorOffset
	}

	localTime := utcTime + 60*(int64(timezone.MinutesWest)+int64(timezone.DstTime))
	if localTimePtr != nil {
		*localTimePtr = localTime
	}
	if timesec != nil {
		timesec.UtcTime = utcTime
		timesec.WestSec = 60 * timezone.MinutesWest
		timesec.DstSec = 60 * timezone.DstTime
	}
	if dstSecPtr != nil {
		*dstSecPtr = uint64(60 * timezone.DstTime)
	}

	return 0
}
