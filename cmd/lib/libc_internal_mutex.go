package lib

import (
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x000000000001BEB0
// __int64 __fastcall Mtxinit(__int64, __int64)
func libSceLibcInternal__Mtxinit(mutexHandlePtr uintptr, nameSuffixPtr uintptr) uintptr {
	// Allocate proper name.
	name := "SceLibcI"
	if nameSuffixPtr != 0 {
		name = fmt.Sprintf("SceLibcI_%s", ReadCString(nameSuffixPtr))
	}
	nameSlice := append([]byte(name), 0)
	namePtr := uintptr(unsafe.Pointer(&nameSlice[0]))

	// Create a mutex with valid attribute and name.
	var attrHandle uint64
	attrHandlePtr := uintptr(unsafe.Pointer(&attrHandle))
	if libKernel_scePthreadMutexattrInit(attrHandlePtr) != 0 {
		return 1
	}
	if libKernel_scePthreadMutexattrSettype(attrHandlePtr, 2) != 0 {
		libKernel_scePthreadMutexattrDestroy(attrHandlePtr)
		return 1
	}
	initErr := libKernel_scePthreadMutexInit(mutexHandlePtr, attrHandlePtr, namePtr)
	destroyErr := libKernel_scePthreadMutexattrDestroy(attrHandlePtr)
	if initErr == 0 {
		// TODO: move print to libKernel_scePthreadMutexInit.
		fmt.Printf("%-120s %s created mutex named %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("_Mtxinit"),
			color.Blue.Sprint(name),
		)
		return destroyErr
	}

	return 1
}
