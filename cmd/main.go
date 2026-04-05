package main

import (
	"runtime"

	"github.com/LamkasDev/sharkie/cmd/app"
	"github.com/LamkasDev/sharkie/cmd/asm"
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/lib"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/LamkasDev/sharkie/cmd/structs/dce"
	"github.com/LamkasDev/sharkie/cmd/structs/fs"
	"github.com/LamkasDev/sharkie/cmd/structs/gc"
	"github.com/LamkasDev/sharkie/cmd/structs/gpu"
	"github.com/LamkasDev/sharkie/cmd/structs/ipmi"
	"github.com/LamkasDev/sharkie/cmd/structs/output"
	"github.com/LamkasDev/sharkie/cmd/structs/rng"
	"github.com/LamkasDev/sharkie/cmd/symbol"
	"github.com/gookit/color"
)

func main() {
	// Lock the goroutine to its current OS thread.
	// This is crucial because we are manipulating the mem and setting up
	// a thread-local exception handler.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	logger.StartLogging()
	// logger.StartProfiling()

	// Setup host stuff.
	logger.Printf("hi from %s :3\n", color.Blue.Sprint("sharkie"))
	asm.ExceptionHandler = emu.ExceptionHandlerGo
	elf.GetSymbolAddress = emu.GetSymbolAddress
	elf.GetDefiningModule = emu.GetDefiningModule
	asm.InitSignalsAddr()
	asm.InitStubAddr()
	asm.SetupCooperativeGC()
	asm.AllocTlsSlots()
	emu.SetupSignalHandler()
	if err := app.SetupApplication(); err != nil {
		panic(err)
	}

	// Setup guest stuff.
	structs.SetupAllocator()
	structs.SetupSemaphores()
	structs.SetupEventFlags()
	fs.SetupFilesystem()
	rng.SetupRngDevice()
	output.SetupOutputDevices()
	ipmi.SetupImpiManager()
	gc.SetupGraphicsController()
	dce.SetupDisplayCoreEngine()
	gpu.SetupLiverpool()
	gpu.GlobalLiverpool.OnFlip = app.GlobalApplication.Renderer.FrameSource.Submit
	gpu.GlobalLiverpool.OnRegisterDisplaySurface = app.GlobalApplication.Renderer.RegisterFramebuffer

	// Register function stubs.
	symbol.LoadSymbolMap("data/aerolib.csv")
	lib.RegisterStubs()

	// Run main executable.
	if err := emu.GlobalModuleManager.LoadModule("eboot.bin"); err != nil {
		panic(err)
	}
	emu.GlobalModuleManager.RunModule("eboot.bin")

	// Render stuff.
	if err := app.RunApplication(); err != nil {
		panic(err)
	}
	logger.StopProfiling()
	logger.StopLogging()
}
