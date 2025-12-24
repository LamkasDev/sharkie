package lib

import (
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

const HW_PAGESIZE = 7

func libKernel_ctl_hw(mib []uint32, namePtr uintptr, nameLen uint32, oldPtr uintptr, oldLenPtr uintptr, newPtr uintptr, newLen uintptr) (uintptr, bool) {
	switch mib[1] {
	case HW_PAGESIZE:
		return libKernel_ctl_hw_pagesize(mib, namePtr, nameLen, oldPtr, oldLenPtr, newPtr, newLen), true
	}

	return ENOENT, false
}

func libKernel_ctl_hw_pagesize(mib []uint32, namePtr uintptr, nameLen uint32, oldPtr uintptr, oldLenPtr uintptr, newPtr uintptr, newLen uintptr) uintptr {
	if oldLenPtr == 0 || oldPtr == 0 {
		logger.Printf("%-120s %s failed due to invalid pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sysctl"),
		)
		return 0
	}
	oldSlice := unsafe.Slice((*byte)(unsafe.Pointer(oldPtr)), 8)
	oldLenSlice := unsafe.Slice((*byte)(unsafe.Pointer(oldLenPtr)), 8)

	memoryPageSize := uint32(MemoryPageSize)
	binary.LittleEndian.PutUint32(oldSlice, memoryPageSize)
	binary.LittleEndian.PutUint64(oldLenSlice, uint64(4))

	logger.Printf("%-120s %s requested memory page size %s (oldPtr=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sysctl"),
		color.Yellow.Sprintf("0x%X", memoryPageSize),
		color.Yellow.Sprintf("0x%X", oldPtr),
	)
	return 0
}
