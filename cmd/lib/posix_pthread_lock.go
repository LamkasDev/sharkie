package lib

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// TODO: finish this
// 0x000000000000A660
// __int64 __fastcall pthread_rwlock_rdlock(__int64 *a1)
func libKernel_pthread_rwlock_rdlock() uintptr {
	thread := emu.GetCurrentThread()
	thread.Lock.Lock()

	if logger.LogSyncing {
		logger.Printf("%-132s %s locked thread %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_rwlock_rdlock"),
			color.Blue.Sprint(thread.Name),
		)
	}
	return 0
}

// TODO: finish this
// 0x000000000000AFA0
// __int64 __fastcall pthread_rwlock_wrlock(__int64)
func libKernel_pthread_rwlock_wrlock() uintptr {
	thread := emu.GetCurrentThread()
	thread.Lock.Lock()

	if logger.LogSyncing {
		logger.Printf("%-132s %s locked thread %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_rwlock_wrlock"),
			color.Blue.Sprint(thread.Name),
		)
	}
	return 0
}

// TODO: finish this
// 0x000000000000B210
// __int64 __fastcall pthread_rwlock_unlock(unsigned __int64 *)
func libKernel_pthread_rwlock_unlock() uintptr {
	thread := emu.GetCurrentThread()
	thread.Lock.Unlock()

	if logger.LogSyncing {
		logger.Printf("%-132s %s unlocked thread %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_rwlock_unlock"),
			color.Blue.Sprint(thread.Name),
		)
	}
	return 0
}
