package emu

import (
	"github.com/LamkasDev/sharkie/cmd/asm"
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// GetSymbolAddress returns the symbol address for given symbol.
func GetSymbolAddress(s *elf.ElfSymbol) (uintptr, bool) {
	if stub, ok := asm.Stubs[s.HashIndex]; ok {
		/* logger.Printf(
			"Found stubbed symbol %s at %s.\n",
			color.Blue.Sprint(fullName),
			color.Yellow.Sprintf("0x%X", stub.Address),
		) */
		return stub.Address, true
	}

	// Let's use a generic stub for now, so we know which functions to patch.
	if s.LibraryName == "libkernel" && s.Type == elf.STT_FUNC && elf.CanStubFunctionName(s.ReadableName) {
		return asm.Stubs[elf.GetSymbolHashIndex("", "__sharkie_generic_stub")].Address, true
	}

	if s.Type == elf.STT_OBJECT {
		// TODO: add more priorities?
		if module, ok := GlobalModuleManager.ModulesMap["libSceLibcInternal.sprx"]; ok {
			if address, ok := TryGetSymbolAddress(s, module); ok {
				return address, true
			}
		}
	}

	// libSceVideoOut:sceVideoOutSubmitEopFlip is at 0x0
	// libSceVideoOut:sceVideoOutGetBufferLabelAddress is at 0x0
	for _, module := range GlobalModuleManager.Modules {
		if module == nil {
			continue
		}
		if address, ok := TryGetSymbolAddress(s, module); ok {
			return address, true
		}
	}
	// logger.Printf("Failed search for symbol %s.\n", color.Red.Sprint(fullName))

	return 0, false
}

// GetDefiningModule returns the module that actually defines given symbol.
func GetDefiningModule(s *elf.ElfSymbol) *elf.Elf {
	GlobalModuleManager.ModulesLock.RLock()
	defer GlobalModuleManager.ModulesLock.RUnlock()

	if s.LibraryName != "" {
		if module, ok := GlobalModuleManager.ModulesMap[s.LibraryName]; ok {
			return module
		}

		return nil
	}

	for _, module := range GlobalModuleManager.Modules {
		if module == nil {
			continue
		}
		if _, found := TryGetSymbolAddress(s, module); found {
			return module
		}
	}

	return nil
}

// TryGetSymbolAddress tries returning the symbol address for given symbol from passed module.
func TryGetSymbolAddress(s *elf.ElfSymbol, module *elf.Elf) (uintptr, bool) {
	if module.DynamicInfo == nil {
		return 0, false
	}
	for _, exportedLibrary := range module.DynamicInfo.ExportLibraries {
		if s.LibraryName != exportedLibrary.Name {
			continue
		}
		for _, symbol := range module.SymbolTable.Symbols {
			if symbol.Address == 0 {
				continue
			}
			if symbol.ReadableName != s.ReadableName {
				// Let's try skipping the #A#B suffix if they match without it and print warning.
				if len(symbol.OriginalName) > 4 && len(s.OriginalName) > 4 &&
					symbol.OriginalName[:len(symbol.OriginalName)-4] != s.OriginalName[:len(s.OriginalName)-4] {
					continue
				}
				logger.Print(color.Gray.Sprintf(
					"  Resolving symbol %s:%s for %s:%s in module %s at 0x%X.\n",
					symbol.LibraryName,
					symbol.ReadableName,
					s.LibraryName,
					s.ReadableName,
					module.Name,
					module.BaseAddress+symbol.Address,
				))
			}

			/* logger.Printf(
				"Found symbol %s in module %s at %s.\n",
				color.Blue.Sprintf("%s:%s", symbol.LibraryName, symbol.ReadableName),
				color.Blue.Sprint(module.Name),
				color.Yellow.Sprintf("0x%X", module.BaseAddress+uintptr(symbol.Address)),
			) */
			return module.BaseAddress + symbol.Address, true
		}
	}

	return 0, false
}
