package lib

import (
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x000000000002CD00
// __int64 __fastcall sceKernelGetModuleInfoForUnwind(unsigned __int64, int, _QWORD *, __m128 _XMM0)
func libKernel_sceKernelGetModuleInfoForUnwind(addr uintptr, flags int32, infoPtr uintptr) uintptr {
	if infoPtr == 0 {
		fmt.Printf("%-120s %s failed due to invalid info pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelGetModuleInfoForUnwind"),
		)
		return structs.SCE_KERNEL_ERROR_EINVAL
	}

	module := emu.GetModuleAtAddress(addr)
	if module == nil {
		fmt.Printf("%-120s %s failed to find module loaded at %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelGetModuleInfoForUnwind"),
			color.Yellow.Sprintf("0x%X", addr),
		)
		return structs.SCE_KERNEL_ERROR_ENOENT
	}
	textSection, _, exceptionFrameSection := emu.GetModuleSections(module)
	infoSlice := unsafe.Slice((*byte)(unsafe.Pointer(infoPtr)), 304)
	for i := range infoSlice {
		infoSlice[i] = 0
	}
	structs.WriteCString(infoPtr+0x08, module.Name)
	binary.LittleEndian.PutUint64(infoSlice[0x108:], uint64(exceptionFrameSection.Address))
	binary.LittleEndian.PutUint64(infoSlice[0x110:], uint64(module.ExceptionFrameDataAddress))
	binary.LittleEndian.PutUint64(infoSlice[0x118:], module.ExceptionFrameDataSize)
	binary.LittleEndian.PutUint64(infoSlice[0x120:], uint64(textSection.Address))
	binary.LittleEndian.PutUint64(infoSlice[0x128:], textSection.LoadedSize)

	fmt.Printf("%-120s %s returned module info for %s (addr=%s, flags=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelGetModuleInfoForUnwind"),
		color.Blue.Sprint(module.Name),
		color.Yellow.Sprintf("0x%X", addr),
		color.Yellow.Sprintf("0x%X", flags),
	)
	return 0
}

// TODO: this might be wrong
// 0x0000000000001EB0
// __int64 __fastcall sub_1EB0()
func libKernel_sys_dynlib_get_info_ex(handle uint32, flags uint32, infoPtr uintptr) uintptr {
	if infoPtr == 0 {
		fmt.Printf("%-120s %s failed due to invalid info pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sys_dynlib_get_info_ex"),
		)
		return structs.SCE_KERNEL_ERROR_EINVAL
	}
	if handle >= uint32(len(emu.GlobalModuleManager.Modules)) {
		fmt.Printf("%-120s %s failed to find module with id %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sys_dynlib_get_info_ex"),
			color.Green.Sprint(handle),
		)
		return structs.SCE_KERNEL_ERROR_ENOENT
	}

	module := emu.GlobalModuleManager.Modules[handle]
	textSection, dataSection, exceptionFrameSection := emu.GetModuleSections(module)
	infoSlice := unsafe.Slice((*byte)(unsafe.Pointer(infoPtr)), 352)
	for i := range infoSlice {
		infoSlice[i] = 0
	}
	binary.LittleEndian.PutUint32(infoSlice[0x8:], uint32(module.ModuleIndex))
	binary.LittleEndian.PutUint32(infoSlice[0xC:], 0)
	structs.WriteCString(infoPtr+0x10, module.Name)
	binary.LittleEndian.PutUint64(infoSlice[0x110:], uint64(textSection.Address))
	binary.LittleEndian.PutUint32(infoSlice[0x118:], uint32(textSection.LoadedSize))
	binary.LittleEndian.PutUint64(infoSlice[0x11C:], uint64(dataSection.Address))
	binary.LittleEndian.PutUint32(infoSlice[0x124:], uint32(dataSection.LoadedSize))
	if exceptionFrameSection != nil {
		binary.LittleEndian.PutUint64(infoSlice[0x128:], uint64(module.ExceptionFrameDataAddress))
		binary.LittleEndian.PutUint32(infoSlice[0x130:], uint32(exceptionFrameSection.LoadedSize))
	} else {
		binary.LittleEndian.PutUint64(infoSlice[0x128:], 0)
		binary.LittleEndian.PutUint32(infoSlice[0x130:], 0)
	}

	fmt.Printf("%-120s %s returned module info for %s (handle=%s, flags=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sys_dynlib_get_info_ex"),
		color.Blue.Sprint(module.Name),
		color.Green.Sprint(handle),
		color.Yellow.Sprintf("0x%X", flags),
	)
	return 0
}
