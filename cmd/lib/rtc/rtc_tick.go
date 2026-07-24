package rtc

import "C"
import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/lib/kernel"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/time"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000000280
// __int64 __fastcall sceRtcGetCurrentTick(_QWORD *)
func libSceRtc_sceRtcGetCurrentTick(tickPtr uintptr) uintptr {
	if tickPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid tick pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceRtcGetCurrentTick"),
		)
		return 0x7FFEF9FE
	}
	tick := (*RtcTick)(unsafe.Pointer(tickPtr))

	var timestamp Timestamp
	err := kernel.SceKernelClockGettime(0, uintptr(unsafe.Pointer(&timestamp)))
	if err == 0 {
		epochMicros := (1000000 * timestamp.Seconds) + (timestamp.Nanoseconds / 1000)
		tick.Tick = uint64(epochMicros + UnixEpochTicks)
	}

	return 0
}

// 0x00000000000006F0
// __int64 __fastcall sceRtcGetCurrentRawNetworkTick(_QWORD *)
func libSceRtc_sceRtcGetCurrentRawNetworkTick(tickPtr uintptr) uintptr {
	if tickPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid tick pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceRtcGetCurrentRawNetworkTick"),
		)
		return 0x7FFEF9FE
	}
	tick := (*RtcTick)(unsafe.Pointer(tickPtr))

	var timestamp Timestamp
	err := kernel.SceKernelClockGettime(0, uintptr(unsafe.Pointer(&timestamp)))
	if err == 0 {
		epochMicros := (1000000 * timestamp.Seconds) + (timestamp.Nanoseconds / 1000)
		tick.Tick = uint64(epochMicros + UnixEpochTicks)
	}

	return 0
}

func SceRtcSetTick(datetimePtr, tickPtr uintptr) uintptr {
	return libSceRtc_sceRtcSetTick(datetimePtr, tickPtr)
}

// 0x0000000000002950
// __int64 __fastcall sceRtcSetTick(__int64, _QWORD *)
func libSceRtc_sceRtcSetTick(datetimePtr, tickPtr uintptr) uintptr {
	if datetimePtr == 0 || tickPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid date time or tick pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceRtcSetTick"),
		)
		return 0x80B50002
	}
	datetime := (*RtcDateTime)(unsafe.Pointer(datetimePtr))
	tick := (*RtcTick)(unsafe.Pointer(tickPtr)).Tick

	// Separate days and microseconds.
	days := tick / 86400000000
	micros := tick % 86400000000

	// Perform fast Gregorian Calendar conversion.
	days += 307
	j := (days << 2) - 1
	ly := j / 146097

	j -= 146097 * ly
	ld := j >> 2
	j = ((ld << 2) + 3) / 1461
	ld = (((ld << 2) + 7) - 1461*j) >> 2
	lm := (5*ld - 3) / 153
	ld = (5*ld + 2 - 153*lm) / 5
	ly = 100*ly + j

	if lm < 10 {
		lm += 3
	} else {
		lm -= 9
		ly++
	}

	// Assign to result.
	datetime.Year = uint16(ly)
	datetime.Month = uint16(lm)
	datetime.Day = uint16(ld)
	datetime.Hour = uint16(micros / 3600000000)
	micros %= 3600000000
	datetime.Minute = uint16(micros / 60000000)
	micros %= 60000000
	datetime.Second = uint16(micros / 1000000)
	datetime.Microsecond = uint32(micros % 1000000)

	return 0
}

// 0x0000000000002D10
// __int64 __fastcall sceRtcGetTick(int *, __int64)
func libSceRtc_sceRtcGetTick(datetimePtr, tickPtr uintptr) uintptr {
	if datetimePtr == 0 || tickPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid date time or tick pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceRtcGetTick"),
		)
		return 0x80B50002
	}
	datetime := (*RtcDateTime)(unsafe.Pointer(datetimePtr))
	tick := (*RtcTick)(unsafe.Pointer(tickPtr))

	// Validate date.
	err := libSceRtc_sceRtcCheckValid(datetimePtr)
	if err != 0 {
		return err
	}

	// Shift Jan/Feb to end of previous year.
	year := int64(datetime.Year)
	month := int64(datetime.Month)
	if month > 2 {
		month -= 3
	} else {
		month += 9
		year -= 1
	}

	// Perform fast Gregorian Calendar conversion.
	c := year / 100
	ya := year - 100*c
	days := ((146097 * c) >> 2) + ((1461 * ya) >> 2) + ((153*month + 2) / 5) + int64(datetime.Day)
	days -= 307

	// Convert days to microsecondS.
	totalTicks := days * 86400000000
	micros := int64(datetime.Hour)*3600000000 +
		int64(datetime.Minute)*60000000 +
		int64(datetime.Second)*1000000 +
		int64(datetime.Microsecond)

	// Assign to result.
	tick.Tick = uint64(totalTicks + micros)

	return 0
}

// 0x0000000000003370
// __int64 __fastcall sceRtcTickAddHours(_QWORD *, _QWORD *, int)
func libSceRtc_sceRtcTickAddHours(tick1Ptr, tick2Ptr uintptr, add int64) uintptr {
	if tick1Ptr == 0 || tick2Ptr == 0 {
		logger.Printf("%-132s %s failed due to invalid tick pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceRtcTickAddHours"),
		)
		return 0x80B50002
	}
	tick1 := (*RtcTick)(unsafe.Pointer(tick1Ptr))
	tick2 := (*RtcTick)(unsafe.Pointer(tick2Ptr))
	tick1.Tick = uint64(int64(tick2.Tick) + (add * 3600000000))

	return 0
}

func SceRtcTickAddMinutes(tick1Ptr, tick2Ptr uintptr, add int64) uintptr {
	return libSceRtc_sceRtcTickAddMinutes(tick1Ptr, tick2Ptr, add)
}

// 0x0000000000003300
// __int64 __fastcall sceRtcTickAddMinutes(_QWORD *, _QWORD *, __int64)
func libSceRtc_sceRtcTickAddMinutes(tick1Ptr, tick2Ptr uintptr, add int64) uintptr {
	if tick1Ptr == 0 || tick2Ptr == 0 {
		logger.Printf("%-132s %s failed due to invalid tick pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceRtcTickAddMinutes"),
		)
		return 0x80B50002
	}
	tick1 := (*RtcTick)(unsafe.Pointer(tick1Ptr))
	tick2 := (*RtcTick)(unsafe.Pointer(tick2Ptr))
	tick1.Tick = uint64(int64(tick2.Tick) + (add * 60000000))

	return 0
}

// 0x00000000000032E0
// __int64 __fastcall sceRtcTickAddSeconds(_QWORD *, _QWORD *, __int64)
func libSceRtc_sceRtcTickAddSeconds(tick1Ptr, tick2Ptr uintptr, add int64) uintptr {
	if tick1Ptr == 0 || tick2Ptr == 0 {
		logger.Printf("%-132s %s failed due to invalid tick pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceRtcTickAddSeconds"),
		)
		return 0x80B50002
	}
	tick1 := (*RtcTick)(unsafe.Pointer(tick1Ptr))
	tick2 := (*RtcTick)(unsafe.Pointer(tick2Ptr))
	tick1.Tick = uint64(int64(tick2.Tick) + (add * 1000000))

	return 0
}
