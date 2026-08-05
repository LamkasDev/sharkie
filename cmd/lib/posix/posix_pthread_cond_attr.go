package posix

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/libc"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/time"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func Pthread_condattr_init(attrHandlePtr *uintptr) uintptr {
	return libScePosix_pthread_mutexattr_init(attrHandlePtr)
}

func libScePosix_pthread_condattr_init(attrHandlePtr *uintptr) uintptr {
	attrAddr := GlobalGoAllocator.Malloc(PthreadCondAttrSize)
	if attrAddr == 0 {
		return ENOMEM
	}

	// Initialize to defaults.
	attr := (*PthreadCondAttr)(unsafe.Pointer(attrAddr))
	attr.Shared = 0
	attr.ClockId = ClockIdRealtime

	// Copy the pointer back to attrHandlePtr.
	*attrHandlePtr = attrAddr

	logger.Printf("%-132s %s created cond attribute at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_condattr_init"),
		color.Yellow.Sprintf("0x%X", attrAddr),
	)
	return 0
}

func Pthread_condattr_destroy(attrHandlePtr *uintptr) uintptr {
	return libScePosix_pthread_condattr_destroy(attrHandlePtr)
}

func libScePosix_pthread_condattr_destroy(attrHandlePtr *uintptr) uintptr {
	// Resolve the handle.
	attr, err := ResolveHandle[PthreadCondAttr](attrHandlePtr)
	if err != 0 {
		logger.Printf("%-132s %s failed due to invalid attribute pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_condattr_destroy"),
		)
		return err
	}

	// Free the memory.
	attrAddr := uintptr(unsafe.Pointer(attr))
	if !GlobalGoAllocator.Free(attrAddr) {
		logger.Printf("%-132s %s failed freeing untracked pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_condattr_destroy"),
		)
		return EFAULT
	}

	// Copy NULL pointer to attrHandlePtr.
	*attrHandlePtr = 0

	logger.Printf("%-132s %s destroyed cond attribute %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_condattr_destroy"),
		color.Yellow.Sprintf("0x%X", attrAddr),
	)
	return 0
}
