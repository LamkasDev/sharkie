package lib

import (
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

const SYSCTL_NAME = 3

func libKernel_ctl_sysctl(mib []uint32, namePtr, nameLen, oldPtr, oldLenPtr, newPtr, newLen uintptr) (uintptr, bool) {
	switch mib[1] {
	case SYSCTL_NAME:
		return libKernel_ctl_sysctl_name(mib, namePtr, nameLen, oldPtr, oldLenPtr, newPtr, newLen), true
	}

	return ENOENT, false
}

func libKernel_ctl_sysctl_name(mib []uint32, namePtr, nameLen, oldPtr, oldLenPtr, newPtr, newLen uintptr) uintptr {
	if oldLenPtr == 0 || oldPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sysctl"),
		)
		return 0
	}
	if newPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid name pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sysctl"),
		)
		return EFAULT
	}
	oldLenSlice := unsafe.Slice((*byte)(unsafe.Pointer(oldLenPtr)), 8)
	providedSize := uintptr(binary.LittleEndian.Uint64(oldLenSlice))

	name := GoString(Cstring(newPtr))
	var resultMib []uint32
	switch name {
	case "kern.smp.cpus":
		resultMib = []uint32{CTL_KERN, KERN_SMP}
	case "kern.proc.ptc":
		resultMib = []uint32{CTL_KERN, KERN_PROC, KERN_PROC_PTC}
	case "kern.proc.appinfo":
		resultMib = []uint32{CTL_KERN, KERN_PROC, KERN_PROC_APPINFO}
	case "kern.usrstack":
		resultMib = []uint32{CTL_KERN, KERN_USRSTACK}
	case "hw.pagesize":
		resultMib = []uint32{CTL_HW, HW_PAGESIZE}
	default:
		logger.Printf("%-132s %s failed due to unknown name %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sysctl"),
			color.Blue.Sprint(name),
		)
		return ENOENT
	}

	requiredSize := uintptr(len(resultMib) * 4)
	if providedSize < requiredSize {
		logger.Printf("%-132s %s failed due to insufficient pointer size.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sysctl"),
		)
		return ENOMEM
	}
	oldSlice := unsafe.Slice((*byte)(unsafe.Pointer(oldPtr)), requiredSize)

	for i, val := range resultMib {
		binary.LittleEndian.PutUint32(oldSlice[i*4:], val)
	}
	binary.LittleEndian.PutUint64(oldLenSlice, uint64(requiredSize))

	logger.Printf("%-132s %s requested MIBs %s (oldPtr=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sysctl"),
		color.Green.Sprintf("%+v", resultMib),
		color.Yellow.Sprintf("0x%X", oldPtr),
	)
	return 0
}
