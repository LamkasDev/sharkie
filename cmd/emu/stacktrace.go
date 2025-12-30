package emu

import (
	"github.com/LamkasDev/sharkie/cmd/asm"
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

type StackTraceFrame struct {
	Address uintptr
	Module  *elf.Elf
}

type StackTrace struct {
	Frames []StackTraceFrame
}

// PrintAddress prints an address and relative position to a module, if within one.
func PrintAddress(address uintptr) {
	hashIndex, ok := asm.StubsMap[address]
	if ok {
		logger.Printf(
			"  %42s (%s)\n",
			color.Blue.Sprintf("%s:%s", asm.Stubs[hashIndex].LibraryName, asm.Stubs[hashIndex].SymbolName),
			color.Yellow.Sprintf("0x%X", address),
		)
	}

	module := GetModuleAtAddress(address)
	if module != nil {
		logger.Printf(
			"  %42s (relative %s)\n",
			color.Blue.Sprint(module.Name),
			color.Yellow.Sprintf("0x%X", address-module.BaseAddress),
		)
	} else {
		logger.Printf(
			"  %42s\n",
			color.Yellow.Sprintf("0x%X", address),
		)
	}
}
