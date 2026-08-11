package kernel

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/app_content"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x000000000001E060
// __int64 __fastcall sceKernelGetAppInfo(int, _DWORD *)
func libKernel_sceKernelGetAppInfo(processId int32, info *AppInfo) uintptr {
	if info == nil {
		logger.Printf("%-132s %s failed due to invalid info pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelGetAppInfo"),
		)
		return SCE_KERNEL_ERROR_EINVAL
	}
	*info = AppInfo{}
	info.HasParamSfo = 1
	CString(Cstring(&info.CusaName), GlobalAppContentInstance.ParamSfo.MapStrings["TITLE_ID"])

	logger.Printf("%-132s %s returned app info.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelGetAppInfo"),
	)
	return 0
}

// 0x000000000001A850
// __int64 __fastcall sceKernelGetSystemSwVersion(__int64 a1, __m128 _XMM0)
func libKernel_sceKernelGetSystemSwVersion(swVersion *SwVersion) uintptr {
	if swVersion == nil {
		return 0
	}
	swVersion.Hex = CurrentFirmwareVersion

	logger.Printf("%-132s %s returned sw version %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelGetAppInfo"),
		color.Yellow.Sprintf("0x%X", swVersion.Hex),
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
