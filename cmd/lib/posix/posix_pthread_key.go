package posix

import (
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

func Pthread_key_create(keyPtr, destructor uintptr) uintptr {
	return libScePosix_pthread_key_create(keyPtr, destructor)
}

func libScePosix_pthread_key_create(keyPtr, destructor uintptr) uintptr {
	GlobalThreadKeyLock.Lock()
	defer GlobalThreadKeyLock.Unlock()

	GlobalThreadKeyCounter++
	newKey := GlobalThreadKeyCounter
	binary.LittleEndian.PutUint32(unsafe.Slice((*byte)(unsafe.Pointer(keyPtr)), 4), newKey)

	logger.Printf("%-132s %s created a new key %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_key_create"),
		color.Yellow.Sprintf("0x%X", newKey),
	)
	return 0
}

func Pthread_key_delete(key uint32) uintptr {
	return libScePosix_pthread_key_delete(key)
}

func libScePosix_pthread_key_delete(key uint32) uintptr {
	GlobalThreadKeyLock.Lock()
	defer GlobalThreadKeyLock.Unlock()

	isValid := key > 0 && key <= GlobalThreadKeyCounter
	if !isValid {
		logger.Printf("%-132s %s failed due to invalid key %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_key_delete"),
			color.Yellow.Sprintf("0x%X", key),
		)
		return EINVAL
	}

	thread := emu.GetCurrentThread()
	thread.Lock.Lock()
	delete(thread.KeyValues, key)
	thread.Lock.Unlock()

	logger.Printf("%-132s %s deleted key %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_key_delete"),
		color.Yellow.Sprintf("0x%X", key),
	)
	return 0
}

func Pthread_getspecific(key uint32) uintptr {
	return libScePosix_pthread_getspecific(key)
}

func libScePosix_pthread_getspecific(key uint32) uintptr {
	thread := emu.GetCurrentThread()
	thread.Lock.RLock()
	value, ok := thread.KeyValues[key]
	thread.Lock.RUnlock()
	if !ok {
		return 0
	}

	return value
}

func Pthread_setspecific(key uint32, value uintptr) uintptr {
	return libScePosix_pthread_setspecific(key, value)
}

func libScePosix_pthread_setspecific(key uint32, value uintptr) uintptr {
	thread := emu.GetCurrentThread()
	thread.Lock.Lock()
	thread.KeyValues[key] = value
	thread.Lock.Unlock()

	return 0
}
