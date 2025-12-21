package emu

import (
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/asm"
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/gookit/color"
)

type StackTraceFrame struct {
	Address uintptr
	Module  *elf.Elf
}

type StackTrace struct {
	Frames []StackTraceFrame
}

// SafeReadUint64 safely reads a uint64 value from the stack.
func SafeReadUint64(address uintptr) (uint64, bool) {
	module := GlobalModuleManager.CurrentModule
	if module != nil && GlobalModuleManager.Stack != nil {
		if address >= GlobalModuleManager.Stack.Address && address+8 <= GlobalModuleManager.Stack.Address+uintptr(len(GlobalModuleManager.Stack.Contents)) {
			return *(*uint64)(unsafe.Pointer(address)), true
		}
	}

	return 0, false
}

// PrintAddress prints an address and relative position to a module, if within one.
func PrintAddress(address uintptr) {
	hashIndex, ok := asm.StubsMap[address]
	if ok {
		fmt.Printf(
			"  %42s (%s)\n",
			color.Blue.Sprintf("%s:%s", asm.Stubs[hashIndex].LibraryName, asm.Stubs[hashIndex].SymbolName),
			color.Yellow.Sprintf("0x%X", address),
		)
	}

	module := GlobalModuleManager.GetModuleForInstructionPointer(address)
	if module != nil {
		fmt.Printf(
			"  %42s (relative %s)\n",
			color.Blue.Sprint(module.Name),
			color.Yellow.Sprintf("0x%X", address-module.BaseAddress),
		)
	} else {
		fmt.Printf(
			"  %42s\n",
			color.Yellow.Sprintf("0x%X", address),
		)
	}
}
