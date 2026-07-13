package kernel

import (
	"time"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/lib/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/time"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

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
	return 0
}

// 0x0000000000014CE0
// __int64 __fastcall sceKernelGettimeofday(__int64)
func libKernel_sceKernelGettimeofday(timevaluePtr uintptr) uintptr {
	err := posix.Clock_gettimeofday(timevaluePtr, 0)
	if err != 0 {
		return emu.GetErrno() - SonyErrorOffset
	}

	return 0
}

func SceKernelConvertUtcToLocaltime(utcTime int64, localTimePtr, timesecPtr, dstSecPtr uintptr) uintptr {
	return libKernel_sceKernelConvertUtcToLocaltime(utcTime, localTimePtr, timesecPtr, dstSecPtr)
}

// 0x00000000000151D0
// __int64 __fastcall sceKernelConvertUtcToLocaltime(__int64, _QWORD *, __int64, _DWORD *)
func libKernel_sceKernelConvertUtcToLocaltime(utcTime int64, localTimePtr, timesecPtr, dstSecPtr uintptr) uintptr {
	var timezone Timezone
	err := posix.Clock_gettimeofday(0, uintptr(unsafe.Pointer(&timezone)))
	if err != 0 {
		return emu.GetErrno() - SonyErrorOffset
	}

	localTime := utcTime + 60*(int64(timezone.MinutesWest)+int64(timezone.DstTime))
	if localTimePtr != 0 {
		*(*int64)(unsafe.Pointer(localTimePtr)) = localTime
	}
	if timesecPtr != 0 {
		st := (*Timesec)(unsafe.Pointer(timesecPtr))
		st.UtcTime = utcTime
		st.WestSec = 60 * timezone.MinutesWest
		st.DstSec = 60 * timezone.DstTime
	}
	if dstSecPtr != 0 {
		*(*uint64)(unsafe.Pointer(dstSecPtr)) = uint64(60 * timezone.DstTime)
	}

	return 0
}
