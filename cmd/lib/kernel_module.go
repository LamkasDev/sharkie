package lib

import (
	"encoding/binary"
	"path/filepath"
	"strings"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	. "github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// 0x000000000002CD00
// __int64 __fastcall sceKernelGetModuleInfoForUnwind(unsigned __int64, int, _QWORD *, __m128 _XMM0)
func libKernel_sceKernelGetModuleInfoForUnwind(addr uintptr, flags int32, infoPtr uintptr) uintptr {
	if infoPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid info pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelGetModuleInfoForUnwind"),
		)
		return SCE_KERNEL_ERROR_EINVAL
	}

	module := emu.GetModuleAtAddress(addr)
	if module == nil {
		logger.Printf("%-132s %s failed to find module loaded at %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelGetModuleInfoForUnwind"),
			color.Yellow.Sprintf("0x%X", addr),
		)
		return SCE_KERNEL_ERROR_ENOENT
	}
	textSection, _ := emu.GetModuleSections(module)

	moduleInfoForUnwind := (*ModuleInfoForUnwind)(unsafe.Pointer(infoPtr))
	WriteCString((uintptr)(unsafe.Pointer(&moduleInfoForUnwind.Name[0])), module.Name)
	moduleInfoForUnwind.ExceptionFrameHeaderAddress = module.ExceptionFrameSection.Address
	moduleInfoForUnwind.ExceptionFrameAddress = module.ExceptionFrameDataAddress
	moduleInfoForUnwind.ExceptionFrameSize = module.ExceptionFrameDataSize
	moduleInfoForUnwind.TextSectionAddress = textSection.Address
	moduleInfoForUnwind.TextSectionSize = textSection.LoadedSize

	logger.Printf("%-132s %s returned unwind module info for %s (addr=%s, flags=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelGetModuleInfoForUnwind"),
		color.Blue.Sprint(module.Name),
		color.Yellow.Sprintf("0x%X", addr),
		color.Yellow.Sprintf("0x%X", flags),
	)
	return 0
}

// 0x000000000002CFF0
// __int64 __fastcall sceKernelGetExecutableModuleHandle()
func libKernel_sceKernelGetExecutableModuleHandle() uintptr {
	handle := ModuleInfoHandleOffset + uintptr(emu.GlobalModuleManager.CurrentModule.ModuleIndex)

	logger.Printf("%-132s %s returned module handle %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelGetExecutableModuleHandle"),
		color.Yellow.Sprintf("0x%X", handle),
	)
	return handle
}

// 0x000000000002C920
// __int64 __fastcall sceKernelGetModuleInfo(unsigned int, __int64)
func libKernel_sceKernelGetModuleInfo(handle uintptr, infoPtr uintptr) uintptr {
	if infoPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid info pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelGetModuleInfo"),
		)
		return SCE_KERNEL_ERROR_EINVAL
	}

	module := emu.GlobalModuleManager.Modules[handle-ModuleInfoHandleOffset]
	if module == nil {
		logger.Printf("%-132s %s failed due to unknown module %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelGetModuleInfo"),
			color.Yellow.Sprintf("0x%X", handle),
		)
		return SCE_KERNEL_ERROR_ENOENT
	}

	info := (*ModuleInfo)(unsafe.Pointer(infoPtr))
	info.Size = uint64(ModuleInfoSize)
	WriteCString((uintptr)(unsafe.Pointer(&info.Name[0])), module.Name)
	segIndex := uint32(0)
	for _, section := range module.LoadSections {
		if segIndex >= 4 {
			break
		}
		if section.LoadedSize == 0 {
			continue
		}
		info.Segments[segIndex] = SegmentInfo{
			Address:    section.Address,
			Size:       uint32(section.LoadedSize),
			Protection: PROT_READ | PROT_WRITE | PROT_EXEC,
		}
		segIndex++
	}
	info.SegmentsCount = segIndex

	logger.Printf("%-132s %s returned module info for %s (infoPtr=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelGetModuleInfo"),
		color.Blue.Sprint(module.Name),
		color.Yellow.Sprintf("0x%X", infoPtr),
	)
	return 0
}

// TODO: this might be wrong
// 0x0000000000001EB0
// __int64 __fastcall sub_1EB0()
func libKernel_sys_dynlib_get_info_ex(handle uint32, flags uint32, infoPtr uintptr) uintptr {
	if infoPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid info pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sys_dynlib_get_info_ex"),
		)
		return SCE_KERNEL_ERROR_EINVAL
	}
	if handle >= uint32(len(emu.GlobalModuleManager.Modules)) {
		logger.Printf("%-132s %s failed to find module with id %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sys_dynlib_get_info_ex"),
			color.Green.Sprint(handle),
		)
		return SCE_KERNEL_ERROR_ENOENT
	}

	module := emu.GlobalModuleManager.Modules[handle]
	textSection, dataSection := emu.GetModuleSections(module)
	infoSlice := unsafe.Slice((*byte)(unsafe.Pointer(infoPtr)), 352)
	for i := range infoSlice {
		infoSlice[i] = 0
	}
	binary.LittleEndian.PutUint32(infoSlice[0x8:], uint32(module.ModuleIndex))
	binary.LittleEndian.PutUint32(infoSlice[0xC:], 0)
	WriteCString(infoPtr+0x10, module.Name)
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

// 0x0000000000001D90
// __int64 __fastcall sub_1D90()
func libKernel_sys_dynlib_process_needed_and_relocate() uintptr {
	return 0
}

// 0x0000000000016BE0
// __int64 sceKernelIsInSandbox()
func libKernel_sceKernelIsInSandbox() uintptr {
	logger.Printf("%-132s %s returning false.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelIsInSandbox"),
	)
	return 0
}

// 0x000000000001A920
// __int64 sceKernelGetCompiledSdkVersion()
func libKernel_sceKernelGetCompiledSdkVersion(versionPtr uintptr) uintptr {
	sdkVersion := GameCompiledSdkVersion
	if versionPtr != 0 {
		versionSlice := unsafe.Slice((*byte)(unsafe.Pointer(versionPtr)), 4)
		binary.LittleEndian.PutUint32(versionSlice, sdkVersion)
	}

	logger.Printf("%-132s %s returning %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelGetCompiledSdkVersion"),
		color.Yellow.Sprintf("0x%X", sdkVersion),
	)
	return 0
}

// 0x000000000002C370
// void sceKernelLoadStartModuleForSysmodule()
func libKernel_sceKernelLoadStartModuleForSysmodule(namePtr uintptr, argc uintptr, argvPtr uintptr, flags uintptr, optionPtr uintptr, statusPtr uintptr) uintptr {
	return libKernel_sys_sceKernelLoadStartModule(namePtr, argc, argvPtr, flags, optionPtr, statusPtr)
}

// 0x000000000002BB00
// __int64 __fastcall sceKernelLoadStartModule(__int64, __int64, __int64, int, __int64, int *, __m128, __m128, __m128, __m128, double, double, __m128, __m128)
func libKernel_sceKernelLoadStartModule(namePtr uintptr, argc uintptr, argvPtr uintptr, flags uintptr, optionPtr uintptr, statusPtr uintptr) uintptr {
	// TODO: this does a check, but not sure about the signature
	return libKernel_sys_sceKernelLoadStartModule(namePtr, argc, argvPtr, flags, optionPtr, statusPtr)
}

func libKernel_sys_sceKernelLoadStartModule(namePtr uintptr, argc uintptr, argvPtr uintptr, flags uintptr, optionPtr uintptr, resultPtr uintptr) uintptr {
	if namePtr == 0 {
		logger.Printf("%-132s %s failed due to invalid name pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelLoadStartModule"),
		)
		return SCE_KERNEL_ERROR_EINVAL
	}

	name := filepath.Base(ReadCString(namePtr))
	name = strings.ReplaceAll(name, ".prx", ".sprx")
	emu.GlobalModuleManager.ModulesLock.RLock()
	if emu.GlobalModuleManager.ModulesMap[name] != nil {
		logger.Printf("%-132s %s skipping already loaded module %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelLoadStartModule"),
			color.Blue.Sprint(name),
		)
		emu.GlobalModuleManager.ModulesLock.RUnlock()
		return 0
	}
	emu.GlobalModuleManager.ModulesLock.RUnlock()

	if err := emu.GlobalModuleManager.LoadModule(name); err != nil {
		logger.Printf("%-132s %s failed loading module %s: %+v\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelLoadStartModule"),
			color.Blue.Sprint(name),
			err.Error(),
		)
		return 0
	}
	emu.GlobalModuleManager.ModulesLock.RLock()
	defer emu.GlobalModuleManager.ModulesLock.RUnlock()

	module := emu.GlobalModuleManager.ModulesMap[name]
	handle := ModuleInfoHandleOffset + uintptr(module.ModuleIndex)
	if resultPtr != 0 {
		WriteResult(resultPtr, 0)
	}

	logger.Printf("%-132s %s loaded module %s (name=%s).\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelLoadStartModule"),
		color.Yellow.Sprintf("0x%X", handle),
		color.Blue.Sprint(name),
	)
	return handle
}
