package kernel

import (
	"github.com/LamkasDev/sharkie/cmd/emu"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/module"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// 0x000000000002CD00
// __int64 __fastcall sceKernelGetModuleInfoForUnwind(unsigned __int64, int, _QWORD *, __m128 _XMM0)
func libKernel_sceKernelGetModuleInfoForUnwind(addr, flags uintptr, moduleInfoForUnwind *ModuleInfoForUnwind) uintptr {
	if moduleInfoForUnwind == nil {
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

	CString(Cstring(&moduleInfoForUnwind.Name[0]), module.Name)
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

// 0x000000000002C920
// __int64 __fastcall sceKernelGetModuleInfo(unsigned int, __int64)
func libKernel_sceKernelGetModuleInfo(handle ModuleHandle, info *ModuleInfo) uintptr {
	if info == nil {
		logger.Printf("%-132s %s failed due to invalid info pointer.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelGetModuleInfo"),
		)
		return SCE_KERNEL_ERROR_EINVAL
	}

	module := emu.GlobalModuleManager.Modules[handle]
	if module == nil {
		logger.Printf("%-132s %s failed due to unknown module %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelGetModuleInfo"),
			color.Yellow.Sprintf("0x%X", handle),
		)
		return SCE_KERNEL_ERROR_ENOENT
	}

	info.Size = uint64(ModuleInfoSize)
	CString(Cstring(&info.Name[0]), module.Name)
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

	logger.Printf("%-132s %s returned module info for %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelGetModuleInfo"),
		color.Blue.Sprint(module.Name),
	)
	return 0
}
