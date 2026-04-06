package lib

import (
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x0000000000034520
// __int64 __fastcall pthread_key_create(_DWORD *, __int64)
func libKernel_pthread_key_create(keyPtr, destructor uintptr) uintptr {
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

// 0x0000000000034BD0
// __int64 __fastcall pthread_getspecific(unsigned int)
func libKernel_pthread_getspecific(key uint32) uintptr {
	thread := emu.GetCurrentThread()
	thread.Lock.Lock()
	value, ok := thread.KeyValues[key]
	thread.Lock.Unlock()
	if !ok {
		return 0
	}

	return value
}

// 0x0000000000034B70
// __int64 __fastcall pthread_setspecific(__int64, __int64)
func libKernel_pthread_setspecific(key uint32, value uintptr) uintptr {
	thread := emu.GetCurrentThread()
	thread.Lock.Lock()
	thread.KeyValues[key] = value
	thread.Lock.Unlock()

	return 0
}
