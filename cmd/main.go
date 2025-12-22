package main

import (
	"fmt"
	"runtime"

	"github.com/LamkasDev/sharkie/cmd/asm"
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/lib"
	"github.com/LamkasDev/sharkie/cmd/symbol"
	"github.com/gookit/color"
)

func main() {
	// Lock the goroutine to its current OS thread.
	// This is crucial because we are manipulating the mem and setting up
	// a thread-local exception handler.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	emu.StartProfiling()

	fmt.Printf("hi from %s :3\n", color.Blue.Sprint("sharkie"))
	asm.ExceptionHandler = emu.ExceptionHandlerGo
	elf.GetSymbolAddress = emu.GetSymbolAddress
	elf.GetDefiningModule = emu.GetDefiningModule
	asm.InitSignalsAddr()
	asm.InitStubAddr()
	asm.SetupCooperativeGC()
	emu.SetupSignalHandler()

	symbol.LoadSymbolMap("data/aerolib.csv")
	lib.RegisterStubs()

	emu.GlobalModuleManager.LoadModule("eboot.bin")
	emu.GlobalModuleManager.RunModule("eboot.bin")
	emu.StopProfiling()
}
