package lib

import (
	"time"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
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
	timestamp.Seconds = uint64(now.Unix())
	timestamp.Nanoseconds = uint64(now.Nanosecond())

	if logger.LogMisc {
		logger.Printf("%-132s %s returned %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("clock_gettime"),
			color.Yellow.Sprintf("0x%X", timestamp.Seconds),
		)
	}
	return 0
}
