package emu

import (
	"fmt"
	"os"
	"path"
	"sync"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/asm"
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

var GlobalModuleManager = NewModuleManager(
	[]string{"fs", path.Join("fs", "lib")},
)

// ModuleManager keeps track of loaded modules.
type ModuleManager struct {
	LinkPaths []string

	CurrentModule *elf.Elf
	Modules       []*elf.Elf
	ModulesMap    map[string]*elf.Elf
	ModulesLock   sync.RWMutex

	MainThread *Thread
}

// NewModuleManager creates a new instance of ModuleManager.
func NewModuleManager(linkPaths []string) *ModuleManager {
	mm := &ModuleManager{
		LinkPaths:   linkPaths,
		Modules:     make([]*elf.Elf, 1),
		ModulesMap:  map[string]*elf.Elf{},
		ModulesLock: sync.RWMutex{},
	}

	return mm
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

// GetModuleAtAddress returns module that is loaded inside given address.
func GetModuleAtAddress(address uintptr) *elf.Elf {
	GlobalModuleManager.ModulesLock.RLock()
	for _, module := range GlobalModuleManager.Modules {
		if module == nil {
			continue
		}
		for _, section := range module.LoadSections {
			if address >= section.Address && address < section.Address+uintptr(section.LoadedSize) {
				GlobalModuleManager.ModulesLock.RUnlock()
				return module
			}
		}
	}
	GlobalModuleManager.ModulesLock.RUnlock()

	return nil
}

// GetModuleSections returns the TEXT and DATA sections for a given module.
func GetModuleSections(module *elf.Elf) (*elf.ElfLoadSection, *elf.ElfLoadSection) {
	var textSection, dataSection *elf.ElfLoadSection
	for _, section := range module.LoadSections {
		if textSection == nil && (section.PFlags&elf.PF_X) != 0 {
			textSection = section
		}
		if dataSection == nil && (section.PFlags&elf.PF_W) != 0 {
			dataSection = section
		}
	}
	if textSection == nil && len(module.LoadSections) > 0 {
		logger.Printf("%-132s %s failed to find TEXT section.\n",
			GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("GetModuleSections"),
		)
	}
	if dataSection == nil {
		logger.Printf("%-132s %s failed to find DATA section.\n",
			GlobalModuleManager.GetCallSiteText(),
			color.Magenta.Sprint("GetModuleSections"),
		)
		dataSection = textSection
	}

	return textSection, dataSection
}

// GetRealCallerAddress return the real caller address for a return address, bypassing any stubs.
func GetRealCallerAddress(e *elf.Elf, returnAddr uintptr) uintptr {
	base := e.BaseAddress
	if returnAddr < base || returnAddr-base >= uintptr(len(e.Memory)) {
		return 0
	}

	if e.Memory[(returnAddr-5)-base] == 0xE8 {
		// Check for Direct Call (E8) - PLT Stub
		callInstAddr := returnAddr - 5

		// Resolve target of the CALL (PLT Stub)
		rel32 := e.ReadInt32(callInstAddr + 1) // Skip opcode E8
		pltStubAddr := returnAddr + uintptr(rel32)
		pltOffset := pltStubAddr - base

		// HACK: this is to handle our own's module patch inside linker.go
		if e.Memory[pltOffset] == 0x48 && e.Memory[pltOffset+1] == 0xB8 &&
			e.Memory[pltOffset+10] == 0xFF && e.Memory[pltOffset+11] == 0xE0 {
			return uintptr(e.ReadInt64(pltStubAddr + 2))
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
		return pltStubAddr + 6 + uintptr(disp32)
	} else if e.Memory[(returnAddr-6)-base] == 0xFF &&
		e.Memory[(returnAddr-5)-base] == 0x15 {
		// Check for Indirect Call (FF 15) - Direct GOT
		callInstAddr := returnAddr - 6

		disp32 := e.ReadInt32(callInstAddr + 2)
		return returnAddr + uintptr(disp32)
	}

	return 0
}

// GetCallSiteText returns text indicating the returnAddr call site.
func (m *ModuleManager) GetCallSiteText() string {
	thread := GetCurrentThread()
	threadContext := asm.GetCurrentThreadContext()

	ctx := (*asm.RegContext)(unsafe.Pointer(threadContext.GlobalStubContext))
	returnAddrPtr := uintptr(unsafe.Pointer(ctx)) + asm.RegContextSize
	returnAddr := *(*uintptr)(unsafe.Pointer(returnAddrPtr))
	module := GetModuleAtAddress(returnAddr)
	if module == nil {
		return fmt.Sprintf("[%s] [unknown address %s]",
			color.Green.Sprint(thread.Name),
			color.Yellow.Sprintf("0x%X", returnAddr),
		)
	}

	callerAddress := GetRealCallerAddress(module, returnAddr)
	hashIndex, ok := module.CallerToFunctionName[callerAddress-module.BaseAddress]
	if !ok {
		hashIndex, ok = asm.StubsTrampolineMap[callerAddress]
		if !ok {
			return fmt.Sprintf(
				"[%s] [%s+%s/%s]",
				color.Green.Sprint(thread.Name),
				color.Blue.Sprint(module.Name),
				color.Magenta.Sprint("unknown function"),
				color.Yellow.Sprintf("0x%X", returnAddr-module.BaseAddress),
			)
		}
	}

	symbol := module.SymbolTable.SymbolsMap[hashIndex]
	return fmt.Sprintf(
		"[%s] [%s+%s/%s]",
		color.Green.Sprint(thread.Name),
		color.Blue.Sprint(module.Name),
		color.Yellow.Sprintf("0x%X", callerAddress-module.BaseAddress),
		color.Magenta.Sprintf("%s:%s", symbol.LibraryName, symbol.ReadableName),
	)
}
