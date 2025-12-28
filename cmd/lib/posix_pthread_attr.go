package lib

import (
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x0000000000003BC0
// __int64 __fastcall pthread_attr_init(__int64 *)
func libKernel_pthread_attr_init(attrHandlePtr uintptr) uintptr {
	attrAddr := GlobalGoAllocator.Malloc(PthreadAttrSize)
	if attrAddr == 0 {
		return ENOMEM
	}

	// Initialize to defaults.
	attr := (*PthreadAttr)(unsafe.Pointer(attrAddr))
	attr.SchedulingPolicy = PthreadSchedulingPolicyFifo
	attr.SchedulingInherit = int32(PthreadFlagsInheritSched)
	attr.Priority = 700
	attr.Flags = PthreadFlagsScopeSystem
	attr.StackSize = 0x100000

	// Copy the pointer back to attrHandlePtr.
	attrHandlePtrSlice := unsafe.Slice((*byte)(unsafe.Pointer(attrHandlePtr)), 8)
	binary.LittleEndian.PutUint64(attrHandlePtrSlice, uint64(attrAddr))

	logger.Printf("%-120s %s created attribute at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_attr_init"),
		color.Yellow.Sprintf("0x%X", attrAddr),
	)
	return 0
}

// 0x0000000000003800
// __int64 __fastcall pthread_attr_destroy(__int64 *)
func libKernel_pthread_attr_destroy(attrHandlePtr uintptr) uintptr {
	// Resolve the handle.
	attr, err := ResolveHandle[PthreadAttr](attrHandlePtr)
	if err != 0 {
		logger.Printf("%-120s %s failed due to invalid attribute pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_attr_destroy"),
		)
		return err
	}

	// Free the memory.
	attrAddr := uintptr(unsafe.Pointer(attr))
	if !GlobalGoAllocator.Free(attrAddr) {
		logger.Printf("%-120s %s failed freeing untracked pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_attr_destroy"),
		)
		return EFAULT
	}

	// Copy NULL pointer to attrHandlePtr.
	attrHandlePtrSlice := unsafe.Slice((*byte)(unsafe.Pointer(attrHandlePtr)), 8)
	binary.LittleEndian.PutUint64(attrHandlePtrSlice, 0)

	logger.Printf("%-120s %s destroyed struct at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_attr_destroy"),
		color.Yellow.Sprintf("0x%X", attrAddr),
	)
	return 0
}
