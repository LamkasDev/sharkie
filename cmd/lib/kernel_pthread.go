package lib

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x00000000000146E0
// __int64 scePthreadGetthreadid()
func libKernel_scePthreadGetthreadid() uintptr {
	threadId := uintptr(emu.GlobalModuleManager.Tcb.Thread.ThreadId)

	logger.Printf("%-120s %s returned thread id %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("scePthreadGetthreadid"),
		color.Yellow.Sprintf("%d", threadId),
	)
	return threadId
}

// 0x00000000000146E0
// __int64 scePthreadSelf()
func libKernel_scePthreadSelf() uintptr {
	return libKernel_pthread_self()
}

// 0x0000000000013920
// __int64 scePthreadEqual()
func libKernel_scePthreadEqual(t1, t2 uintptr) uintptr {
	err := libKernel_pthread_equal(t1, t2)
	if err != 0 {
		return err - 0x7FFE0000
	}

	return 0
}
