package posix

import (
	"context"
	"runtime"
	"runtime/pprof"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/pthread"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func Pthread_create(threadPtr uintptr, attrHandlePtr *uintptr, entryPoint, arg uintptr) uintptr {
	return libScePosix_pthread_create(threadPtr, attrHandlePtr, entryPoint, arg)
}

func libScePosix_pthread_create(threadPtr uintptr, attrHandlePtr *uintptr, entryPoint, arg uintptr) uintptr {
	return libScePosix_pthread_create_name_np(threadPtr, attrHandlePtr, entryPoint, arg, nil)
}

func Pthread_create_name_np(threadPtr uintptr, attrHandlePtr *uintptr, entryPoint, arg uintptr, namePtr Cstring) uintptr {
	return libScePosix_pthread_create_name_np(threadPtr, attrHandlePtr, entryPoint, arg, namePtr)
}

func libScePosix_pthread_create_name_np(threadPtr uintptr, attrHandlePtr *uintptr, entryPoint, arg uintptr, namePtr Cstring) uintptr {
	// Check if entry point is valid.
	module := emu.GetModuleAtAddress(entryPoint)
	if module == nil {
		logger.Printf("%-132s %s failed due to invalid entry point %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_create_name_np"),
			color.Yellow.Sprintf("0x%X", entryPoint),
		)
		return EINVAL
	}

	// Figure out stack size beforehand.
	stackSize := uint64(StackDefaultSize)
	attr, _ := ResolveHandle[PthreadAttr](attrHandlePtr)
	if attr != nil {
		stackSize = attr.StackSize
	}

	// Create thread and assign attributes.
	thread := emu.CreateThread(GoString(namePtr), stackSize)
	thread.Tcb.Thread.StartFunc = entryPoint
	thread.Tcb.Thread.Arg = arg
	if attr != nil {
		thread.Tcb.Thread.Attr = *attr
	}
	thread.Tcb.Thread.Magic = PthreadMagic

	// Write back result.
	pthreadAddr := (uintptr)(unsafe.Pointer(thread.Tcb.Thread))
	if threadPtr != 0 {
		WriteAddress(threadPtr, pthreadAddr)
	}

	go pprof.Do(context.Background(), pprof.Labels("name", thread.Name), func(ctx context.Context) {
		thread.CallAndWait(entryPoint, arg)
		thread.Exit(0xDEAD)
	})

	logger.Printf("%-132s %s created thread %s at %s (%s at %s, arg=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_create_name_np"),
		color.Blue.Sprint(thread.Name),
		color.Yellow.Sprintf("0x%X", pthreadAddr),
		color.Blue.Sprint(module.Name),
		color.Yellow.Sprintf("0x%X", entryPoint-module.BaseAddress),
		color.Yellow.Sprintf("0x%X", arg),
	)
	return 0
}

func Pthread_join(threadPtr, retValPtr uintptr) uintptr {
	return libScePosix_pthread_join(threadPtr, retValPtr)
}

func libScePosix_pthread_join(threadPtr, retValPtr uintptr) uintptr {
	if threadPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid thread pointer %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_join"),
			color.Yellow.Sprintf("0x%X", threadPtr),
		)
		return EINVAL
	}
	thread := emu.GetThreadForPtr(threadPtr)
	if thread == nil {
		logger.Printf("%-132s %s failed due to invalid thread %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_join"),
			color.Yellow.Sprintf("0x%X", threadPtr),
		)
		return ENOENT
	}

	// No being naughty.
	if thread == emu.GetCurrentThread() {
		logger.Printf("%-132s %s failed trying to join itself.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_join"),
		)
		return EDEADLK
	}

	// Wait for thread to exit.
	thread.Lock.Lock()
	for !thread.Exited {
		logger.Printf("%-132s %s waiting for thread %s to exit...\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_join"),
			color.Blue.Sprint(thread.Name),
		)
		thread.JoinCond.Wait()
	}
	exitCode := thread.ExitCode
	thread.Lock.Unlock()

	// Write back exit code.
	if retValPtr != 0 {
		WriteAddress(retValPtr, exitCode)
	}

	logger.Printf("%-132s %s joined thread %s (exitCode=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_join"),
		color.Blue.Sprint(thread.Name),
		color.Yellow.Sprintf("0x%X", exitCode),
	)
	return 0
}

func Pthread_self() uintptr {
	return libScePosix_pthread_self()
}

func libScePosix_pthread_self() uintptr {
	thread := emu.GetCurrentThread()
	threadPtr := (uintptr)(unsafe.Pointer(thread.Tcb.Thread))
	/* logger.Printf("%-132s %s returned thread %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_self"),
		color.Yellow.Sprintf("0x%X", thread),
	) */
	return threadPtr
}

func Pthread_equal(t1, t2 uintptr) uintptr {
	return libScePosix_pthread_equal(t1, t2)
}

func libScePosix_pthread_equal(t1, t2 uintptr) uintptr {
	if t1 == t2 {
		return 1
	}
	return 0
}

func Pthread_exit(retValue uintptr) uintptr {
	return libScePosix_pthread_exit(retValue)
}

func libScePosix_pthread_exit(retValue uintptr) uintptr {
	// Mark thread as done and exit goroutine.
	thread := emu.GetCurrentThread()
	thread.Exit(retValue)
	runtime.Goexit()

	return 0
}

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
