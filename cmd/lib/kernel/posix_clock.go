package kernel

import (
	"time"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/time"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func libKernel_clock_gettime(clockId uint32, timestampPtr uintptr) int32 {
	if timestampPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid time pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("clock_gettime"),
		)
		SetErrno(EINVAL)
		return ERR_PTRI
	}

	now := time.Now()
	timestamp := (*Timestamp)(unsafe.Pointer(timestampPtr))
	timestamp.Seconds = now.Unix()
	timestamp.Nanoseconds = now.UnixNano()

	if logger.LogMisc {
		logger.Printf("%-132s %s returned %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("clock_gettime"),
			color.Yellow.Sprintf("0x%X", timestamp.Seconds),
		)
	}
	return 0
}

func libKernel_clock_gettimeofday(timevaluePtr, timezonePtr uintptr) int32 {
	now := time.Now()
	if timevaluePtr != 0 {
		timevalue := (*Timevalue)(unsafe.Pointer(timevaluePtr))
		timevalue.Seconds = now.Unix()
		timevalue.Microseconds = now.UnixMicro()
	}
	if timezonePtr != 0 {
		_, offset := now.Zone()
		timezone := (*Timezone)(unsafe.Pointer(timezonePtr))
		timezone.MinutesWest = int32(-offset / 60)
		if now.IsDST() {
			timezone.DstTime = 1
		} else {
			timezone.DstTime = 0
		}
	}

	if logger.LogMisc {
		logger.Printf("%-132s %s returned %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("clock_gettimeofday"),
			color.Yellow.Sprintf("0x%X", now.Unix()),
		)
	}
	return 0
}
