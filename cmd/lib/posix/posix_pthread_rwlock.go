package posix

import (
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/libc"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func InitStaticRwlock(rwlockHandlePtr *uintptr) uintptr {
	rwlockAddr := GlobalGoAllocator.Malloc(PthreadRwlockSize)
	if rwlockAddr == 0 {
		return ENOMEM
	}

	// Initialize to defaults.
	rwlock := (*PthreadRwlock)(unsafe.Pointer(rwlockAddr))
	rwlock.Name = ""

	// Copy the pointer back to rwlockHandlePtr.
	*rwlockHandlePtr = rwlockAddr

	logger.Printf("%-132s %s created rwlock at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("InitStaticRwlock"),
		color.Yellow.Sprintf("0x%X", rwlockAddr),
	)
	return 0
}

func Pthread_rwlock_init(rwlockHandlePtr, attrHandlePtr *uintptr) uintptr {
	return libScePosix_pthread_rwlock_init(rwlockHandlePtr, attrHandlePtr)
}

func libScePosix_pthread_rwlock_init(rwlockHandlePtr, attrHandlePtr *uintptr) uintptr {
	rwlockAddr := GlobalGoAllocator.Malloc(PthreadRwlockSize)
	if rwlockAddr == 0 {
		return ENOMEM
	}

	// Initialize to defaults.
	rwlock := (*PthreadRwlock)(unsafe.Pointer(rwlockAddr))
	rwlock.Name = ""

	// Copy the pointer back to rwlockHandlePtr.
	*rwlockHandlePtr = rwlockAddr

	logger.Printf("%-132s %s created rwlock at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_rwlock_init"),
		color.Yellow.Sprintf("0x%X", rwlockAddr),
	)
	return 0
}
