package lib

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs/tcb"
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

	// Find the DTV entry for module index.
	currentThread := emu.GetCurrentThread()
	tlsIndex := (*TlsIndex)(unsafe.Pointer(tlsIndexPtr))
	dtvEntryPtr := uintptr(unsafe.Pointer(currentThread.Tcb.Dtv)) + (uintptr(tlsIndex.ModuleId+1) * DtvEntrySize)
	dtvEntry := (*DtvEntry)(unsafe.Pointer(dtvEntryPtr))

	// Check address.
	address := dtvEntry.Pointer
	if address == 0 {
		logger.Printf("%-132s %s failed due to invalid address inside DTV (moduleId=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("__tls_get_addr"),
			color.Green.Sprintf("%d", tlsIndex.ModuleId),
		)
		return 0
	}

	if logger.LogMisc {
		logger.Printf("%-132s %s returned tls address %s for %s (moduleId=%s, offset=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("__tls_get_addr"),
			color.Yellow.Sprintf("0x%X", address),
			color.Blue.Sprint(emu.GlobalModuleManager.Modules[tlsIndex.ModuleId].Name),
			color.Green.Sprintf("%d", tlsIndex.ModuleId),
			color.Yellow.Sprintf("0x%X", tlsIndex.Offset),
		)
	}
	return address + tlsIndex.Offset
}
