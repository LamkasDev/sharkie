package posix

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func Pthread_getaffinity_np(threadPtr, cpuSetSize uintptr, cpuSet *ThreadCpuSet) uintptr {
	return libScePosix_pthread_getaffinity_np(threadPtr, cpuSetSize, cpuSet)
}

func libScePosix_pthread_getaffinity_np(threadPtr, cpuSetSize uintptr, cpuSet *ThreadCpuSet) uintptr {
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

func Pthread_setaffinity_np(threadPtr, cpuSetSize uintptr, cpuSet *ThreadCpuSet) uintptr {
	return libScePosix_pthread_setaffinity_np(threadPtr, cpuSetSize, cpuSet)
}

func libScePosix_pthread_setaffinity_np(threadPtr, cpuSetSize uintptr, cpuSet *ThreadCpuSet) uintptr {
	if threadPtr == 0 || cpuSet == nil || cpuSetSize < 8 {
		logger.Printf("%-132s %s failed due to invalid thread or cpu set pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_setaffinity_np"),
		)
		return EINVAL
	}
	thread := emu.GetThreadForPtr(threadPtr)
	if thread == nil {
		logger.Printf("%-132s %s failed due to invalid thread %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_setaffinity_np"),
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

func Pthread_getschedparam() uintptr {
	return libScePosix_pthread_getschedparam()
}

// TODO: finish this.
func libScePosix_pthread_getschedparam() uintptr {
	return 0
}

func Pthread_setschedparam() uintptr {
	return libScePosix_pthread_setschedparam()
}

// TODO: finish this.
func libScePosix_pthread_setschedparam() uintptr {
	thread := emu.GetCurrentThread()
	_ = thread

	logger.Printf("%-132s %s tried setting sched param.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_setschedparam"),
	)
	return 0
}

func Pthread_setcancelstate() uintptr {
	return libScePosix_pthread_setcancelstate()
}

// TODO: finish this.
func libScePosix_pthread_setcancelstate() uintptr {
	thread := emu.GetCurrentThread()
	_ = thread

	logger.Printf("%-132s %s tried setting cancel state.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_setcancelstate"),
	)
	return 0
}
