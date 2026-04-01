package emu

import (
	"fmt"
	"path/filepath"
	"runtime"

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

// SprintAddress prints an address and relative position to a module, if within one.
func SprintAddress(address uintptr) string {
	stubInfo, ok := asm.StubsMap[address]
	if ok {
		return fmt.Sprintf(
			"  %42s (%s)\n",
			color.Blue.Sprintf("%s:%s", stubInfo.LibraryName, stubInfo.SymbolName),
			color.Yellow.Sprintf("0x%X", address),
		)
	}

	module := GetModuleAtAddress(address)
	if module != nil {
		return fmt.Sprintf(
			"  %42s (relative %s)\n",
			color.Blue.Sprint(module.Name),
			color.Yellow.Sprintf("0x%X", address-module.BaseAddress),
		)
	}

	if fn := runtime.FuncForPC(address); fn != nil {
		file, line := fn.FileLine(address)
		return fmt.Sprintf(
			"  %42s (%s:%d)\n",
			color.Magenta.Sprint(filepath.Base(fn.Name())),
			filepath.Base(file),
			line,
		)
	}

	return fmt.Sprintf(
		"  %42s\n",
		color.Yellow.Sprintf("0x%X", address),
	)
}
