package lib

import (
	"time"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x0000000000014CB0
// __int64 __fastcall sceKernelClockGettime(__int64, __int64)
func libKernel_sceKernelClockGettime(clockId, timestampPtr uintptr) uintptr {
	if timestampPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid time pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelClockGettime"),
		)
		return SCE_KERNEL_ERROR_EINVAL
	}

	now := time.Now()
	timestamp := (*Timestamp)(unsafe.Pointer(timestampPtr))
	timestamp.Seconds = uint64(now.Unix())
	timestamp.Nanoseconds = uint64(now.Nanosecond())

	logger.Printf("%-132s %s returned %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelClockGettime"),
		color.Yellow.Sprintf("0x%X", timestamp.Seconds),
	)
	return 0
}

// 0x0000000000014D50
// __int64 sceKernelGetProcessTime()
func libKernel_sceKernelGetProcessTime() uintptr {
	elapsed := time.Since(TscStartTime)
	micros := uintptr(elapsed.Microseconds())

	logger.Printf("%-132s %s returned %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelGetProcessTime"),
		color.Yellow.Sprintf("0x%X", micros),
	)
	return 0
}

// 0x0000000000014CE0
// __int64 __fastcall sceKernelGettimeofday(__int64)
func libKernel_sceKernelGettimeofday(timevaluePtr uintptr) uintptr {
	if timevaluePtr == 0 {
		logger.Printf("%-132s %s failed due to invalid time pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelGettimeofday"),
		)
		return SCE_KERNEL_ERROR_EINVAL
	}

	now := time.Now()
	timevalue := (*Timevalue)(unsafe.Pointer(timevaluePtr))
	timevalue.Seconds = uint64(now.Unix())
	timevalue.Microseconds = uint64(now.Nanosecond() / 1000)

	logger.Printf("%-132s %s returned %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelGettimeofday"),
		color.Yellow.Sprintf("0x%X", timevalue.Seconds),
	)
	return 0
}
