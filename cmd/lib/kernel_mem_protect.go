package lib

import (
	"fmt"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x0000000000014950
// __int64 __fastcall sceKernelMprotect()
func libKernel_sceKernelMprotect(addr, length, prot uintptr) uintptr {
	err := libKernel_sys_mprotect(addr, length, prot)
	if err == ERR_PTR {
		return GetErrno() - 0x7FFE0000
	}

	return 0
}

func libKernel_sys_mprotect(addr, length, prot uintptr) uintptr {
	if addr == 0 {
		fmt.Printf("%-120s %s failed due to invalid address.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelMprotect"),
		)
		SetErrno(EINVAL)
		return ERR_PTR
	}

	ret, err := libKernel_protect(addr, length, prot)
	if ret == 0 {
		fmt.Printf("%-120s %s failed changing protection: %s\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelMprotect"),
			err.Error(),
		)
		SetErrno(EPERM)
		return ERR_PTR
	}

	fmt.Printf("%-120s %s changed protection of %s bytes at %s to %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelMprotect"),
		color.Yellow.Sprintf("0x%X", length),
		color.Yellow.Sprintf("0x%X", addr),
		color.Yellow.Sprintf("0x%X", prot),
	)
	return 0
}
