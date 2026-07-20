package emu

import (
	"fmt"
	"path/filepath"
	"runtime"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/asm"
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/lib_structs"
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

// SprintStackTrace prints stack trace using thread context.
func SprintStackTrace() string {
	threadContext := asm.GetCurrentThreadContext()
	return "Stack trace:\n" + SprintStackTraceFromSP(threadContext.PlaystationSP)
}

// SprintStackTraceFromSP prints stack trace starting from a given stack pointer.
func SprintStackTraceFromSP(stackPtr uintptr) (result string) {
	thread := GetCurrentThread()
	if stackPtr <= 0x1000 {
		return result
	}
	stackTop := thread.Stack.Address + lib_structs.StackDefaultSize
	for i := 0; i < 40; i++ {
		if stackPtr >= stackTop {
			break
		}
		address := *(*uint64)(unsafe.Pointer(stackPtr))
		result += SprintAddress(uintptr(address))

		stackPtr += 8
	}

	return result
}
