package kernel

import (
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x0000000000001D90
// __int64 __fastcall sub_1D90()
func libKernel_sys_dynlib_process_needed_and_relocate() uintptr {
	return 0
}

// TODO: this might be wrong
// 0x0000000000001EB0
// __int64 __fastcall sub_1EB0()
func libKernel_sys_dynlib_get_info_ex(handle, flags, infoPtr uintptr) uintptr {
	if infoPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid info pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sys_dynlib_get_info_ex"),
		)
		return SCE_KERNEL_ERROR_EINVAL
	}
	if handle >= uintptr(len(emu.GlobalModuleManager.Modules)) {
		logger.Printf("%-132s %s failed to find module with id %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sys_dynlib_get_info_ex"),
			color.Green.Sprint(handle),
		)
		return SCE_KERNEL_ERROR_ENOENT
	}

	emu.GlobalModuleManager.ModulesLock.RLock()
	module := emu.GlobalModuleManager.Modules[handle]
	emu.GlobalModuleManager.ModulesLock.RUnlock()
	textSection, dataSection := emu.GetModuleSections(module)
	infoSlice := unsafe.Slice((*byte)(unsafe.Pointer(infoPtr)), 352)
	for i := range infoSlice {
		infoSlice[i] = 0
	}
	binary.LittleEndian.PutUint32(infoSlice[0x8:], uint32(module.ModuleIndex))
	binary.LittleEndian.PutUint32(infoSlice[0xC:], 0)
	CString(Cstring(unsafe.Add(unsafe.Pointer(infoPtr), 0x10)), module.Name)
	binary.LittleEndian.PutUint64(infoSlice[0x110:], uint64(textSection.Address))
	binary.LittleEndian.PutUint32(infoSlice[0x118:], uint32(textSection.LoadedSize))
	binary.LittleEndian.PutUint64(infoSlice[0x11C:], uint64(dataSection.Address))
	binary.LittleEndian.PutUint32(infoSlice[0x124:], uint32(dataSection.LoadedSize))
	if module.ExceptionFrameSection != nil {
		binary.LittleEndian.PutUint64(infoSlice[0x128:], uint64(module.ExceptionFrameDataAddress))
		binary.LittleEndian.PutUint32(infoSlice[0x130:], uint32(module.ExceptionFrameSection.LoadedSize))
	} else {
		binary.LittleEndian.PutUint64(infoSlice[0x128:], 0)
		binary.LittleEndian.PutUint32(infoSlice[0x130:], 0)
	}

	logger.Printf("%-132s %s returned module info for %s (handle=%s, flags=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sys_dynlib_get_info_ex"),
		color.Blue.Sprint(module.Name),
		color.Green.Sprint(handle),
		color.Yellow.Sprintf("0x%X", flags),
	)
	return 0
}
