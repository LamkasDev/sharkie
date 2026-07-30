package kernel

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000016640
// __int64 __fastcall sceKernelCheckReachability(char *)
func libKernel_sceKernelCheckReachability(pathPtr Cstring) uintptr {
	if pathPtr == nil {
		logger.Printf("%-132s %s failed due to invalid path pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelCheckReachability"),
		)
		return SCE_KERNEL_ERROR_ENOENT
	}

	path := GetUsablePath(GoString(pathPtr))
	exists := GlobalFilesystem.Exists(path)
	if !exists {
		return SCE_KERNEL_ERROR_ENOENT
	}

	return 0
}
