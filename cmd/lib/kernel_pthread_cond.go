package lib

import (
	"fmt"
	"strings"
	"unsafe"

	. "github.com/LamkasDev/sharkie/cmd/structs"
	. "github.com/LamkasDev/sharkie/cmd/structs/pthread"
)

// 0x00000000000137A0
// __int64 scePthreadCondInit()
func libKernel_scePthreadCondInit(condHandlePtr, attrHandlePtr uintptr, namePtr uintptr) uintptr {
	err := libKernel_pthread_cond_init(condHandlePtr, attrHandlePtr)
	if err != 0 {
		return err - SonyErrorOffset
	}

	// Retrieve structs back.
	condAddr := *(*uintptr)(unsafe.Pointer(condHandlePtr))
	cond := (*PthreadCond)(unsafe.Pointer(condAddr))

	// Set name.
	var name string
	if namePtr != 0 {
		name = ReadCString(namePtr)
	} else {
		name = fmt.Sprintf("Cond_%x", condAddr)
	}
	cond.Name = strings.Clone(name)

	return 0
}

// 0x0000000000013840
// __int64 scePthreadCondDestroy()
func libKernel_scePthreadCondDestroy(condHandlePtr uintptr) uintptr {
	err := libKernel_pthread_cond_destroy(condHandlePtr)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x0000000000013780
// __int64 __fastcall scePthreadCondBroadcast(__int64 *, __int64, int, int, int, int)
func libKernel_scePthreadCondBroadcast(condHandlePtr uintptr) uintptr {
	err := libKernel_pthread_cond_broadcast(condHandlePtr)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x0000000000013860
// __int64 scePthreadCondSignal()
func libKernel_scePthreadCondSignal(condHandlePtr uintptr) uintptr {
	err := libKernel_pthread_cond_signal(condHandlePtr)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x00000000000138C0
// __int64 scePthreadCondWait()
func libKernel_scePthreadCondWait(condHandlePtr uintptr, mutexHandlePtr uintptr) uintptr {
	err := libKernel_pthread_cond_wait(condHandlePtr, mutexHandlePtr)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}

// 0x00000000000138A0
// __int64 scePthreadCondTimedwait()
func libKernel_scePthreadCondTimedwait(condHandlePtr uintptr, mutexHandlePtr uintptr, micros uintptr) uintptr {
	err := libKernel_pthread_cond_reltimedwait_np(condHandlePtr, mutexHandlePtr, micros)
	if err != 0 {
		return err - SonyErrorOffset
	}

	return 0
}
