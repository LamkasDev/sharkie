package lib

import (
	"encoding/binary"
	"runtime"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/asm"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
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

	mainThreadSlice := unsafe.Slice((*byte)(unsafe.Pointer(base+MainThreadGlobalOffset)), 8)
	mainThreadPtr := (uintptr)(unsafe.Pointer(mainThread.Tcb.Thread))
	binary.LittleEndian.PutUint64(mainThreadSlice, uint64(mainThreadPtr))

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
func libKernel_pthread_create_name_np(threadPtr, attrHandlePtr, entryPoint, arg, namePtr uintptr) uintptr {
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
	stackSize := StackDefaultSize
	attr, _ := ResolveHandle[PthreadAttr](attrHandlePtr)
	if attr != nil {
		stackSize = attr.StackSize
	}

	// Create thread and assign attributes.
	thread := emu.CreateThread(namePtr, stackSize)
	thread.Tcb.Thread.StartFunc = entryPoint
	thread.Tcb.Thread.Arg = arg
	if attr != nil {
		thread.Tcb.Thread.Attr = *attr
	}
	thread.Tcb.Thread.Magic = PthreadMagic

	// Write back result.
	pthreadAddr := (uintptr)(unsafe.Pointer(thread.Tcb.Thread))
	if threadPtr != 0 {
		threadSlice := unsafe.Slice((*byte)(unsafe.Pointer(threadPtr)), 8)
		binary.LittleEndian.PutUint64(threadSlice, uint64(pthreadAddr))
	}

	threadName := thread.Name
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		thread.Setup()
		asm.GuestEnter()
		asm.Call(entryPoint, thread.Stack.CurrentPointer, arg, 0)
		asm.GuestLeave()
		logger.Printf("Thread %s exited.\n",
			color.Blue.Sprint(threadName),
		)
	}()

	logger.Printf("%-132s %s created thread %s at %s (%s at %s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_create_name_np"),
		color.Blue.Sprint(thread.Name),
		color.Yellow.Sprintf("0x%X", pthreadAddr),
		color.Blue.Sprint(module.Name),
		color.Yellow.Sprintf("0x%X", entryPoint-module.BaseAddress),
	)
	return 0
}

// 0x0000000000007770
// void __fastcall __noreturn pthread_exit(__int64)
func libKernel_pthread_exit(retValue uintptr) uintptr {
	return libKernel_sys_pthread_exit(retValue)
}

func libKernel_sys_pthread_exit(retValue uintptr) uintptr {
	thread := emu.GetCurrentThread()
	logger.Printf("%-132s %s exiting thread %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_exit"),
		color.Blue.Sprint(thread.Name),
	)
	thread.Tcb.Thread.ReturnValue = retValue

	// Process cleanup handlers.
	cleanupHandlerPtr := thread.Tcb.Thread.CleanupStack
	if cleanupHandlerPtr != 0 {
		limit := 0
		for limit < 20 {
			entry := (*PthreadCleanupEntry)(unsafe.Pointer(cleanupHandlerPtr))
			module := emu.GetModuleAtAddress(entry.Handler)
			if module == nil {
				logger.Printf("%-132s %s failed finding cleanup handler at %s...\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("pthread_exit"),
					color.Yellow.Sprintf("0x%X", entry.Handler),
				)
				continue
			}
			logger.Printf("%-132s %s skipped cleanup handler %s/%s with %s argument...\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("pthread_exit"),
				color.Blue.Sprint(module.Name),
				color.Yellow.Sprintf("0x%X", entry.Handler-module.BaseAddress),
				color.Yellow.Sprintf("0x%X", entry.Arg),
			)
			cleanupHandlerPtr = entry.Next
			limit++
		}
	} else {
		logger.Printf("%-132s %s skipping empty cleanup handlers...\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_exit"),
		)
	}
	thread.Tcb.Thread.CleanupStack = 0

	thread.Lock.Lock()
	thread.Exited = true
	thread.ExitCode = retValue
	thread.JoinCond.Broadcast()
	thread.Lock.Unlock()
	runtime.Goexit()

	return 0
}
