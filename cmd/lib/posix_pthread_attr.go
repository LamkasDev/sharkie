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
	attr.InheritScheduling = PthreadInheritSchedulingInherit
	attr.Priority = 700
	attr.Flags = PthreadAttrFlagsScopeSystem
	attr.StackSize = 0x100000

	// Copy the pointer back to attrHandlePtr.
	attrHandlePtrSlice := unsafe.Slice((*byte)(unsafe.Pointer(attrHandlePtr)), 8)
	binary.LittleEndian.PutUint64(attrHandlePtrSlice, uint64(attrAddr))

	logger.Printf("%-132s %s created attribute at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_attr_init"),
		color.Yellow.Sprintf("0x%X", attrAddr),
	)
	return 0
}

// 0x00000000000134C0
// __int64 __fastcall pthread_attr_setstacksize(__int64, unsigned __int64)
func libKernel_pthread_attr_setstacksize(attrHandlePtr uintptr, stackSize uintptr) uintptr {
	if stackSize < 0x4000 {
		logger.Printf("%-132s %s failed due to invalid stack size %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_attr_setstacksize"),
			color.Yellow.Sprintf("0x%X", stackSize),
		)
		return EINVAL
	}

	// Resolve the handle.
	attr, err := ResolveHandle[PthreadAttr](attrHandlePtr)
	if err != 0 {
		logger.Printf("%-132s %s failed due to invalid attribute pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_attr_setstacksize"),
		)
		return err
	}

	// Set stack size.
	attr.StackSize = stackSize

	logger.Printf("%-132s %s set stack size to %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_attr_setstacksize"),
		color.Yellow.Sprintf("0x%X", stackSize),
	)
	return 0
}

// 0x0000000000003D10
// __int64 __fastcall pthread_attr_setschedpolicy(_DWORD **, int)
func libKernel_pthread_attr_setschedpolicy(attrHandlePtr uintptr, schedulingPolicy uintptr) uintptr {
	if schedulingPolicy != 1 && schedulingPolicy != 2 && schedulingPolicy != 3 {
		logger.Printf("%-132s %s failed due to invalid scheduling policy %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_attr_setschedpolicy"),
			color.Yellow.Sprintf("0x%X", schedulingPolicy),
		)
		return EINVAL
	}

	// Resolve the handle.
	attr, err := ResolveHandle[PthreadAttr](attrHandlePtr)
	if err != 0 {
		logger.Printf("%-132s %s failed due to invalid attribute pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_attr_setschedpolicy"),
		)
		return err
	}

	// Set scheduling policy.
	attr.SchedulingPolicy = PthreadSchedulingPolicy(schedulingPolicy)

	logger.Printf("%-132s %s set scheduling policy to %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_attr_setschedpolicy"),
		color.Blue.Sprint(SchedulingPolicyNames[attr.SchedulingPolicy]),
	)
	return 0
}

// 0x0000000000003C90
// __int64 __fastcall pthread_attr_setinheritsched(__int64, int)
func libKernel_pthread_attr_setinheritsched(attrHandlePtr uintptr, inheritScheduling uintptr) uintptr {
	if inheritScheduling != 0 && inheritScheduling != 4 {
		logger.Printf("%-132s %s failed due to invalid inherit scheduling %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_attr_setinheritsched"),
			color.Yellow.Sprintf("0x%X", inheritScheduling),
		)
		return EINVAL
	}

	// Resolve the handle.
	attr, err := ResolveHandle[PthreadAttr](attrHandlePtr)
	if err != 0 {
		logger.Printf("%-132s %s failed due to invalid attribute pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_attr_setinheritsched"),
		)
		return err
	}

	// Set inherit scheduling.
	attr.InheritScheduling = PthreadInheritScheduling(inheritScheduling)

	logger.Printf("%-132s %s set inherit scheduling to %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_attr_setinheritsched"),
		color.Blue.Sprint(InheritSchedulingNames[attr.InheritScheduling]),
	)
	return 0
}

// 0x0000000000003CC0
// __int64 __fastcall pthread_attr_setschedparam(int **, int *)
func libKernel_pthread_attr_setschedparam(attrHandlePtr uintptr, schedulingParameterPtr uintptr) uintptr {
	if schedulingParameterPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid scheduling parameter pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_attr_setschedparam"),
		)
		return EINVAL
	}
	schedulingParameter := *(*int32)(unsafe.Pointer(schedulingParameterPtr))
	if schedulingParameter < 0 || schedulingParameter > 1024 {
		logger.Printf("%-132s %s failed due to invalid scheduling parameter %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_attr_setschedparam"),
			color.Yellow.Sprintf("0x%X", schedulingParameter),
		)
		return EINVAL
	}

	// Resolve the handle.
	attr, err := ResolveHandle[PthreadAttr](attrHandlePtr)
	if err != 0 {
		logger.Printf("%-132s %s failed due to invalid attribute pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_attr_setschedparam"),
		)
		return err
	}

	// Set scheduling parameter.
	attr.Priority = schedulingParameter

	logger.Printf("%-132s %s set scheduling parameter to %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_attr_setschedparam"),
		color.Yellow.Sprintf("0x%X", schedulingParameter),
	)
	return 0
}

// 0x0000000000003C70
// __int64 __fastcall pthread_attr_setguardsize(__int64, __int64)
func libKernel_pthread_attr_setguardsize(attrHandlePtr uintptr, guardSize uintptr) uintptr {
	// Resolve the handle.
	attr, err := ResolveHandle[PthreadAttr](attrHandlePtr)
	if err != 0 {
		logger.Printf("%-132s %s failed due to invalid attribute pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_attr_setguardsize"),
		)
		return err
	}

	// Set guard size.
	attr.GuardSize = guardSize

	logger.Printf("%-132s %s set guard size to %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_attr_setguardsize"),
		color.Yellow.Sprintf("0x%X", guardSize),
	)
	return 0
}

// 0x0000000000003C40
// __int64 __fastcall pthread_attr_setdetachstate(__int64 *, unsigned int)
func libKernel_pthread_attr_setdetachstate(attrHandlePtr uintptr, detachState uintptr) uintptr {
	if detachState != 0 && detachState != 1 {
		logger.Printf("%-132s %s failed due to invalid detach state %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_attr_setdetachstate"),
			color.Yellow.Sprintf("0x%X", detachState),
		)
		return EINVAL
	}

	// Resolve the handle.
	attr, err := ResolveHandle[PthreadAttr](attrHandlePtr)
	if err != 0 {
		logger.Printf("%-132s %s failed due to invalid attribute pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_attr_setdetachstate"),
		)
		return err
	}

	// Set detach state.
	if PthreadDetachState(detachState) == PthreadDetachStateDetached {
		attr.Flags |= PthreadAttrFlagsDetached
	} else {
		attr.Flags &^= PthreadAttrFlagsDetached
	}

	logger.Printf("%-132s %s set detach state to %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_attr_setdetachstate"),
		color.Blue.Sprint(DetachStateNames[PthreadDetachState(detachState)]),
	)
	return 0
}

// 0x0000000000003D50
// __int64 __fastcall pthread_attr_setscope(__int64 *, int)
func libKernel_pthread_attr_setscope(attrHandlePtr uintptr, scope uintptr) uintptr {
	if scope != 0 && scope != 2 {
		logger.Printf("%-132s %s failed due to invalid scope %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_attr_setscope"),
			color.Yellow.Sprintf("0x%X", scope),
		)
		return EINVAL
	}

	// Resolve the handle.
	attr, err := ResolveHandle[PthreadAttr](attrHandlePtr)
	if err != 0 {
		logger.Printf("%-132s %s failed due to invalid attribute pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_attr_setscope"),
		)
		return err
	}

	// Set scope.
	if PthreadScope(scope) == PthreadScopeSystem {
		attr.Flags |= PthreadAttrFlagsScopeSystem
	} else {
		attr.Flags &^= PthreadAttrFlagsScopeSystem
	}

	logger.Printf("%-132s %s set scope to %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_attr_setscope"),
		color.Blue.Sprint(ScopeNames[PthreadScope(scope)]),
	)
	return 0
}

// 0x0000000000003800
// __int64 __fastcall pthread_attr_destroy(__int64 *)
func libKernel_pthread_attr_destroy(attrHandlePtr uintptr) uintptr {
	// Resolve the handle.
	attr, err := ResolveHandle[PthreadAttr](attrHandlePtr)
	if err != 0 {
		logger.Printf("%-132s %s failed due to invalid attribute pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_attr_destroy"),
		)
		return err
	}

	// Free the memory.
	attrAddr := uintptr(unsafe.Pointer(attr))
	if !GlobalGoAllocator.Free(attrAddr) {
		logger.Printf("%-132s %s failed freeing untracked pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("pthread_attr_destroy"),
		)
		return EFAULT
	}

	// Copy NULL pointer to attrHandlePtr.
	attrHandlePtrSlice := unsafe.Slice((*byte)(unsafe.Pointer(attrHandlePtr)), 8)
	binary.LittleEndian.PutUint64(attrHandlePtrSlice, 0)

	logger.Printf("%-132s %s destroyed struct at %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_attr_destroy"),
		color.Yellow.Sprintf("0x%X", attrAddr),
	)
	return 0
}
