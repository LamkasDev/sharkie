package kernel

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000003720
// __int64 __fastcall pthread_getaffinity_np(signed __int32 *, __int64, __int64)
func libKernel_pthread_getaffinity_np(threadPtr, cpuSetSize uintptr, cpuSet *ThreadCpuSet) uintptr {
	if threadPtr == 0 || cpuSet == nil || cpuSetSize < 8 {
		logger.Printf("%-132s %s failed due to invalid thread or cpu set pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_getaffinity_np"),
		)
		return EINVAL
	}
	thread := emu.GetThreadForPtr(threadPtr)
	if thread == nil {
		logger.Printf("%-132s %s failed due to invalid thread %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_getaffinity_np"),
			color.Yellow.Sprintf("0x%X", threadPtr),
		)
		return ENOENT
	}

	// Get thread's affinity.
	cpuSet.Low = uint64(thread.AffinityMask)
	cpuSet.High = 0

	logger.Printf("%-132s %s returned affinity %s of %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_getaffinity_np"),
		color.Yellow.Sprintf("0x%X", cpuSet.Low),
		color.Green.Sprint(thread.Name),
	)
	return 0
}

// 0x0000000000003640
// __int64 __fastcall pthread_setaffinity_np(signed __int32 *, __int64, __int64)
func libKernel_pthread_setaffinity_np(threadPtr, cpuSetSize uintptr, cpuSet *ThreadCpuSet) uintptr {
	if threadPtr == 0 || cpuSet == nil || cpuSetSize < 8 {
		logger.Printf("%-132s %s failed due to invalid thread or cpu set pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_getaffinity_np"),
		)
		return EINVAL
	}
	thread := emu.GetThreadForPtr(threadPtr)
	if thread == nil {
		logger.Printf("%-132s %s failed due to invalid thread %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_getaffinity_np"),
			color.Yellow.Sprintf("0x%X", threadPtr),
		)
		return ENOENT
	}

	// Set thread's affinity.
	thread.Lock.Lock()
	thread.AffinityMask = ThreadAffinityMask(cpuSet.Low)
	thread.Lock.Unlock()

	logger.Printf("%-132s %s set affinity of %s to %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_setaffinity_np"),
		color.Green.Sprint(thread.Name),
		color.Yellow.Sprintf("0x%X", cpuSet.Low),
	)
	return 0
}
