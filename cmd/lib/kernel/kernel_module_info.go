package kernel

import (
	"github.com/LamkasDev/sharkie/cmd/elf"
	. "github.com/LamkasDev/sharkie/cmd/elf_symbol"
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
		return SCE_KERNEL_ERROR_EFAULT
	}

	module := emu.GetModuleAtAddress(addr)
	if module == nil {
		logger.Printf("%-132s %s failed to find module loaded at %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelGetModuleInfoForUnwind"),
			color.Yellow.Sprintf("0x%X", addr),
		)
		return SCE_KERNEL_ERROR_ESRCH
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
		return SCE_KERNEL_ERROR_EFAULT
	}

	emu.GlobalModuleManager.ModulesLock.RLock()
	module := emu.GlobalModuleManager.Modules[handle]
	emu.GlobalModuleManager.ModulesLock.RUnlock()
	if module == nil {
		logger.Printf("%-132s %s failed due to unknown module %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelGetModuleInfo"),
			color.Yellow.Sprintf("0x%X", handle),
		)
		return SCE_KERNEL_ERROR_ESRCH
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

// 0x000000000002C6F0
// __int64 __fastcall sceKernelDlsym(unsigned int, __int64, __int64)
func libKernel_sceKernelDlsym(handle ModuleHandle, symbolNamePtr Cstring, addressPtr uintptr) uintptr {
	if symbolNamePtr == nil || addressPtr == 0 {
		logger.Printf("%-132s %s failed due to invalid pointers.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelDlsym"),
		)
		return SCE_KERNEL_ERROR_EFAULT
	}

	emu.GlobalModuleManager.ModulesLock.RLock()
	defer emu.GlobalModuleManager.ModulesLock.RUnlock() // GetSymbolAddress needs lock.
	module := emu.GlobalModuleManager.Modules[handle]
	if module == nil {
		logger.Printf("%-132s %s failed due to unknown module %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelDlsym"),
			color.Yellow.Sprintf("0x%X", handle),
		)
		return SCE_KERNEL_ERROR_ESRCH
	}
	symbolName := GoString(symbolNamePtr)
	mangledSymbolName := ReadableToMangled(symbolName)
	logger.Printf("%-132s %s resolving symbol %s (%s) in module %s.\n",
		emu.GlobalModuleManager.GetCallSiteText(),
		color.Magenta.Sprint("sceKernelDlsym"),
		color.Blue.Sprint(symbolName),
		color.Blue.Sprint(mangledSymbolName),
		color.Blue.Sprint(module.Name),
	)

	// Search for the symbol in the module's symbol table.
	var foundAddress uintptr
	var found bool
	for _, symbol := range module.SymbolTable.Symbols {
		if symbol.ReadableName == symbolName || symbol.NidBase == mangledSymbolName {
			if address, ok := elf.GetSymbolAddress(symbol); ok {
				foundAddress = address
				found = true
				break
			}
		}
	}
	if !found {
		logger.Printf("%-132s %s failed to find symbol %s.\n",
			emu.GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("sceKernelDlsym"),
			color.Blue.Sprint(symbolName),
		)
		return SCE_KERNEL_ERROR_ESRCH
	}
	WriteAddress(addressPtr, foundAddress)

	return 0
}
