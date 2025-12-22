package lib

import (
	"fmt"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/gookit/color"
)

// 0x00000000000146E0
// __int64 scePthreadGetthreadid()
func libKernel_scePthreadGetthreadid() uintptr {
	threadId := uintptr(emu.GlobalModuleManager.Tcb.Thread.ThreadId)

	fmt.Printf("%-120s %s returned thread id %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("scePthreadGetthreadid"),
		color.Yellow.Sprintf("%d", threadId),
	)
	return threadId
}
