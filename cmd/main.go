package main

import (
	"runtime"

	"github.com/LamkasDev/sharkie/cmd/app"
	"github.com/LamkasDev/sharkie/cmd/asm"
	"github.com/LamkasDev/sharkie/cmd/elf"
	symbol "github.com/LamkasDev/sharkie/cmd/elf_symbol"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/lib"
	"github.com/LamkasDev/sharkie/cmd/lib/libc"
	"github.com/LamkasDev/sharkie/cmd/lib_structs"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/audio"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/dce"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gc"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gpu"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/ipmi"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/rng"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

func main() {
	color.Disable()

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
	structs.SetupMemoryManagerSignalHandler()
	lib_structs.SetupAllocator()
	if err := app.SetupApplication(); err != nil {
		panic(err)
	}

	// Setup guest stuff.
	libc.SetupMspaceAllocator()
	lib_structs.SetupSemaphores()
	lib_structs.SetupEventFlags()
	fs.SetupFilesystem()
	rng.SetupRngDevice()
	ipmi.SetupImpiManager()
	gc.SetupGraphicsController()
	dce.SetupDisplayCoreEngine()
	audio.SetupAudioEngine()
	gpu.SetupLiverpool()

	// Hook functions.
	fs.OutputPrintf = func(message string) {
		logger.Printf("%-132s %s",
			emu.GlobalModuleManager.GetCallSiteText(),
			message,
		)
	}
	fs.OutputPrintln = func() {
		logger.Println()
	}
	gpu.GlobalLiverpool.OnFlip = app.GlobalApplication.Renderer.FrameSource.Submit
	gpu.GlobalLiverpool.OnRegisterDisplaySurface = app.GlobalApplication.Renderer.RegisterFramebuffer
	gpu.GlobalLiverpool.WaitOnFence = app.GlobalApplication.Renderer.WaitOnFence

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
