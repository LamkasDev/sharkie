package kernel

import (
	"encoding/binary"
	"runtime"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

const MainThreadGlobalOffset = 0x8E430
const PidGlobalOffset = 0x8E580
const PageSizeGlobalOffset = 0x8E450
const PageSizeGlobalOffset64 = 0x8E448
const InitFlagOffset = 0x8DF78
const SmpFlagOffset = 0x8DEB0

var MainThreadInitialized = false

// 0x000000000000B530
// unsigned __int64 pthread_self()
func libKernel_pthread_self() uintptr {
	if !MainThreadInitialized {
		libKernel_sys_pthread_self()
	}

	thread := emu.GetCurrentThread()
	threadPtr := (uintptr)(unsafe.Pointer(thread.Tcb.Thread))
	/* logger.Printf("%-132s %s returned thread %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_self"),
		color.Yellow.Sprintf("0x%X", thread),
	) */
	return threadPtr
}

func libKernel_sys_pthread_self() {
	emu.GlobalModuleManager.ModulesLock.RLock()
	defer emu.GlobalModuleManager.ModulesLock.RUnlock()

	mainThread := emu.GlobalModuleManager.MainThread
	base := emu.GlobalModuleManager.ModulesMap["libkernel.sprx"].BaseAddress

	mainThreadPtr := (uintptr)(unsafe.Pointer(mainThread.Tcb.Thread))
	WriteAddress(base+MainThreadGlobalOffset, mainThreadPtr)

	pidSlice := unsafe.Slice((*byte)(unsafe.Pointer(base+PidGlobalOffset)), 4)
	binary.LittleEndian.PutUint32(pidSlice, uint32(libKernel_getpid()))

	pageSizeSlice := unsafe.Slice((*byte)(unsafe.Pointer(base+PageSizeGlobalOffset)), 4)
	binary.LittleEndian.PutUint32(pageSizeSlice, uint32(MemoryPageSize))

	pageSize64Slice := unsafe.Slice((*byte)(unsafe.Pointer(base+PageSizeGlobalOffset64)), 8)
	binary.LittleEndian.PutUint64(pageSize64Slice, uint64(MemoryPageSize))

	initFlagSlice := unsafe.Slice((*byte)(unsafe.Pointer(base+InitFlagOffset)), 1)
	initFlagSlice[0] = 1
	smpFlagSlice := unsafe.Slice((*byte)(unsafe.Pointer(base+SmpFlagOffset)), 4)
	binary.LittleEndian.PutUint32(smpFlagSlice, 1)

	MainThreadInitialized = true
	logger.Printf("%-132s %s initialized main thread.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_self"),
	)
}

// 0x0000000000007590
// _BOOL8 __fastcall pthread_equal(__int64, __int64)
func libKernel_pthread_equal(t1, t2 uintptr) uintptr {
	if t1 == t2 {
		return 1
	}
	return 0
}

// 0x0000000000003720
// __int64 __fastcall pthread_getaffinity_np(signed __int32 *, __int64, __int64)
func libKernel_pthread_getaffinity_np(threadPtr uintptr, cpuSetSize uintptr, cpuSetPtr uintptr) uintptr {
	if threadPtr == 0 || cpuSetPtr == 0 || cpuSetSize < 8 {
		logger.Printf("%-132s %s failed due to invalid thread or cpu set pointer %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_getaffinity_np"),
			color.Yellow.Sprintf("0x%X", cpuSetPtr),
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
	cpuSet := (*ThreadCpuSet)(unsafe.Pointer(cpuSetPtr))
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
func libKernel_pthread_setaffinity_np(threadPtr, cpuSetSize, cpuSetPtr uintptr) uintptr {
	if threadPtr == 0 || cpuSetPtr == 0 || cpuSetSize < 8 {
		logger.Printf("%-132s %s failed due to invalid thread or cpu set pointer %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_getaffinity_np"),
			color.Yellow.Sprintf("0x%X", cpuSetPtr),
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
	cpuSet := (*ThreadCpuSet)(unsafe.Pointer(cpuSetPtr))
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

// 0x0000000000007770
// void __fastcall __noreturn pthread_exit(__int64)
func libKernel_pthread_exit(retValue uintptr) uintptr {
	return libKernel_sys_pthread_exit(retValue)
}

func libKernel_sys_pthread_exit(retValue uintptr) uintptr {
	// Mark thread as done and exit goroutine.
	thread := emu.GetCurrentThread()
	thread.Exit(retValue)
	runtime.Goexit()

	return 0
}
