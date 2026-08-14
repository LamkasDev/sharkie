package rtc

import "github.com/LamkasDev/sharkie/cmd/elf"

func RegisterRtcStubs() {
	// Time functions.
	elf.RegisterStub("libSceRtc", "sceRtcGetCurrentClockLocalTime", libSceRtc_sceRtcGetCurrentClockLocalTime)
	elf.RegisterStub("libSceRtc", "sceRtcParseDateTime", libSceRtc_sceRtcParseDateTime)
	elf.RegisterStub("libSceRtc", "sceRtcParseRFC3339", libSceRtc_sceRtcParseRFC3339)
	elf.RegisterStub("libSceRtc", "sceRtcGetDaysInMonth", libSceRtc_sceRtcGetDaysInMonth)
	elf.RegisterStub("libSceRtc", "sceRtcGetDayOfWeek", libSceRtc_sceRtcGetDayOfWeek)
	elf.RegisterStub("libSceRtc", "sceRtcCheckValid", libSceRtc_sceRtcCheckValid)

	// Tick functions.
	elf.RegisterStub("libSceRtc", "sceRtcGetCurrentTick", libSceRtc_sceRtcGetCurrentTick)
	elf.RegisterStub("libSceRtc", "sceRtcGetCurrentRawNetworkTick", libSceRtc_sceRtcGetCurrentRawNetworkTick)
	elf.RegisterStub("libSceRtc", "sceRtcSetTick_t", libSceRtc_sceRtcSetTick_t)
	elf.RegisterStub("libSceRtc", "sceRtcSetTick", libSceRtc_sceRtcSetTick)
	elf.RegisterStub("libSceRtc", "sceRtcGetTick", libSceRtc_sceRtcGetTick)
	elf.RegisterStub("libSceRtc", "sceRtcTickAddHours", libSceRtc_sceRtcTickAddHours)
	elf.RegisterStub("libSceRtc", "sceRtcTickAddMinutes", libSceRtc_sceRtcTickAddMinutes)
	elf.RegisterStub("libSceRtc", "sceRtcTickAddSeconds", libSceRtc_sceRtcTickAddSeconds)
	elf.RegisterStub("libSceRtc", "sceRtcGetTickResolution", libSceRtc_sceRtcGetTickResolution)
}
