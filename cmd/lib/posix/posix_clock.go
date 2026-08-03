package posix

import (
	"time"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/time"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func Clock_getres(clockId ClockId, timestamp *Timestamp) uintptr {
	return libScePosix_clock_getres(clockId, timestamp)
}

func libScePosix_clock_getres(clockId ClockId, timestamp *Timestamp) uintptr {
	if timestamp == nil {
		logger.Printf("%-132s %s failed due to invalid time pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("clock_getres"),
		)
		emu.SetErrno(EFAULT)
		return ERR_PTR
	}

	// Handle network clocks.
	switch clockId {
	case ClockIdExtNetwork, ClockIdExtAdNetwork, ClockIdExtRawNetwork, ClockIdExtDebugNetwork:
		logger.Printf("%-132s %s skipping unsupported clock %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("clock_getres"),
			color.Green.Sprint(clockId),
		)
		clockId = ClockIdMonotonic
	}

	// Get resolution based on clock type.
	var sec int64
	var nsec int64
	switch clockId {
	case ClockIdSecond, ClockIdRealtimeFast:
		sec = 0
		nsec = 1000000
	case ClockIdRealtime, ClockIdRealtimePrecise, ClockIdUptime, ClockIdUptimePrecise,
		ClockIdMonotonic, ClockIdMonotonicPrecise, ClockIdUptimeFast, ClockIdMonotonicFast:
		sec = 0
		nsec = 1
	default:
		logger.Printf("%-132s %s failed due to invalid clock %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("clock_getres"),
			color.Green.Sprint(clockId),
		)
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}
	timestamp.Seconds = sec
	timestamp.Nanoseconds = nsec

	if logger.LogMisc {
		logger.Printf("%-132s %s returned resolution for clock %s\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("clock_getres"),
			color.Green.Sprint(clockId),
		)
	}
	return 0
}

func Clock_gettime(clockId ClockId, timestamp *Timestamp) uintptr {
	return libScePosix_clock_gettime(clockId, timestamp)
}

func libScePosix_clock_gettime(clockId ClockId, timestamp *Timestamp) uintptr {
	if timestamp == nil {
		logger.Printf("%-132s %s failed due to invalid time pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("clock_gettime"),
		)
		emu.SetErrno(EINVAL)
		return ERR_PTR
	}

	now := time.Now()
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

func Gettimeofday(timevalue *Timevalue, timezone *Timezone) uintptr {
	return libScePosix_gettimeofday(timevalue, timezone)
}

func libScePosix_gettimeofday(timevalue *Timevalue, timezone *Timezone) uintptr {
	now := time.Now()
	if timevalue != nil {
		timevalue.Seconds = now.Unix()
		timevalue.Microseconds = now.UnixMicro()
	}
	if timezone != nil {
		_, offset := now.Zone()
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
			color.Magenta.Sprint("gettimeofday"),
			color.Yellow.Sprintf("0x%X", now.Unix()),
		)
	}
	return 0
}
