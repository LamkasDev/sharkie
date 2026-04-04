// Package emu handles high-level emulation setup.
package emu

import (
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/asm"
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/structs"
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

// GetCallSiteText returns text indicating the return address call site.
func (m *ModuleManager) GetCallSiteText() string {
	thread := GetCurrentThread()
	return fmt.Sprintf(
		"[%s] [%s]",
		color.Green.Sprint(thread.Name),
		m.GetCallSiteTextShort(),
	)
}

// GetCallSiteTextShort returns short-text indicating the return address call site.
func (m *ModuleManager) GetCallSiteTextShort() string {
	threadContext := asm.GetCurrentThreadContext()

	ctx := (*asm.RegContext)(unsafe.Pointer(threadContext.GlobalStubContext))
	returnAddrPtr := uintptr(unsafe.Pointer(ctx)) + asm.RegContextSize
	returnAddr := *(*uintptr)(unsafe.Pointer(returnAddrPtr))
	module := GetModuleAtAddress(returnAddr)
	if module == nil {
		return fmt.Sprintf("unknown address %s", color.Yellow.Sprintf("0x%X", returnAddr))
	}

	var location string
	callerAddress := GetRealCallerAddress(module, returnAddr)
	if symbolInfo, ok := module.CallerToFunctionName[callerAddress-module.BaseAddress]; !ok {
		stubInfo, ok := asm.StubsTrampolineMap[callerAddress]
		if !ok {
			return fmt.Sprintf(
				"%s+%s/%s",
				color.Blue.Sprint(module.Name),
				color.Magenta.Sprint("unknown function"),
				color.Yellow.Sprintf("0x%X", returnAddr-module.BaseAddress),
			)
		}
		location = color.Magenta.Sprintf("%s:%s", stubInfo.LibraryName, stubInfo.SymbolName)
	} else {
		location = color.Magenta.Sprintf("%s:%s", symbolInfo.LibraryName, symbolInfo.ReadableName)
	}

	return fmt.Sprintf(
		"%s+%s/%s",
		color.Blue.Sprint(module.Name),
		color.Yellow.Sprintf("0x%X", callerAddress-module.BaseAddress),
		location,
	)
}

// GetStackTraceShort returns a trace by walking the stack starting from
// the return address immediately following the GlobalStubContext.
func (m *ModuleManager) GetStackTraceShort() string {
	threadContext := asm.GetCurrentThreadContext()
	if threadContext == nil || threadContext.GlobalStubContext == 0 {
		return "Unable to capture thread context"
	}

	// Skip the RegContext to find the first return address on the stack
	// This is the same logic as your GetCallSiteTextShort
	stackPtr := uintptr(threadContext.GlobalStubContext) + asm.RegContextSize

	// Get stack boundaries to prevent out-of-bounds reads
	thread := GetCurrentThread()
	stackTop := thread.Stack.Address + structs.StackDefaultSize
	stackBottom := thread.Stack.Address

	var sb strings.Builder
	sb.WriteString("Guest Stack Trace (from stub):\n")

	// Walk the stack (limit to 15 frames to keep it "short")
	for i := 0; i < 128; i++ {
		// Safety check: is the stack pointer still within the allocated stack?
		if stackPtr < stackBottom || stackPtr >= stackTop {
			sb.WriteString(color.Red.Sprint("  [End of Stack]\n"))
			break
		}

		// Read the address at the current stack location
		addr := *(*uintptr)(unsafe.Pointer(stackPtr))

		// Format and append the frame
		frameText := m.formatAddressForTrace(addr)
		sb.WriteString(fmt.Sprintf("  frame %d: %s\n", i, frameText))

		// Move to the next 8-byte slot on the stack
		stackPtr += 8
	}

	return sb.String()
}

// formatAddressForTrace handles the module/symbol resolution logic
func (m *ModuleManager) formatAddressForTrace(addr uintptr) string {
	module := GetModuleAtAddress(addr)
	if module == nil {
		return color.Yellow.Sprintf("0x%X", addr)
	}

	// Determine the caller address (subtracting 1 or using logic from GetRealCallerAddress)
	callerAddress := GetRealCallerAddress(module, addr)
	relativeAddr := callerAddress - module.BaseAddress

	var location string
	if symbolInfo, ok := module.CallerToFunctionName[relativeAddr]; ok {
		location = color.Magenta.Sprintf("%s:%s", symbolInfo.LibraryName, symbolInfo.ReadableName)
	} else if stubInfo, ok := asm.StubsTrampolineMap[callerAddress]; ok {
		location = color.Magenta.Sprintf("%s:%s", stubInfo.LibraryName, stubInfo.SymbolName)
	} else {
		location = color.Magenta.Sprintf("unknown function (rel 0x%X)", addr-module.BaseAddress)
	}

	return fmt.Sprintf(
		"%s+%s/%s",
		color.Blue.Sprint(module.Name),
		color.Yellow.Sprintf("0x%X", relativeAddr),
		location,
	)
}
