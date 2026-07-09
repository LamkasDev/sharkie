package rtc

import "C"
import (
	"strconv"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib/kernel"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/time"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func parseInt(s string) int {
	val, _ := strconv.Atoi(s)
	return val
}

// 0x00000000000003D0
// __int64 __fastcall sceRtcGetCurrentClockLocalTime(__int64)
func libSceRtc_sceRtcGetCurrentClockLocalTime(datetimePtr uintptr) uintptr {
	if datetimePtr == 0 {
		logger.Printf("%-132s %s failed due to invalid date time pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("__sys_regmgr_call"),
		)
		return 0x7FFEF9FE
	}

	// Get current UTC time.
	var timestamp Timestamp
	err := SceKernelClockGettime(0, uintptr(unsafe.Pointer(&timestamp)))
	if err != 0 {
		return err
	}
	epochMicros := (1000000 * timestamp.Seconds) + (timestamp.Nanoseconds / 1000)
	tick := RtcTick{
		Tick: uint64(epochMicros + UnixEpochTicks),
	}

	// Convert it to local time.
	var localTime int64
	var timesec Timesec
	err = SceKernelConvertUtcToLocaltime(epochMicros/1000000, uintptr(unsafe.Pointer(&localTime)), uintptr(unsafe.Pointer(&timesec)), 0)
	if err != 0 {
		return err
	}

	// Adjust precision.
	adjustment := (timestamp.Nanoseconds + (timestamp.Nanoseconds >> 32)) / 60
	SceRtcTickAddMinutes(uintptr(unsafe.Pointer(&tick)), uintptr(unsafe.Pointer(&tick)), adjustment)
	SceRtcSetTick(datetimePtr, uintptr(unsafe.Pointer(&tick)))

	return 0
}

// 0x0000000000000A70
// __int64 __fastcall sceRtcParseDateTime(_QWORD *, unsigned __int8 *)
func libSceRtc_sceRtcParseDateTime(tickUtcPtr uintptr, datetimeStrPtr Cstring) uintptr {
	if tickUtcPtr == 0 || datetimeStrPtr == nil {
		logger.Printf("%-132s %s failed due to invalid tick or date time string pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceRtcParseDateTime"),
		)
		return 0x80B50002
	}
	datetimeStr := GoString(datetimeStrPtr)
	formatKey := datetimeStr[19]

	if formatKey == 'Z' || formatKey == '-' || formatKey == '+' || formatKey == '.' {
		return libSceRtc_sceRtcParseRFC3339(tickUtcPtr, datetimeStrPtr)
	} else if formatKey == ':' {
		// Prepare date time (RFC 2822 - Day, DD Mon YYYY HH:mm:ss ±ZZZZ).
		var datetime RtcDateTime
		datetime.Day = uint16(parseInt(datetimeStr[5:7]))
		datetime.Month = GetMonthFromString(datetimeStr[8:11])
		datetime.Year = uint16(parseInt(datetimeStr[12:16]))
		datetime.Hour = uint16(parseInt(datetimeStr[17:19]))
		datetime.Minute = uint16(parseInt(datetimeStr[20:22]))
		datetime.Second = uint16(parseInt(datetimeStr[23:25]))

		// Populate result.
		libSceRtc_sceRtcGetTick(uintptr(unsafe.Pointer(&datetime)), tickUtcPtr)

		// Handle timezone if present.
		if len(datetimeStr) > 26 && (datetimeStr[26] == '+' || datetimeStr[26] == '-') {
			offset := parseInt(datetimeStr[27:29])*60 + parseInt(datetimeStr[29:31])
			if datetimeStr[26] == '-' {
				offset *= -1
			}
			libSceRtc_sceRtcTickAddMinutes(tickUtcPtr, tickUtcPtr, int64(offset))
		}
	} else {
		// Prepare date time (asctime - Www Mmm dd hh:mm:ss yyyy).
		var datetime RtcDateTime
		datetime.Month = GetMonthFromString(datetimeStr[4:7])
		datetime.Day = uint16(parseInt(datetimeStr[8:10]))
		datetime.Hour = uint16(parseInt(datetimeStr[11:13]))
		datetime.Minute = uint16(parseInt(datetimeStr[14:16]))
		datetime.Second = uint16(parseInt(datetimeStr[17:19]))
		datetime.Year = uint16(parseInt(datetimeStr[20:24]))

		// Populate result.
		libSceRtc_sceRtcGetTick(uintptr(unsafe.Pointer(&datetime)), tickUtcPtr)
	}

	return 0
}

// 0x0000000000001230
// __int64 __fastcall sceRtcParseRFC3339(_QWORD *, unsigned __int8 *)
func libSceRtc_sceRtcParseRFC3339(tickUtcPtr uintptr, datetimeStrPtr Cstring) uintptr {
	if tickUtcPtr == 0 || datetimeStrPtr == nil {
		logger.Printf("%-132s %s failed due to invalid tick or date time string pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceRtcParseRFC3339"),
		)
		return 0x80B50002
	}

	// Prepare date time (RFC 3339 - YYYY-MM-DDTHH:MM:SSZ).
	var datetime RtcDateTime
	datetimeStr := GoString(datetimeStrPtr)
	datetime.Year = uint16(parseInt(datetimeStr[0:4]))
	datetime.Month = uint16(parseInt(datetimeStr[5:7]))
	datetime.Day = uint16(parseInt(datetimeStr[8:10]))
	datetime.Hour = uint16(parseInt(datetimeStr[11:13]))
	datetime.Minute = uint16(parseInt(datetimeStr[14:16]))
	datetime.Second = uint16(parseInt(datetimeStr[17:19]))

	// Parse floating point.
	tzPos := 19
	if datetimeStr[19] == '.' {
		datetime.Microsecond = uint32(parseInt(datetimeStr[20:22])) * 10000 // Convert hundredths to micro.
		tzPos = 22
	}

	// Populate result.
	libSceRtc_sceRtcGetTick(uintptr(unsafe.Pointer(&datetime)), tickUtcPtr)

	// Handle timezone if present.
	if datetimeStr[tzPos] != 'Z' {
		offset := parseInt(datetimeStr[tzPos+1:tzPos+3])*60 + parseInt(datetimeStr[tzPos+4:tzPos+6])
		if datetimeStr[tzPos] == '-' {
			offset *= -1
		}
		libSceRtc_sceRtcTickAddMinutes(tickUtcPtr, tickUtcPtr, int64(offset))
	}

	return 0
}

// 0x0000000000002130
// __int64 __fastcall sceRtcGetDaysInMonth(int, int)
func libSceRtc_sceRtcGetDaysInMonth(year, month int) uintptr {
	if year <= 0 {
		return 0x80B50008
	}
	if month <= 0 || month > 12 {
		return 0x80B50009
	}

	return uintptr(GetDaysInMonth(year, month))
}

// 0x0000000000002190
// __int64 __fastcall sceRtcGetDayOfWeek(int, int, int)
func libSceRtc_sceRtcGetDayOfWeek(year, month, day int) uintptr {
	// Basic validation
	if year < 1 || year > 9999 {
		return 0x80B50008
	}
	if month < 1 || month > 12 {
		return 0x80B50009
	}
	if day < 1 || day > int(libSceRtc_sceRtcGetDaysInMonth(year, month)) {
		return 0x80B5000A
	}

	// Jan/Feb are months 13/14 of previous year.
	if month <= 2 {
		month += 12
		year--
	}

	// Zeller's congruence.
	// h = (q + floor(13(m+1)/5) + K + floor(K/4) + floor(J/4) + 5J) mod 7.
	k := year % 100
	j := year / 100
	h := (day + (13 * (month + 1) / 5) + k + (k / 4) + (j / 4) + (5 * j)) % 7

	return uintptr(h)
}

// 0x0000000000002650
// __int64 __fastcall sceRtcCheckValid(int *)
func libSceRtc_sceRtcCheckValid(datetimePtr uintptr) uintptr {
	if datetimePtr == 0 {
		logger.Printf("%-132s %s failed due to invalid date time pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceRtcCheckValid"),
		)
		return 0x80B50002
	}
	datetime := (*RtcDateTime)(unsafe.Pointer(datetimePtr))
	if datetime.Year == 0 || datetime.Year > 9999 {
		return 0x80B50008
	}
	if datetime.Month == 0 || datetime.Month > 12 {
		return 0x80B50009
	}
	maxDays := uint16(libSceRtc_sceRtcGetDaysInMonth(int(datetime.Year), int(datetime.Month)))
	if datetime.Day == 0 || datetime.Day > maxDays {
		return 0x80B5000A
	}
	if datetime.Hour >= 24 {
		return 0x80B5000B
	}
	if datetime.Minute >= 60 {
		return 0x80B5000C
	}
	if datetime.Second >= 60 {
		return 0x80B5000D
	}
	if datetime.Microsecond >= 1000000 {
		return 0x80B5000E
	}

	return 0
}
