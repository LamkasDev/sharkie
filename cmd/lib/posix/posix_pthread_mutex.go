package posix

import (
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/libc"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func InitStaticMutex(mutexHandlePtr *uintptr, initType uintptr) uintptr {
	mutexAddr := GlobalGoAllocator.Malloc(PthreadMutexSize)
	if mutexAddr == 0 {
		return ENOMEM
	}

	// Initialize to defaults.
	mutex := (*PthreadMutex)(unsafe.Pointer(mutexAddr))
	if initType == ThrAdaptiveMutexInitializer {
		mutex.Flags = uint32(PthreadMutexTypeAdaptiveNp)
		mutex.SpinLoops = 2000
	} else {
		mutex.Flags = uint32(PthreadMutexTypeErrorCheck)
		mutex.SpinLoops = 0
	}
	mutex.Owner = 0
	mutex.Count = 0
	mutex.YieldLoops = 0
	mutex.Protocol = PthreadMutexProtocolNone
	mutex.Name = fmt.Sprintf("0x%X", mutexAddr)

	// Copy the pointer back to mutexHandlePtr.
	*mutexHandlePtr = mutexAddr

	logger.Printf("%-132s %s created mutex at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("InitStaticMutex"),
		color.Yellow.Sprintf("0x%X", mutexAddr),
	)
	return 0
}

func Pthread_mutex_init(mutexHandlePtr, attrHandlePtr *uintptr) uintptr {
	return libScePosix_pthread_mutex_init(mutexHandlePtr, attrHandlePtr)
}

func libScePosix_pthread_mutex_init(mutexHandlePtr, attrHandlePtr *uintptr) uintptr {
	mutexAddr := GlobalGoAllocator.Malloc(PthreadMutexSize)
	if mutexAddr == 0 {
		return ENOMEM
	}

	// Initialize to defaults.
	mutex := (*PthreadMutex)(unsafe.Pointer(mutexAddr))
	mutex.Flags = uint32(PthreadMutexTypeErrorCheck)
	mutex.Owner = 0
	mutex.Count = 0
	mutex.SpinLoops = 0
	mutex.YieldLoops = 0
	mutex.Protocol = PthreadMutexProtocolNone
	mutex.Name = fmt.Sprintf("0x%X", mutexAddr)

	// Apply attributes.
	attr, err := ResolveHandle[PthreadMutexAttr](attrHandlePtr)
	if err == 0 {
		if attr.Type < PthreadMutexTypeErrorCheck || attr.Type > PthreadMutexTypeAdaptiveNp ||
			attr.Protocol > PthreadMutexProtocolProtect {
			tempHandleAddr := &mutexAddr
			if err = libScePosix_pthread_mutex_destroy(tempHandleAddr); err != 0 {
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
	*mutexHandlePtr = mutexAddr

	logger.Printf("%-132s %s created mutex at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_mutex_init"),
		color.Yellow.Sprintf("0x%X", mutexAddr),
	)
	return 0
}

func Pthread_mutex_destroy(mutexHandlePtr *uintptr) uintptr {
	return libScePosix_pthread_mutex_destroy(mutexHandlePtr)
}

func libScePosix_pthread_mutex_destroy(mutexHandlePtr *uintptr) uintptr {
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
		color.Yellow.Sprintf("0x%X", mutexAddr),
	)
	return 0
}
