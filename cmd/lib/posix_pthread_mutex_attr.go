package lib

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs/pthread"
	"github.com/gookit/color"
)

// 0x0000000000009360
// __int64 __fastcall pthread_mutexattr_init(__int64 *)
func libKernel_pthread_mutexattr_init(attrHandlePtr uintptr) uintptr {
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
	WriteAddress(attrHandlePtr, attrAddr)

	logger.Printf("%-132s %s created mutex attribute at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_mutexattr_init"),
		color.Yellow.Sprintf("0x%X", attrAddr),
	)
	return 0
}

// 0x0000000000009450
// __int64 __fastcall pthread_mutexattr_settype(_DWORD **, int)
func libKernel_pthread_mutexattr_settype(attrHandlePtr, attrType uintptr) uintptr {
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
		return err
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

// 0x0000000000009490
// __int64 __fastcall scePthreadMutexattrDestroy(__int64 *)
func libKernel_pthread_mutexattr_destroy(attrHandlePtr uintptr) uintptr {
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
	WriteAddress(attrHandlePtr, 0)

	logger.Printf("%-132s %s destroyed mutex attribute %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_mutexattr_destroy"),
		color.Yellow.Sprintf("0x%X", attrAddr),
	)
	return 0
}
