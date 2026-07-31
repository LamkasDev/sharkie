package posix

import (
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
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

func Pthread_getspecific(key uint32) uintptr {
	return libScePosix_pthread_getspecific(key)
}

func libScePosix_pthread_getspecific(key uint32) uintptr {
	thread := emu.GetCurrentThread()
	thread.Lock.Lock()
	value, ok := thread.KeyValues[key]
	thread.Lock.Unlock()
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

func Pthread_getschedparam() uintptr {
	return libScePosix_pthread_getschedparam()
}

// TODO: finish this.
func libScePosix_pthread_getschedparam() uintptr {
	return 0
}

func Pthread_setschedparam() uintptr {
	return libScePosix_pthread_setschedparam()
}

// TODO: finish this.
func libScePosix_pthread_setschedparam() uintptr {
	return 0
}
