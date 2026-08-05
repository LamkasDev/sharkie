package kernel

import (
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/module"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x000000000002CFF0
// __int64 __fastcall sceKernelGetExecutableModuleHandle()
func libKernel_sceKernelGetExecutableModuleHandle() uintptr {
	handle := ModuleHandle(emu.GlobalModuleManager.CurrentModule.ModuleIndex)

	logger.Printf("%-132s %s returned module handle %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelGetExecutableModuleHandle"),
		color.Yellow.Sprintf("0x%X", handle),
	)
	return uintptr(handle)
}

// 0x0000000000016BE0
// __int64 sceKernelIsInSandbox()
func libKernel_sceKernelIsInSandbox() uintptr {
	logger.Printf("%-132s %s returned false.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelIsInSandbox"),
	)
	return 0
}

// 0x000000000001A920
// __int64 sceKernelGetCompiledSdkVersion()
func libKernel_sceKernelGetCompiledSdkVersion(versionPtr uintptr) uintptr {
	sdkVersion := GameCompiledSdkVersion
	if versionPtr != 0 {
		versionSlice := unsafe.Slice((*byte)(unsafe.Pointer(versionPtr)), 4)
		binary.LittleEndian.PutUint32(versionSlice, sdkVersion)
	}

	if logger.LogMisc {
		logger.Printf("%-132s %s returned %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelGetCompiledSdkVersion"),
			color.Yellow.Sprintf("0x%X", sdkVersion),
		)
	}
	return 0
}

// 0x0000000000022280
// __int64 __fastcall sceKernelSetCallRecord(int)
func libKernel_sceKernelSetCallRecord() uintptr {
	return 0
}
