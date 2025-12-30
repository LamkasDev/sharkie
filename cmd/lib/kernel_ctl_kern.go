package lib

import (
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

const KERN_PROC = 14
const KERN_SMP = 24
const KERN_USRSTACK = 33

const KERN_PROC_APPINFO = 35
const KERN_PROC_PTC = 42

func libKernel_ctl_kern(mib []uint32, namePtr uintptr, nameLen uint32, oldPtr uintptr, oldLenPtr uintptr, newPtr uintptr, newLen uintptr) (uintptr, bool) {
	switch mib[1] {
	case KERN_PROC:
		return libKernel_ctl_kern_proc(mib, namePtr, nameLen, oldPtr, oldLenPtr, newPtr, newLen), true
	case KERN_SMP:
		return libKernel_ctl_kern_smp(mib, namePtr, nameLen, oldPtr, oldLenPtr, newPtr, newLen), true
	case KERN_USRSTACK:
		return libKernel_ctl_usrstack(mib, namePtr, nameLen, oldPtr, oldLenPtr, newPtr, newLen), true
	}

	return ENOENT, false
}

func libKernel_ctl_kern_proc(mib []uint32, namePtr uintptr, nameLen uint32, oldPtr uintptr, oldLenPtr uintptr, newPtr uintptr, newLen uintptr) uintptr {
	if oldLenPtr == 0 || oldPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sysctl"),
		)
		return 0
	}
	if len(mib) < 3 {
		logger.Printf("%-132s %s failed due to short MIBs.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sysctl"),
		)
		return EINVAL
	}
	oldLenSlice := unsafe.Slice((*byte)(unsafe.Pointer(oldLenPtr)), 8)
	providedSize := uintptr(binary.LittleEndian.Uint64(oldLenSlice))

	switch mib[2] {
	case KERN_PROC_APPINFO:
		requiredSize := uintptr(72)
		if providedSize < requiredSize {
			logger.Printf("%-132s %s failed due to insufficient pointer size.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sysctl"),
			)
			return ENOMEM
		}
		oldSlice := unsafe.Slice((*byte)(unsafe.Pointer(oldPtr)), requiredSize)

		for i := uintptr(0); i < requiredSize; i++ {
			oldSlice[i] = 0
		}
		binary.LittleEndian.PutUint32(oldSlice, mib[3])
		binary.LittleEndian.PutUint64(oldLenSlice, uint64(requiredSize))

		logger.Printf("%-132s %s requested app info for process %s (oldPtr=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sysctl"),
			color.Yellow.Sprintf("0x%X", mib[3]),
			color.Yellow.Sprintf("0x%X", oldPtr),
		)
		return 0
	case KERN_PROC_PTC:
		requiredSize := uintptr(16)
		if providedSize < 8 {
			logger.Printf("%-132s %s failed due to insufficient pointer size.\n",
				emu.GlobalModuleManager.GetCallSiteText(),
				color.Magenta.Sprint("sysctl"),
			)
			return ENOMEM
		}
		oldSlice := unsafe.Slice((*byte)(unsafe.Pointer(oldPtr)), requiredSize)

		counter := uint64(0)
		freq := PTC_FREQUENCY
		binary.LittleEndian.PutUint64(oldSlice, counter) // Current counter.
		if providedSize >= 16 {
			binary.LittleEndian.PutUint64(oldSlice[8:], freq) // Counter frequency.
		}
		binary.LittleEndian.PutUint64(oldLenSlice, uint64(requiredSize))

		logger.Printf("%-132s %s requested process time counter %s with frequency %s (oldPtr=%s).\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sysctl"),
			color.Green.Sprintf("%d", counter),
			color.Yellow.Sprintf("0x%X", freq),
			color.Yellow.Sprintf("0x%X", oldPtr),
		)
		return 0
	}

	logger.Printf("%-132s %s failed due to unknown OIDs %+v.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sysctl"),
		mib,
	)
	return ENOENT
}

func libKernel_ctl_kern_smp(mib []uint32, namePtr uintptr, nameLen uint32, oldPtr uintptr, oldLenPtr uintptr, newPtr uintptr, newLen uintptr) uintptr {
	if oldLenPtr == 0 || oldPtr == 0 {
		return 0
	}
	oldLenSlice := unsafe.Slice((*byte)(unsafe.Pointer(oldLenPtr)), 8)
	providedSize := uintptr(binary.LittleEndian.Uint64(oldLenSlice))

	requiredSize := uintptr(4)
	if providedSize < requiredSize {
		logger.Printf("%-132s %s failed due to insufficient pointer size.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sysctl"),
		)
		return ENOMEM
	}
	oldSlice := unsafe.Slice((*byte)(unsafe.Pointer(oldPtr)), requiredSize)

	cpus := 4
	binary.LittleEndian.PutUint32(oldSlice, uint32(cpus))
	binary.LittleEndian.PutUint64(oldLenSlice, uint64(requiredSize))

	logger.Printf("%-132s %s requested number of cores %s (oldPtr=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sysctl"),
		color.Green.Sprintf("%d", cpus),
		color.Yellow.Sprintf("0x%X", oldPtr),
	)
	return 0
}

func libKernel_ctl_usrstack(mib []uint32, namePtr uintptr, nameLen uint32, oldPtr uintptr, oldLenPtr uintptr, newPtr uintptr, newLen uintptr) uintptr {
	if oldLenPtr == 0 || oldPtr == 0 {
		return 0
	}
	oldLenSlice := unsafe.Slice((*byte)(unsafe.Pointer(oldLenPtr)), 8)
	providedSize := uintptr(binary.LittleEndian.Uint64(oldLenSlice))

	requiredSize := uintptr(8)
	if providedSize < requiredSize {
		logger.Printf("%-132s %s failed due to insufficient pointer size.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sysctl"),
		)
		return ENOMEM
	}
	oldSlice := unsafe.Slice((*byte)(unsafe.Pointer(oldPtr)), requiredSize)

	thread := emu.GetCurrentThread()
	stackTop := thread.Stack.ArgumentsAddress
	binary.LittleEndian.PutUint64(oldSlice, uint64(stackTop))
	binary.LittleEndian.PutUint64(oldLenSlice, uint64(requiredSize))

	logger.Printf("%-132s %s requested stack top %s (oldPtr=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sysctl"),
		color.Yellow.Sprintf("0x%X", stackTop),
		color.Yellow.Sprintf("0x%X", oldPtr),
	)
	return 0
}
