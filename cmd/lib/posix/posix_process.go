package posix

import (
	"time"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/time"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func Usleep(micros uint32) uintptr {
	return libScePosix_usleep(micros)
}

func libScePosix_usleep(micros uint32) uintptr {
	timestamp := Timestamp{
		Seconds:     int64(micros / 1_000_000),
		Nanoseconds: int64(micros%1_000_000) * 1000,
	}
	return libScePosix_nanosleep(&timestamp, nil)
}

func Nanosleep(timestamp, remainingTimestamp *Timestamp) uintptr {
	return libScePosix_nanosleep(timestamp, remainingTimestamp)
}

// TODO: make this interruptible
func libScePosix_nanosleep(timestamp, remainingTimestamp *Timestamp) uintptr {
	if timestamp == nil {
		logger.Printf("%-132s %s failed due to invalid timestamp pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("nanosleep"),
		)
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}
	if timestamp.Seconds < 0 || timestamp.Nanoseconds < 0 || timestamp.Nanoseconds > 1_000_000_000 {
		logger.Printf("%-132s %s failed due to invalid timestamp values.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("nanosleep"),
		)
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}

	timeout := time.Duration(timestamp.Seconds)*time.Second + time.Duration(timestamp.Nanoseconds)*time.Nanosecond
	if logger.LogSleep {
		logger.Printf("%-132s %s sleeping for %ss and %sns.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("nanosleep"),
			color.Yellow.Sprintf("0x%X", timestamp.Seconds),
			color.Yellow.Sprintf("0x%X", timestamp.Nanoseconds),
		)
	}
	time.Sleep(timeout)
	if remainingTimestamp != nil {
		remainingTimestamp.Seconds = 0
		remainingTimestamp.Nanoseconds = 0
	}

	return 0
}
