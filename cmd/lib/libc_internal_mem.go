package lib

import (
	"fmt"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x00000000000311F0
// __int64 __fastcall sceLibcMspaceCalloc(__int64, unsigned __int64, unsigned __int64, __int64)
func libSceLibcInternal_sceLibcMspaceCalloc(mspace, nmemb, size uintptr) uintptr {
	size *= nmemb
	addr := libKernel_mmap(0, size, PROT_READ|PROT_WRITE, MAP_ANON|MAP_PRIVATE, ERR_PTR, 0)
	if addr == ERR_PTR {
		return 0
	}
	GlobalAllocator.Allocations[addr] = size

	return addr
}

// 0x0000000000033CF0
// __int64 __fastcall sceLibcMspaceFree(__int64, __int64 *, __int64, __int64, __m128)
func libSceLibcInternal_sceLibcMspaceFree(ptr uintptr) {
	if ptr == 0 {
		fmt.Printf("%-120s %s failed due to invalid pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceLibcMspaceFree"),
		)
		return
	}

	_, ok := GlobalAllocator.Allocations[ptr]
	if !ok {
		fmt.Printf("%-120s %s failed freeing untracked pointer %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceLibcMspaceFree"),
			color.Yellow.Sprintf("0x%X", ptr),
		)
		return
	}

	delete(GlobalAllocator.Allocations, ptr)
	fmt.Printf("%-120s %s freed memory at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceLibcMspaceFree"),
		color.Yellow.Sprintf("0x%X", ptr),
	)
}
