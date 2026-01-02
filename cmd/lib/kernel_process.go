package lib

import (
	"time"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x00000000000006F0
// __int64 __fastcall getpid()
func libKernel_getpid() uintptr {
	processId := uintptr(1001)
	logger.Printf("%-132s %s returned process id %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("getpid"),
		color.Green.Sprintf("%d", processId),
	)

	return processId
}

// 0x00000000000233E0
// __int64 __fastcall sceKernelGetProcessType()
func libKernel_sceKernelGetProcessType() uintptr {
	processType := uintptr(1)
	logger.Printf("%-132s %s returned process type %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelGetProcessType"),
		color.Blue.Sprintf("0x%X", processType),
	)

	return processType
}

// 0x000000000001A790
// __int64 sceKernelGetProcParam()
func libKernel_sceKernelGetProcParam() uintptr {
	module := emu.GlobalModuleManager.CurrentModule
	if module.ProcessParamSection != nil {
		addr := module.BaseAddress + uintptr(module.ProcessParamSection.PVaddr)
		logger.Printf("%-132s %s returning process parameters %s (relative=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelGetProcParam"),
			color.Yellow.Sprintf("0x%X", addr),
			color.Yellow.Sprintf("0x%X", module.ProcessParamSection.POffset),
		)
		return addr
	}

	logger.Printf("%-132s %s failed to return process parameters.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelGetProcParam"),
	)
	return 0
}

// 0x0000000000014BE0
// __int64 __fastcall sceKernelUsleep(unsigned int)
func libKernel_sceKernelUsleep(micros uintptr) uintptr {
	logger.Printf("%-132s %s sleeping for %s microseconds.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelUsleep"),
		color.Yellow.Sprintf("0x%X", micros),
	)
	time.Sleep(time.Duration(micros) * time.Microsecond)
	return 0
}

// 0x0000000000014B50
// __int64 __fastcall sceKernelNanosleep(__int128 *, __int64)
func libKernel_sceKernelNanosleep(timestampPtr uintptr) uintptr {
	if timestampPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid time pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelNanosleep"),
		)
		return SCE_KERNEL_ERROR_EINVAL
	}

	timestamp := (*Timestamp)(unsafe.Pointer(timestampPtr))
	timeout := time.Duration(timestamp.Seconds)*time.Second + time.Duration(timestamp.Nanoseconds)*time.Nanosecond

	logger.Printf("%-132s %s sleeping for %ss and %sns.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelNanosleep"),
		color.Yellow.Sprintf("0x%X", timestamp.Seconds),
		color.Yellow.Sprintf("0x%X", timestamp.Nanoseconds),
	)
	time.Sleep(timeout)
	return 0
}
