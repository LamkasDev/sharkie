package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/LamkasDev/sharkie/cmd/app"
	"github.com/LamkasDev/sharkie/cmd/asm"
	"github.com/LamkasDev/sharkie/cmd/config"
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

	// Load config and game paths.
	if err := config.LoadConfig(); err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}
	if len(os.Args) < 2 {
		fmt.Println("Usage: sharkie <game name or path>")
		os.Exit(1)
	}
	if err := config.ResolveGame(os.Args[1]); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Add game directories to module linker paths.
	emu.GlobalModuleManager.LinkPaths = append(
		emu.GlobalModuleManager.LinkPaths,
		config.GetLibDir(),
		filepath.Join(config.GameDirectory, "Image0"),
		filepath.Join(config.GameDirectory, "Image0", "sce_module"),
	)

	// Log any interesting info.
	logger.Printf("hi from %s :3\n", color.Blue.Sprint("sharkie"))
	cachePath, _ := config.AppScope.CacheDir()
	logger.Printf("cache path: %s\n", cachePath)
	dataPath, _ := config.AppScope.DataPath("")
	logger.Printf("data path: %s\n", dataPath)
	logger.Printf("launched game name/path: %s\n", config.GameDirectory)

	// Setup host stuff.
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

	// Convert eboot.bin to eboot.elf, if necessary.
	ebootElfPath := filepath.Join(config.GameDirectory, "Image0", "eboot.elf")
	ebootBinPath := filepath.Join(config.GameDirectory, "Image0", "eboot.bin")
	if _, err := os.Stat(ebootElfPath); err != nil {
		logger.Printf("converting to eboot.elf using ps4_unfself.py...\n")
		err = config.RunTool("ps4_unfself.py", ebootBinPath)
		if err != nil {
			logger.Printf("failed to convert eboot.bin to eboot.elf: %v\n", err)
			os.Exit(1)
		}
	}

	// Run main executable.
	if err := emu.GlobalModuleManager.LoadModule("eboot.elf"); err != nil {
		panic(err)
	}
	emu.GlobalModuleManager.RunModule("eboot.elf")

	// Render stuff.
	if err := app.RunApplication(); err != nil {
		panic(err)
	}
	logger.StopProfiling()
	logger.StopLogging()
}
