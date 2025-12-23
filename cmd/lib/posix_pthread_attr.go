package lib

import (
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/LamkasDev/sharkie/cmd/sys_struct"
	"github.com/gookit/color"
)

// 0x0000000000003BC0
// __int64 __fastcall pthread_attr_init(__int64 *)
func libKernel_pthread_attr_init(attrHandlePtr uintptr) uintptr {
	attrAddr, _ := sys_struct.AllocReadWriteMemory(unsafe.Sizeof(PthreadAttr{}))
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
	fmt.Printf("%-120s %s created struct at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_attr_init"),
		color.Yellow.Sprintf("0x%X", attrAddr),
	)

	return 0
}
