package kernel

import (
	"strings"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x000000000002C370
// void sceKernelLoadStartModuleForSysmodule()
func libKernel_sceKernelLoadStartModuleForSysmodule(namePtr Cstring, argc, argvPtr uintptr, flags uint32, optionPtr uintptr, resultPtr *int32) uintptr {
	return libKernel_sys_sceKernelLoadStartModule(namePtr, argc, argvPtr, flags, optionPtr, resultPtr)
}

// 0x000000000002BB00
// __int64 __fastcall sceKernelLoadStartModule(__int64, __int64, __int64, int, __int64, int *, __m128, __m128, __m128, __m128, double, double, __m128, __m128)
func libKernel_sceKernelLoadStartModule(namePtr Cstring, argc, argvPtr uintptr, flags uint32, optionPtr uintptr, resultPtr *int32) uintptr {
	// TODO: this does a check, but not sure about the signature
	return libKernel_sys_sceKernelLoadStartModule(namePtr, argc, argvPtr, flags, optionPtr, resultPtr)
}

func libKernel_sys_sceKernelLoadStartModule(namePtr Cstring, argc, argvPtr uintptr, flags uint32, optionPtr uintptr, resultPtr *int32) uintptr {
	if namePtr == nil {
		logger.Printf("%-132s %s failed due to invalid name pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelLoadStartModule"),
		)
		return SCE_KERNEL_ERROR_EINVAL
	}
	moduleNameOrPath := GoString(namePtr)
	if strings.HasPrefix(moduleNameOrPath, "/") {
		if hostPath, err := fs.GlobalFilesystem.GetHostPath(moduleNameOrPath); err == nil {
			moduleNameOrPath = hostPath
		}
	}
	if module := emu.GlobalModuleManager.GetModule(moduleNameOrPath); module != nil {
		logger.Printf("%-132s %s skipping already loaded module %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelLoadStartModule"),
			color.Blue.Sprint(moduleNameOrPath),
		)
		return uintptr(module.ModuleIndex)
	}
	logger.Printf("%-132s %s loading %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelLoadStartModule"),
		color.Blue.Sprint(moduleNameOrPath),
	)

	// Load the module and its dependencies into memory and link them.
	mod, err := emu.GlobalModuleManager.LoadModule(moduleNameOrPath, true)
	if err != nil {
		logger.Printf("%-132s %s failed due to load error (%s)\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelLoadStartModule"),
			err.Error(),
		)
		return 0
	}

	// Initialize the module and its dependencies synchronously using CallAndWait.
	param := uintptr(0)
	if mod.ProcessParamSection != nil {
		param = mod.BaseAddress + uintptr(mod.ProcessParamSection.PVaddr)
	}
	ret := emu.GlobalModuleManager.RunModuleInitializers(emu.GetCurrentThread(), mod, true, false, argc, argvPtr, param)

	// Write back return value of init function.
	if resultPtr != nil {
		*resultPtr = int32(ret)
	}

	return uintptr(mod.ModuleIndex)
}
