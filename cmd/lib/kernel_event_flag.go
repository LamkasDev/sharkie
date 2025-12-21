package lib

import (
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x00000000000231C0
// __int64 __fastcall sceKernelCreateEventFlag(_QWORD *, __int64, unsigned int, __int64, __int64)
func libKernel_sceKernelCreateEventFlag(efHandlePtr, namePtr, attr, initPattern, optParamPtr uintptr) uintptr {
	// This is correct, btw.
	if efHandlePtr == 0 || optParamPtr != 0 {
		fmt.Printf("%-120s %s failed due to invalid pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelCreateEventFlag"),
		)
		return SCE_KERNEL_ERROR_EINVAL
	}

	handle := libKernel_sys_evf_create(namePtr, uint32(attr), uint64(initPattern))
	if handle == ERR_PTR {
		return GetErrno() - 0x7FFE0000
	}

	efHandleSlice := unsafe.Slice((*byte)(unsafe.Pointer(efHandlePtr)), 4)
	binary.LittleEndian.PutUint32(efHandleSlice, uint32(handle))

	return 0
}

func libKernel_sys_evf_create(namePtr uintptr, attr uint32, initPattern uint64) uintptr {
	name := "unnamed"
	if namePtr != 0 {
		name = ReadCString(namePtr)
	}
	if len(name) >= EVF_NAME_MAX {
		fmt.Printf("%-120s %s failed due to too long name.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sys_evf_create"),
		)
		SetErrno(ENAMETOOLONG)
		return ERR_PTR
	}

	queueMode := attr & 0xF
	if queueMode != EVF_ATTR_TH_FIFO && queueMode != 0 {
		fmt.Printf("%-120s %s requesting unknown queue mode %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sys_evf_create"),
			color.Yellow.Sprintf("0x%X", queueMode),
		)
	}

	ef := &EventFlag{
		Name:           name,
		Attributes:     attr,
		CurrentPattern: initPattern,
		InitialPattern: initPattern,
	}
	efHandle := AddEventFlag(ef)

	fmt.Printf("%-120s %s created an event flag %s (name=%s, attr=%s, initPattern=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sys_evf_create"),
		color.Green.Sprintf("%d", efHandle),
		color.Blue.Sprint(name),
		color.Yellow.Sprintf("0x%X", attr),
		color.Yellow.Sprintf("0x%X", initPattern),
	)
	return uintptr(efHandle)
}
