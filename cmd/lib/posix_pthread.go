package lib

import (
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

const MainThreadGlobalOffset = 0x8E430
const PidGlobalOffset = 0x8E580
const PageSizeGlobalOffset = 0x8E450
const PageSizeGlobalOffset64 = 0x8E448
const InitFlagOffset = 0x8DF78
const SmpFlagOffset = 0x8DEB0

var MainThreadInitialized = false

// 0x000000000000B530
// unsigned __int64 pthread_self()
func libKernel_pthread_self() uintptr {
	if !MainThreadInitialized {
		libKernel_sys_pthread_self()
	}

	return emu.GetCurrentThread()
}

func libKernel_sys_pthread_self() {
	mainThread := emu.GetMainThread()
	base := emu.GlobalModuleManager.ModulesMap["libkernel.sprx"].BaseAddress

	mainThreadSlice := unsafe.Slice((*byte)(unsafe.Pointer(base+MainThreadGlobalOffset)), 8)
	binary.LittleEndian.PutUint64(mainThreadSlice, uint64(mainThread))

	pidSlice := unsafe.Slice((*byte)(unsafe.Pointer(base+PidGlobalOffset)), 4)
	binary.LittleEndian.PutUint32(pidSlice, uint32(libKernel_getpid()))

	pageSizeSlice := unsafe.Slice((*byte)(unsafe.Pointer(base+PageSizeGlobalOffset)), 4)
	binary.LittleEndian.PutUint32(pageSizeSlice, uint32(MemoryPageSize))

	pageSize64Slice := unsafe.Slice((*byte)(unsafe.Pointer(base+PageSizeGlobalOffset64)), 8)
	binary.LittleEndian.PutUint64(pageSize64Slice, uint64(MemoryPageSize))

	initFlagSlice := unsafe.Slice((*byte)(unsafe.Pointer(base+InitFlagOffset)), 1)
	initFlagSlice[0] = 1
	smpFlagSlice := unsafe.Slice((*byte)(unsafe.Pointer(base+SmpFlagOffset)), 4)
	binary.LittleEndian.PutUint32(smpFlagSlice, 1)

	MainThreadInitialized = true
	logger.Printf("%-120s %s initialized thread %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("pthread_self"),
		color.Yellow.Sprintf("0x%X", mainThread),
	)
}
