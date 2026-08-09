package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/LamkasDev/sharkie/cmd/app"
	"github.com/LamkasDev/sharkie/cmd/asm"
	"github.com/LamkasDev/sharkie/cmd/config"
	"github.com/LamkasDev/sharkie/cmd/elf"
	symbol "github.com/LamkasDev/sharkie/cmd/elf_symbol"
	"github.com/LamkasDev/sharkie/cmd/emu"
	"github.com/LamkasDev/sharkie/cmd/lib"
	"github.com/LamkasDev/sharkie/cmd/lib_structs"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/app_content"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/audio"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/dce"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gc"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/gpu"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/ipmi"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/libc"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/net"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/pad"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/posix"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/rng"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/save_data"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/semaphore"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/system_service"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/user"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/user_service"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/structs"
	"github.com/gookit/color"
)

// SetupEmulatorHost initializes the application window and base systems.
func SetupEmulatorHost() error {
	color.Disable()
	logger.StartLogging()

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
	posix.SetupAllocator()
	libc.SetupGoAllocator()
	libc.SetupMspaceAllocator()

	if err := app.SetupApplication(); err != nil {
		return err
	}
	if err := config.LoadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	return nil
}

// SetupEmulatorGuest initializes the PS4 guest environment and launches the game.
func SetupEmulatorGuest(gameNameOrPath string) error {
	// Load game paths.
	if err := config.ResolveGame(gameNameOrPath); err != nil {
		return err
	}

	// Add game directories to module linker paths.
	emu.GlobalModuleManager.LinkPaths = []string{
		filepath.Join(config.GameDirectory, "Image0", "sce_module"),
		filepath.Join(config.GameDirectory, "Image0"),
		config.GetLibDir(),
	}

	// Log any interesting info.
	logger.Printf("hi from %s :3\n", color.Blue.Sprint("sharkie"))
	cachePath, _ := config.AppScope.CacheDir()
	logger.Printf("cache path: %s\n", cachePath)
	dataPath, _ := config.AppScope.DataPath("")
	logger.Printf("data path: %s\n", dataPath)
	logger.Printf("launched game name/path: %s\n", config.GameDirectory)

	// Setup guest stuff.
	semaphore.SetupSemaphores()
	lib_structs.SetupEventFlags()
	fs.SetupFilesystem()
	rng.SetupRngDevice()
	ipmi.SetupImpiManager()
	gc.SetupGraphicsController()
	dce.SetupDisplayCoreEngine()
	audio.SetupAudioEngine()
	pad.SetupPadEngine()
	user.SetupUserManager()
	save_data.SetupSaveDataManager()
	system_service.SetupSystemService()
	user_service.SetupUserService()
	app_content.SetupAppContentInstance()
	net.SetupNetConnectionInstance()
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

	gpu.GlobalLiverpool.OnRegisterDisplaySurface = app.GlobalApplication.Renderer.RegisterFramebuffer
	gpu.GlobalLiverpool.OnFlip = app.GlobalApplication.Renderer.FrameSource.Submit

	gpu.GlobalLiverpool.OnRingWork = app.GlobalApplication.Renderer.RingWorkSource.Submit
	gpu.GlobalLiverpool.WaitOnFinishRingWork = app.GlobalApplication.Renderer.GpuTranslator.WaitOnFence

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
			return fmt.Errorf("failed to convert eboot.bin to eboot.elf: %v", err)
		}
	}

	// Hacky check for Unity games to ensure archive.psarc is generated.
	globalGameManagersPath := filepath.Join(config.GameDirectory, "Image0", "Media", "globalgamemanagers")
	archivePsarcPath := filepath.Join(config.GameDirectory, "Image0", "archive.psarc")
	if _, err := os.Stat(globalGameManagersPath); err == nil {
		if _, err = os.Stat(archivePsarcPath); err != nil {
			logger.Printf("generating archive.psarc for Unity game...\n")
			settingsPath := filepath.Join(config.GameDirectory, "Image0", ".psarc-cl-settings")
			_ = os.WriteFile(settingsPath, []byte("compressionType=zlib\nendianness=big\n"), 0644)
			err = config.RunTool("psarc-cl-linux", "pack", filepath.Join(config.GameDirectory, "Image0"), archivePsarcPath)
			_ = os.Remove(settingsPath)
			if err != nil {
				return fmt.Errorf("failed to pack archive.psarc: %v", err)
			}
		}
	}

	// Run main executable.
	if _, err := emu.GlobalModuleManager.LoadModule("eboot.elf", false); err != nil {
		return err
	}
	emu.GlobalModuleManager.RunModule("eboot.elf")

	return nil
}
