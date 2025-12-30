package lib

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x00000000000289C0
// __int64 __fastcall _tls_get_addr(_QWORD *, __int64, __int64, __int64, __int64, int)
func libKernel___tls_get_addr(tlsIndexPtr uintptr) uintptr {
	if tlsIndexPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid tls index pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("__tls_get_addr"),
		)
		return EFAULT
	}

	tlsIndex := (*TlsIndex)(unsafe.Pointer(tlsIndexPtr))
	address, ok := TlsBaseRepo[tlsIndex.ModuleId]
	if !ok {
		logger.Printf("%-132s %s failed due to invalid module index %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("__tls_get_addr"),
			color.Green.Sprint(tlsIndex.ModuleId),
		)
		return 0
	}

	logger.Printf("%-132s %s returning tls address %s for module %s (offset=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("__tls_get_addr"),
		color.Yellow.Sprintf("0x%X", address),
		color.Green.Sprintf("%d", tlsIndex.ModuleId),
		color.Yellow.Sprintf("0x%X", tlsIndex.Offset),
	)
	return address + tlsIndex.Offset
}
