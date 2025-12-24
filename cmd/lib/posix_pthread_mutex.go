package lib

import (
	"encoding/binary"
	"runtime"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x000000000002FDF0
// __int64 __fastcall pthread_mutex_init(__int64, _QWORD *)
func libKernel_pthread_mutex_init(mutexHandlePtr uintptr, attrHandlePtr uintptr) uintptr {
	mutexAddr := GlobalGoAllocator.Malloc(PthreadMutexSize)
	if mutexAddr == 0 {
		return ENOMEM
	}

	// Initialize to defaults.
	mutex := (*PthreadMutex)(unsafe.Pointer(mutexAddr))
	mutex.Lock = 0
	mutex.Flags = uint32(PthreadMutexTypeRecursive)
	mutex.Owner = 0
	mutex.Count = 0
	mutex.SpinLoops = 0
	mutex.YieldLoops = 0
	mutex.Protocol = PthreadMutexProtocolNone

	// Apply attributes.
	attr, err := ResolveHandle[PthreadMutexAttr](attrHandlePtr)
	if err == 0 {
		if attr.Type < PthreadMutexTypeErrorCheck || attr.Type > PthreadMutexTypeAdaptiveNp ||
			attr.Protocol > PthreadMutexProtocolProtect {
			tempHandleAddr := mutexAddr
			if err = libKernel_pthread_mutex_destroy(tempHandleAddr); err != 0 {
				return err
			}
			logger.Printf("%-120s %s failed due to invalid attribute.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("pthread_mutex_init"),
			)
			return EINVAL
		}

		mutex.Flags = uint32(attr.Type)
		mutex.Protocol = attr.Protocol
		if attr.Type == PthreadMutexTypeAdaptiveNp {
			mutex.SpinLoops = 2000
		}
	}

	// Copy the pointer back to mutexHandlePtr.
	mutexHandlePtrSlice := unsafe.Slice((*byte)(unsafe.Pointer(mutexHandlePtr)), 8)
	binary.LittleEndian.PutUint64(mutexHandlePtrSlice, uint64(mutexAddr))
	logger.Printf("%-120s %s created mutex at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_mutex_init"),
		color.Yellow.Sprintf("0x%X", mutexAddr),
	)

	return 0
}

func libKernel_initStaticMutex(mutexHandlePtr uintptr, initType uintptr) uintptr {
	mutexAddr := GlobalGoAllocator.Malloc(PthreadMutexSize)
	if mutexAddr == 0 {
		return ENOMEM
	}

	// Initialize to defaults.
	mutex := (*PthreadMutex)(unsafe.Pointer(mutexAddr))
	mutex.Lock = 0
	if initType == ThrAdaptiveMutexInitializer {
		mutex.Flags = uint32(PthreadMutexTypeAdaptiveNp)
		mutex.SpinLoops = 2000
	} else {
		mutex.Flags = uint32(PthreadMutexTypeRecursive)
		mutex.SpinLoops = 0
	}
	mutex.Owner = 0
	mutex.Count = 0
	mutex.YieldLoops = 0
	mutex.Protocol = PthreadMutexProtocolNone

	// Copy the pointer back to mutexHandlePtr.
	mutexHandlePtrSlice := unsafe.Slice((*byte)(unsafe.Pointer(mutexHandlePtr)), 8)
	binary.LittleEndian.PutUint64(mutexHandlePtrSlice, uint64(mutexAddr))

	logger.Printf("%-120s %s created mutex at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("libKernel_initStaticMutex"),
		color.Yellow.Sprintf("0x%X", mutexAddr),
	)
	return 0
}

// 0x0000000000030CB0
// __int64 __fastcall pthread_mutex_destroy(__int64 *)
func libKernel_pthread_mutex_destroy(mutexHandlePtr uintptr) uintptr {
	// Resolve the handle.
	mutex, err := ResolveHandle[PthreadMutex](mutexHandlePtr)
	if err != 0 {
		logger.Printf("%-120s %s failed due to invalid mutex pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_mutex_destroy"),
		)
		return err
	}

	// Free the memory.
	mutexAddr := uintptr(unsafe.Pointer(mutex))
	if !GlobalGoAllocator.Free(mutexAddr, PthreadMutexSize) {
		logger.Printf("%-120s %s failed freeing untracked pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_mutex_destroy"),
		)
		return EFAULT
	}

	logger.Printf("%-120s %s destroyed mutex %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_mutex_destroy"),
		color.Yellow.Sprintf("0x%X", mutexAddr),
	)
	return 0
}

// TODO: shouldn't we do the other one too?
// 0x0000000000031B40
// __int64 __fastcall sub_31B40(__int64 *, __int64, int, int, int, int)
func libKernel_pthread_mutex_lock(mutexHandlePtr uintptr) uintptr {
	if mutexHandlePtr == 0 {
		return EINVAL
	}

	// Try initializing a mutex, if it wasn't initialized yet.
	currentThread := emu.GetCurrentThread()
	mutexAddr := *(*uintptr)(unsafe.Pointer(mutexHandlePtr))
	if mutexAddr <= ThrMutexDestroyed {
		if mutexAddr == ThrMutexDestroyed {
			logger.Printf("%-120s %s failed trying to lock destroyed mutex (thread %s).\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("pthread_mutex_lock"),
				color.Yellow.Sprintf("0x%X", currentThread),
			)
			return EINVAL
		}
		if err := libKernel_initStaticMutex(mutexHandlePtr, mutexAddr); err != 0 {
			return err
		}
		mutexAddr = *(*uintptr)(unsafe.Pointer(mutexHandlePtr))
	}

	// Process special mutex types.
	mutex := (*PthreadMutex)(unsafe.Pointer(mutexAddr))
	if mutex.Owner == currentThread {
		mutexType := mutex.Flags & PthreadMutexTypeMask
		switch mutexType {
		case uint32(PthreadMutexTypeAdaptiveNp):
			if mutex.Count+1 > 0 {
				mutex.Count++
				logger.Printf("%-120s %s incremented adaptive mutex %s (thread=%s, count=%s).\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("pthread_mutex_lock"),
					GetMutexNameText(mutex, mutexAddr),
					color.Yellow.Sprintf("0x%X", currentThread),
					color.Green.Sprintf("%d", mutex.Count),
				)
				return 0
			}
			logger.Printf("%-120s %s incremented invalid adaptive mutex %s (thread=%s, count=%s).\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("pthread_mutex_lock"),
				GetMutexNameText(mutex, mutexAddr),
				color.Yellow.Sprintf("0x%X", currentThread),
				color.Green.Sprintf("%d", mutex.Count),
			)
			return EAGAIN
		case uint32(PthreadMutexTypeRecursive):
			if mutex.Count+1 > 0 {
				mutex.Count++
				logger.Printf("%-120s %s incremented recursive mutex %s (thread=%s, count=%s).\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("pthread_mutex_lock"),
					GetMutexNameText(mutex, mutexAddr),
					color.Yellow.Sprintf("0x%X", currentThread),
					color.Green.Sprintf("%d", mutex.Count),
				)
				return 0
			}
			logger.Printf("%-120s %s incremented invalid recursive mutex %s (thread=%s, count=%s).\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("pthread_mutex_lock"),
				GetMutexNameText(mutex, mutexAddr),
				color.Yellow.Sprintf("0x%X", currentThread),
				color.Green.Sprintf("%d", mutex.Count),
			)
			return EAGAIN
		case uint32(PthreadMutexTypeErrorCheck):
			logger.Printf("%-120s %s tried to lock a mutex %s it already owns (thread=%s).\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("pthread_mutex_lock"),
				GetMutexNameText(mutex, mutexAddr),
				color.Yellow.Sprintf("0x%X", currentThread),
			)
			return EDEADLK
		default:
			// We should just deadlock here, but let's be nice.
			logger.Printf("%-120s %s tried to lock a mutex %s it already owns (thread=%s).\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("pthread_mutex_lock"),
				GetMutexNameText(mutex, mutexAddr),
				color.Yellow.Sprintf("0x%X", currentThread),
			)
			return EDEADLK
		}
	}

	hostMutex := GetMutex(mutexAddr)

	// For adaptive mutexes, spin for a bit.
	if mutex.Protocol == PthreadMutexProtocolNone {
		spinCount := mutex.SpinLoops
		for spinCount > 0 {
			if hostMutex.TryLock() {
				mutex.Owner = currentThread
				logger.Printf("%-120s %s locked mutex %s (thread=%s).\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("pthread_mutex_lock"),
					GetMutexNameText(mutex, mutexAddr),
					color.Yellow.Sprintf("0x%X", currentThread),
				)
				return 0
			}
			spinCount--
		}

		yieldCount := mutex.YieldLoops
		for yieldCount > 0 {
			runtime.Gosched()
			if hostMutex.TryLock() {
				mutex.Owner = currentThread
				logger.Printf("%-120s %s locked mutex %s (thread=%s).\n",
					emu.GlobalModuleManager.GetCallSiteText(),
					color.Magenta.Sprint("pthread_mutex_lock"),
					GetMutexNameText(mutex, mutexAddr),
					color.Yellow.Sprintf("0x%X", currentThread),
				)
				return 0
			}
			yieldCount--
		}
	}

	// Fallback to a blocking lock.
	hostMutex.Lock()
	mutex.Owner = currentThread

	logger.Printf("%-120s %s locked mutex %s (thread=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_mutex_lock"),
		GetMutexNameText(mutex, mutexAddr),
		color.Yellow.Sprintf("0x%X", currentThread),
	)
	return 0
}

// 0x00000000000327F0
// __int64 __fastcall sub_327F0(__int64 *, __int64, __int64, __int64)
func libKernel_pthread_mutex_unlock(mutexHandlePtr uintptr) uintptr {
	if mutexHandlePtr == 0 {
		return EINVAL
	}

	// Try initializing a mutex, if it wasn't initialized yet.
	currentThread := emu.GetCurrentThread()
	mutexAddr := *(*uintptr)(unsafe.Pointer(mutexHandlePtr))
	if mutexAddr <= ThrMutexDestroyed {
		if mutexAddr == ThrMutexDestroyed {
			logger.Printf("%-120s %s failed trying to unlock destroyed mutex (thread=%s).\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("pthread_mutex_unlock"),
				color.Yellow.Sprintf("0x%X", currentThread),
			)
			return EINVAL
		}
		logger.Printf("%-120s %s failed trying to unlock uninitialized mutex (thread=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_mutex_unlock"),
			color.Yellow.Sprintf("0x%X", currentThread),
		)
		return EPERM
	}

	// Check permissions.
	mutex := (*PthreadMutex)(unsafe.Pointer(mutexAddr))
	if mutex.Owner != currentThread {
		logger.Printf("%-120s %s failed trying to unlock unowned mutex (thread=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_mutex_unlock"),
			color.Yellow.Sprintf("0x%X", currentThread),
		)
		return EPERM
	}

	// Handle special mutex types.
	mutexType := mutex.Flags & PthreadMutexTypeMask
	if mutexType == uint32(PthreadMutexTypeAdaptiveNp) && mutex.Count > 0 {
		mutex.Count--
		logger.Printf("%-120s %s decremented adaptive mutex %s (thread=%s, count=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_mutex_unlock"),
			GetMutexNameText(mutex, mutexAddr),
			color.Yellow.Sprintf("0x%X", currentThread),
			color.Green.Sprintf("%d", mutex.Count),
		)
		return 0
	}
	if mutexType == uint32(PthreadMutexTypeRecursive) && mutex.Count > 0 {
		mutex.Count--
		logger.Printf("%-120s %s decremented recursive mutex %s (thread=%s, count=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_mutex_unlock"),
			GetMutexNameText(mutex, mutexAddr),
			color.Yellow.Sprintf("0x%X", currentThread),
			color.Green.Sprintf("%d", mutex.Count),
		)
		return 0
	}

	// Unlock the mutex.
	mutex.Owner = 0
	hostMutex := GetMutex(mutexAddr)
	hostMutex.Unlock()

	logger.Printf("%-120s %s unlocked mutex %s (thread=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_mutex_unlock"),
		GetMutexNameText(mutex, mutexAddr),
		color.Yellow.Sprintf("0x%X", currentThread),
	)
	return 0
}
