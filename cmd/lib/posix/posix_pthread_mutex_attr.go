package posix

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/libc"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/pthread"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func Pthread_mutexattr_init(attrHandlePtr *uintptr) uintptr {
	return libScePosix_pthread_mutexattr_init(attrHandlePtr)
}

func libScePosix_pthread_mutexattr_init(attrHandlePtr *uintptr) uintptr {
	attrAddr := GlobalGoAllocator.Malloc(PthreadMutexAttrSize)
	if attrAddr == 0 {
		return ENOMEM
	}

	// Initialize to defaults.
	attr := (*PthreadMutexAttr)(unsafe.Pointer(attrAddr))
	attr.Type = PthreadMutexTypeErrorCheck
	attr.Protocol = PthreadMutexProtocolNone
	attr.Ceiling = 0

	// Copy the pointer back to attrHandlePtr.
	*attrHandlePtr = attrAddr

	logger.Printf("%-132s %s created mutex attribute at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_mutexattr_init"),
		color.Yellow.Sprintf("0x%X", attrAddr),
	)
	return 0
}

func Pthread_mutexattr_settype(attrHandlePtr *uintptr, attrType uintptr) uintptr {
	return libScePosix_pthread_mutexattr_settype(attrHandlePtr, attrType)
}

func libScePosix_pthread_mutexattr_settype(attrHandlePtr *uintptr, attrType uintptr) uintptr {
	if attrType < 1 || attrType > 4 {
		return EINVAL
	}

	// Resolve the handle.
	attr, err := ResolveHandle[PthreadMutexAttr](attrHandlePtr)
	if err != 0 {
		logger.Printf("%-132s %s failed due to invalid attribute pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_mutexattr_settype"),
		)
		emu.SetErrno(err)
		return ERR_PTR
	}

	// Set type.
	attr.Type = PthreadMutexType(attrType)

	logger.Printf("%-132s %s set type to %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_mutexattr_settype"),
		color.Blue.Sprint(MutexTypeNames[attr.Type]),
	)
	return 0
}

func Pthread_mutexattr_destroy(attrHandlePtr *uintptr) uintptr {
	return libScePosix_pthread_mutexattr_destroy(attrHandlePtr)
}

func libScePosix_pthread_mutexattr_destroy(attrHandlePtr *uintptr) uintptr {
	// Resolve the handle.
	attr, err := ResolveHandle[PthreadMutexAttr](attrHandlePtr)
	if err != 0 {
		logger.Printf("%-132s %s failed due to invalid attribute pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_mutexattr_destroy"),
		)
		return err
	}

	// Free the memory.
	attrAddr := uintptr(unsafe.Pointer(attr))
	if !GlobalGoAllocator.Free(attrAddr) {
		logger.Printf("%-132s %s failed freeing untracked pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_mutexattr_destroy"),
		)
		return EFAULT
	}

	// Copy NULL pointer to attrHandlePtr.
	*attrHandlePtr = 0

	logger.Printf("%-132s %s destroyed mutex attribute %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_mutexattr_destroy"),
		color.Yellow.Sprintf("0x%X", attrAddr),
	)
	return 0
}
