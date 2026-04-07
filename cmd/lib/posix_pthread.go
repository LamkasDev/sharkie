package lib

import (
	"context"
	"encoding/binary"
	"runtime"
	"runtime/pprof"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs/pthread"
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

// 0x0000000000006DA0
// __int64 __fastcall pthread_create_name_np(int **, __int64 *, __int64, __int64, _BYTE *, __m128 _XMM0)
func libKernel_pthread_create_name_np(threadPtr, attrHandlePtr, entryPoint, arg uintptr, namePtr Cstring) uintptr {
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
		thread.Call(entryPoint, arg)
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

// 0x0000000000008880
// __int64 __fastcall pthread_join(__int64, __int64)
func libKernel_pthread_join(threadPtr, retValPtr uintptr) uintptr {
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
