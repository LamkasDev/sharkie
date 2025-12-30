package lib

import (
	"encoding/binary"
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
	seconds := now.Unix()
	nanoSeconds := now.Nanosecond()

	timestampSlice := unsafe.Slice((*byte)(unsafe.Pointer(timestampPtr)), 16)
	binary.LittleEndian.PutUint64(timestampSlice, uint64(seconds))
	binary.LittleEndian.PutUint64(timestampSlice[8:], uint64(nanoSeconds))

	logger.Printf("%-132s %s returned %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelClockGettime"),
		color.Yellow.Sprintf("0x%X", seconds),
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
	seconds := uint64(now.Unix())
	uSeconds := uint64(now.Nanosecond() / 1000)

	timevalueSlice := unsafe.Slice((*byte)(unsafe.Pointer(timevaluePtr)), 16)
	binary.LittleEndian.PutUint64(timevalueSlice, seconds)
	binary.LittleEndian.PutUint64(timevalueSlice[8:], uSeconds)

	logger.Printf("%-132s %s returned %s (%sus).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelGettimeofday"),
		color.Yellow.Sprintf("0x%X", seconds),
		color.Yellow.Sprintf("0x%X", uSeconds),
	)
	return 0
}
