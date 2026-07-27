package kernel

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/module"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x000000000002C7B0
// __int64 __fastcall sceKernelGetModuleList(__int64, __int64, __int64)
func libKernel_sceKernelGetModuleList(handles *ModuleHandle, numArray uint64, outCount *uint64) uintptr {
	if handles == nil || outCount == nil {
		logger.Printf("%-132s %s failed due to invalid handles or out count pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelGetModuleList"),
		)
		return SCE_KERNEL_ERROR_EFAULT
	}
	emu.GlobalModuleManager.ModulesLock.RLock()
	defer emu.GlobalModuleManager.ModulesLock.RUnlock()

	var count uint64 = 0
	var handlesSlice []ModuleHandle
	if numArray > 0 {
		handlesSlice = unsafe.Slice(handles, numArray)
	}
	for _, mod := range emu.GlobalModuleManager.Modules {
		if mod == nil {
			continue
		}
		if count < numArray {
			handlesSlice[count] = ModuleHandle(mod.ModuleIndex)
			count++
		} else {
			logger.Printf("%-132s %s failed due to exceeding array size.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sceKernelGetModuleList"),
			)
			return SCE_KERNEL_ERROR_ENOMEM
		}
	}
	*outCount = count

	logger.Printf("%-132s %s returned %s modules.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelGetModuleList"),
		color.Green.Sprint(count),
	)
	return 0
}
