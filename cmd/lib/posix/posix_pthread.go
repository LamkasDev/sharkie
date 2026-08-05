package posix

import (
	"context"
	"math"
	"runtime"
	"runtime/pprof"
	"sync/atomic"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
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

func Pthread_self() uintptr {
	return libScePosix_pthread_self()
}

func libScePosix_pthread_self() uintptr {
	thread := emu.GetCurrentThread()
	threadPtr := (uintptr)(unsafe.Pointer(thread.Tcb.Thread))
	/* logger.Printf("%-132s %s returned thread %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_self"),
		color.Yellow.Sprintf("0x%X", threadPtr),
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

func Pthread_detach(threadPtr uintptr) uintptr {
	return libScePosix_pthread_detach(threadPtr)
}

// TODO: finish this.
func libScePosix_pthread_detach(threadPtr uintptr) uintptr {
	if threadPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid thread pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_detach"),
		)
		return EINVAL
	}
	thread := emu.GetThreadForPtr(threadPtr)
	if thread == nil {
		logger.Printf("%-132s %s failed due to invalid thread %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_detach"),
			color.Yellow.Sprintf("0x%X", threadPtr),
		)
		return ENOENT
	}

	logger.Printf("%-132s %s detached %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_detach"),
		color.Green.Sprint(thread.Name),
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

func Pthread_cancel(threadPtr uintptr) uintptr {
	return libScePosix_pthread_cancel(threadPtr)
}

func libScePosix_pthread_cancel(threadPtr uintptr) uintptr {
	if threadPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid thread pointer %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_cancel"),
			color.Yellow.Sprintf("0x%X", threadPtr),
		)
		return EINVAL
	}
	thread := emu.GetThreadForPtr(threadPtr)
	if thread == nil {
		logger.Printf("%-132s %s failed due to invalid thread %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_cancel"),
			color.Yellow.Sprintf("0x%X", threadPtr),
		)
		return ENOENT
	}

	logger.Printf("%-132s %s tried cancelling thread %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_cancel"),
		color.Blue.Sprint(thread.Name),
	)
	return 0
}

func Pthread_once(onceControl *PthreadOnce, initRoutine uintptr) uintptr {
	return libScePosix_pthread_once(onceControl, initRoutine)
}

func libScePosix_pthread_once(onceControl *PthreadOnce, initRoutine uintptr) uintptr {
	if onceControl == nil || initRoutine == 0 {
		logger.Printf("%-132s %s failed due to invalid control or init routine.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_once"),
		)
		return EINVAL
	}

	thread := emu.GetCurrentThread()
	statePtr := &onceControl.State
	umtxAddress := uintptr(unsafe.Pointer(statePtr))

	for {
		state := PthreadOnceState(atomic.LoadUint32(statePtr))
		if state == PthreadOnceStateDone {
			return 0
		}

		if state == PthreadOnceStateNeverDone {
			if atomic.CompareAndSwapUint32(statePtr, uint32(state), uint32(PthreadOnceStateInProgress)) {
				break
			}
		} else if state == PthreadOnceStateInProgress {
			if atomic.CompareAndSwapUint32(statePtr, uint32(state), uint32(PthreadOnceStateWait)) {
				// Suspend thread using futex wait.
				libScePosix_sys_umtx_op(umtxAddress, UMTX_OP_WAIT_UINT_PRIVATE, uintptr(PthreadOnceStateWait), 0, 0)
			}
		} else if state == PthreadOnceStateWait {
			// Already in wait state, suspend thread using futex wait.
			libScePosix_sys_umtx_op(umtxAddress, UMTX_OP_WAIT_UINT_PRIVATE, uintptr(state), 0, 0)
		} else {
			return EINVAL
		}
	}

	// TODO: finish cleanup routines.

	// Execute the guest init function.
	thread.CallAndWaitFromStub(initRoutine, 0)

	// Try a clean transition from InProgress to Done.
	if atomic.CompareAndSwapUint32(statePtr, uint32(PthreadOnceStateInProgress), uint32(PthreadOnceStateDone)) {
		return 0
	}

	// If we couldn't cleanly swap, it means another thread set it to Wait; force state to Done.
	atomic.StoreUint32(statePtr, uint32(PthreadOnceStateDone))

	// Broadcast to all waiting threads.
	libScePosix_sys_umtx_op(umtxAddress, UMTX_OP_WAKE_PRIVATE, math.MaxInt32, 0, 0)

	return 0
}

func Pthread_yield() uintptr {
	return libScePosix_pthread_yield()
}

func libScePosix_pthread_yield() uintptr {
	runtime.Gosched()
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
