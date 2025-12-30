package lib

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x000000000001E060
// __int64 __fastcall sceKernelGetAppInfo(int, _DWORD *)
func libKernel_sceKernelGetAppInfo(processId uintptr, infoPtr uintptr) uintptr {
	if infoPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid info pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelGetAppInfo"),
		)
		return SCE_KERNEL_ERROR_EINVAL
	}

	info := (*AppInfo)(unsafe.Pointer(infoPtr))
	info.AppId = 0
	info.MmapFlags = 0
	info.AttributeExecutable = 0
	info.Attribute2 = 0
	namePtr := uintptr(unsafe.Pointer(&info.CusaName[0]))
	WriteCString(namePtr, "CUSA00001")
	info.DebugLevel = 0
	info.SlvFlags = 0
	info.MiniAppDmemFlags = 0
	info.RenderMode = 0
	info.MdbgOut = 0
	info.RequiredHdcpType = 0
	info.PreloadPrxFlags = 0
	info.Attribute1 = 0
	info.HasParamSfo = 1
	info.TitleWorkaround = TitleWorkaround{}

	logger.Printf("%-132s %s returned app info.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelGetAppInfo"),
	)
	return 0
}

// 0x000000000001E2E0
// __int64 __fastcall sceKernelTitleWorkaroundIsEnabled(__int64, unsigned __int64, _DWORD *)
func libKernel_sceKernelTitleWorkaroundIsEnabled() uintptr {
	titleWorkaround := uintptr(0)
	logger.Printf("%-132s %s returned %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelTitleWorkaroundIsEnabled"),
		color.Green.Sprintf("%d", titleWorkaround),
	)
	return titleWorkaround
}
