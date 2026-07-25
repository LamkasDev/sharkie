package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/LamkasDev/sharkie/cmd/app"
	"github.com/LamkasDev/sharkie/cmd/bootstrap"
	"github.com/LamkasDev/sharkie/cmd/logger"
)

func main() {
	// Lock the goroutine to its current OS thread.
	// This is crucial because we are manipulating the mem and setting up
	// a thread-local exception handler.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if len(os.Args) < 2 {
		fmt.Println("Usage: sharkie <game name or path>")
		os.Exit(1)
	}

	// Setup emulator stuff.
	if err := bootstrap.SetupEmulatorHost(); err != nil {
		fmt.Printf("Failed to setup emulator host: %v\n", err)
		os.Exit(1)
	}
	if err := bootstrap.SetupEmulatorGuest(os.Args[1]); err != nil {
		fmt.Printf("Failed to setup emulator guest: %v\n", err)
		os.Exit(1)
	}

	// Render stuff.
	if err := app.RunApplication(); err != nil {
		panic(err)
	}
	logger.StopProfiling()
	logger.StopLogging()
}
