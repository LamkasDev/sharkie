package lib

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs/pthread"
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
			logger.Printf("%-132s %s failed due to invalid attribute.\n",
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
	WriteAddress(mutexHandlePtr, mutexAddr)

	logger.Printf("%-132s %s created mutex at %s.\n",
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
	WriteAddress(mutexHandlePtr, mutexAddr)

	logger.Printf("%-132s %s created mutex at %s.\n",
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
		logger.Printf("%-132s %s failed due to invalid mutex pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_mutex_destroy"),
		)
		return err
	}

	// Free the memory.
	mutexAddr := uintptr(unsafe.Pointer(mutex))
	if !GlobalGoAllocator.Free(mutexAddr) {
		logger.Printf("%-132s %s failed freeing untracked pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_mutex_destroy"),
		)
		return EFAULT
	}

	logger.Printf("%-132s %s destroyed mutex %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_mutex_destroy"),
		GetMutexNameText(mutex, mutexAddr),
	)
	return 0
}
