package emu

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/asm"
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/linker"
	"github.com/LamkasDev/sharkie/cmd/patcher"
	"github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

var GlobalModuleManager = NewModuleManager(
	[]string{"fs", path.Join("fs", "lib")},
)

// ModuleManager keeps track of loaded modules.
type ModuleManager struct {
	LinkPaths       []string
	CurrentModule   *elf.Elf
	NextModuleIndex uint64
	Modules         map[string]*elf.Elf

	Stack *structs.Stack
	Tcb   *structs.Tcb
}

// NewModuleManager creates a new instance of ModuleManager.
func NewModuleManager(linkPaths []string) *ModuleManager {
	mm := &ModuleManager{
		LinkPaths:       linkPaths,
		NextModuleIndex: 1,
		Modules:         map[string]*elf.Elf{},
	}

	return mm
}

// LoadModule loads & links module specified by name.
func (m *ModuleManager) LoadModule(name string) {
	// Only load the modules.
	m._RecursiveLoadModule(name)

	// Link & patch everything now.
	for _, module := range m.Modules {
		if !module.Linked {
			fmt.Printf(
				"\nLinking module %s from %s...\n",
				color.Blue.Sprint(module.Name),
				color.Blue.Sprint(module.Path),
			)
			linker.GlobalLinker.Link(module)
			patcher.GlobalPatcher.Patch(module)
			module.Linked = true
		}
	}
}

func GetBaseModuleName(name string) string {
	name = strings.ReplaceAll(name, ".prx", "")
	name = strings.ReplaceAll(name, ".sprx", "")
	return name
}

// _RecursiveLoadModule loads a module and dependencies without linking.
func (m *ModuleManager) _RecursiveLoadModule(name string) {
	if m.Modules[name] != nil {
		return
	}

	modulePath := m.GetModulePath(name)
	if modulePath == nil {
		log.Panicf("Could not find module %s!\n", name)
	}

	moduleIndex := m.NextModuleIndex
	fmt.Printf(
		"\nLoading module %s from %s...\n",
		color.Green.Sprint(moduleIndex),
		color.Blue.Sprint(*modulePath),
	)
	data, err := os.ReadFile(*modulePath)
	if err != nil {
		panic(err)
	}

	module := elf.NewElf(data)
	module.ModuleIndex = moduleIndex
	module.Path = *modulePath
	m.Modules[name] = module
	m.NextModuleIndex++

	for _, needed := range module.DynamicInfo.Needed {
		needed = strings.ReplaceAll(needed, ".prx", ".sprx")
		if needed == "libSceGnmDriver_padebug.sprx" ||
			needed == "libSceDbgAddressSanitizer.sprx" ||
			needed == "libSceDipsw.sprx" {
			continue
		}
		m._RecursiveLoadModule(needed)
	}
}

// RunModuleInitializers recursively executes init functions of modules.
func (m *ModuleManager) RunModuleInitializers(module *elf.Elf, visited map[string]bool, skipOwnInit bool) {
	if visited[module.Name] {
		return
	}
	visited[module.Name] = true

	for _, needed := range module.DynamicInfo.Needed {
		needed = strings.ReplaceAll(needed, ".prx", ".sprx")
		if needed == "libSceGnmDriver_padebug.sprx" ||
			needed == "libSceDbgAddressSanitizer.sprx" ||
			needed == "libSceDipsw.sprx" {
			continue
		}
		if dependency := m.Modules[needed]; dependency != nil {
			m.RunModuleInitializers(dependency, visited, false)
		}
	}

	isSelfContained := module.Name == "libSceLibcInternal.sprx"
	if skipOwnInit {
		return
	}

	// Call initialization functions.
	if !isSelfContained {
		for _, funcAddr := range module.DynamicInfo.PreInitArray {
			fmt.Printf(
				"Calling %s's %s function at %s...\n",
				color.Blue.Sprint(module.Name),
				color.Magenta.Sprint("DT_PREINIT_ARRAY"),
				color.Yellow.Sprintf("0x%X", funcAddr),
			)
			m.Call(uintptr(funcAddr))
		}
	}
	if module.DynamicInfo.InitFunc != nil {
		fmt.Printf(
			"Calling %s's %s function at %s...\n",
			color.Blue.Sprint(module.Name),
			color.Magenta.Sprint("DT_INIT"),
			color.Yellow.Sprintf("0x%X", module.DynamicInfo.InitFunc),
		)
		m.Call(uintptr(*module.DynamicInfo.InitFunc))
	}
	if !isSelfContained {
		for _, funcAddr := range module.DynamicInfo.InitArray {
			fmt.Printf(
				"Calling %s's %s function at %s...\n",
				color.Blue.Sprint(module.Name),
				color.Magenta.Sprint("DT_INIT_ARRAY"),
				color.Yellow.Sprintf("0x%X", funcAddr),
			)
			m.Call(uintptr(funcAddr))
		}
	}
}

// RunModule runs module specified by name.
func (m *ModuleManager) RunModule(name string) {
	m.CurrentModule = m.Modules[name]
	if m.CurrentModule == nil {
		log.Panicf("Module %s is not loaded!\n", name)
	}

	fmt.Printf(
		"\nRunning module %s...\n",
		color.Blue.Sprint(name),
	)
	m.Prepare(linker.GlobalLinker)
	visited := make(map[string]bool)
	m.RunModuleInitializers(m.CurrentModule, visited, true)
	m.Run(m.CurrentModule)
}

// GetModulePath returns the first valid path for a module name.
func (m *ModuleManager) GetModulePath(name string) *string {
	for _, linkPath := range m.LinkPaths {
		modulePath := path.Join(linkPath, name)
		if _, err := os.Stat(modulePath); err == nil {
			return &modulePath
		}
	}

	return nil
}

// GetSymbolAddress returns the symbol address for given symbol.
func GetSymbolAddress(s *elf.ElfSymbol) (uint64, bool) {
	if stub, ok := asm.Stubs[s.HashIndex]; ok {
		/* fmt.Printf(
			"Found stubbed symbol %s at %s.\n",
			color.Blue.Sprint(fullName),
			color.Yellow.Sprintf("0x%X", stub.Address),
		) */
		return uint64(stub.Address), true
	}

	// Let's use a generic stub for now, so we know which functions to patch.
	if s.LibraryName == "libkernel" && s.Type == elf.STT_FUNC &&
		s.ReadableName != "scePthreadSelf" {
		return uint64(asm.Stubs[elf.GetSymbolHashIndex("", "__sharkie_generic_stub")].Address), true
	}

	if s.Type == elf.STT_OBJECT {
		// TODO: add more priorities?
		if module, ok := GlobalModuleManager.Modules["libSceLibcInternal.sprx"]; ok {
			if address, ok := TryGetSymbolAddress(s, module); ok {
				return address, true
			}
		}
	}

	// libSceVideoOut:sceVideoOutSubmitEopFlip is at 0x0
	// libSceVideoOut:sceVideoOutGetBufferLabelAddress is at 0x0
	for _, module := range GlobalModuleManager.Modules {
		if address, ok := TryGetSymbolAddress(s, module); ok {
			return address, true
		}
	}
	// fmt.Printf("Failed search for symbol %s.\n", color.Red.Sprint(fullName))

	return 0, false
}

// GetDefiningModule returns the module that actually defines given symbol.
func GetDefiningModule(s *elf.ElfSymbol) *elf.Elf {
	if s.LibraryName != "" {
		if module, ok := GlobalModuleManager.Modules[s.LibraryName]; ok {
			return module
		}

		return nil
	}

	for _, module := range GlobalModuleManager.Modules {
		if _, found := TryGetSymbolAddress(s, module); found {
			return module
		}
	}

	return nil
}

// TryGetSymbolAddress tries returning the symbol address for given symbol from passed module.
func TryGetSymbolAddress(s *elf.ElfSymbol, module *elf.Elf) (uint64, bool) {
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
				color.Gray.Printf(
					"Resolving symbol %s:%s for %s:%s in module %s at 0x%X.\n",
					symbol.LibraryName, symbol.ReadableName,
					s.LibraryName, s.ReadableName,
					module.Name,
					module.BaseAddress+uintptr(symbol.Address),
				)
			}

			/* fmt.Printf(
				"Found symbol %s in module %s at %s.\n",
				color.Blue.Sprintf("%s:%s", symbol.LibraryName, symbol.ReadableName),
				color.Blue.Sprint(module.Name),
				color.Yellow.Sprintf("0x%X", module.BaseAddress+uintptr(symbol.Address)),
			) */
			return uint64(module.BaseAddress) + symbol.Address, true
		}
	}

	return 0, false
}

// GetModuleForInstructionPointer return the module that's loaded within given address.
func (m *ModuleManager) GetModuleForInstructionPointer(address uintptr) *elf.Elf {
	for _, module := range m.Modules {
		if address >= module.BaseAddress && address <= module.BaseAddress+uintptr(len(module.Memory)) {
			return module
		}
	}

	return nil
}

// GetRealCallerAddress return the real caller address for a return address, bypassing any stubs.
func GetRealCallerAddress(e *elf.Elf, returnAddr uint64) uint64 {
	base := uint64(e.BaseAddress)
	if returnAddr < base || returnAddr-base >= uint64(len(e.Memory)) {
		return 0
	}

	if e.Memory[(returnAddr-5)-base] == 0xE8 {
		// Check for Direct Call (E8) - PLT Stub
		callInstAddr := returnAddr - 5

		// Resolve target of the CALL (PLT Stub)
		rel32 := e.ReadInt32(callInstAddr + 1) // Skip opcode E8
		pltStubAddr := returnAddr + uint64(rel32)
		pltOffset := pltStubAddr - base

		// HACK: this is to handle our own's module patch inside linker.go
		if e.Memory[pltOffset] == 0x48 && e.Memory[pltOffset+1] == 0xB8 &&
			e.Memory[pltOffset+10] == 0xFF && e.Memory[pltOffset+11] == 0xE0 {
			return uint64(e.ReadInt64(pltStubAddr + 2))
		}

		// Read the PLT Stub instruction
		// Expecting: JMP [RIP + disp] (FF 25 xx xx xx xx).
		if e.Memory[pltOffset] != 0xFF || e.Memory[pltOffset+1] != 0x25 {
			return 0
		}

		// Resolve GOT Slot
		// Target = RIP (next instruction) + disp
		// RIP for this instruction is pltStubAddr + 6
		disp32 := e.ReadInt32(pltStubAddr + 2) // Skip opcode FF 25
		return pltStubAddr + 6 + uint64(disp32)
	} else if e.Memory[(returnAddr-6)-base] == 0xFF &&
		e.Memory[(returnAddr-5)-base] == 0x15 {
		// Check for Indirect Call (FF 15) - Direct GOT
		callInstAddr := returnAddr - 6

		disp32 := e.ReadInt32(callInstAddr + 2)
		return returnAddr + uint64(disp32)
	}

	return 0
}

// GetCallSiteText returns text indicating the returnAddr call site.
func (m *ModuleManager) GetCallSiteText() string {
	ctx := (*asm.RegContext)(unsafe.Pointer(asm.GlobalStubContext))
	returnAddr := *(*uintptr)(unsafe.Pointer(ctx.BP + 8))
	module := m.GetModuleForInstructionPointer(returnAddr)
	if module == nil {
		return fmt.Sprintf("[unknown address %s]",
			color.Yellow.Sprintf("0x%X", returnAddr),
		)
	}

	callerAddress := GetRealCallerAddress(module, uint64(returnAddr))
	hashIndex, ok := module.CallerToFunctionName[callerAddress-uint64(module.BaseAddress)]
	if !ok {
		hashIndex, ok = asm.StubsTrampolineMap[uintptr(callerAddress)]
		if !ok {
			return fmt.Sprintf(
				"[%s called %s at %s]",
				color.Blue.Sprint(module.Name),
				color.Magenta.Sprint("unknown function"),
				color.Yellow.Sprintf("0x%X", returnAddr),
			)
		}
	}
	symbol := module.SymbolTable.SymbolsMap[hashIndex]

	return fmt.Sprintf(
		"[%s called %s at %s]",
		color.Blue.Sprint(module.Name),
		color.Magenta.Sprintf("%s:%s", symbol.LibraryName, symbol.ReadableName),
		color.Yellow.Sprintf("0x%X", (returnAddr-5)-module.BaseAddress),
	)
}
