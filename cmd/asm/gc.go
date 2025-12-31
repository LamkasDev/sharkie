package asm

import (
	"runtime"
	"runtime/debug"
	"sync/atomic"
	"time"
)

var (
	NeedsGC atomic.Bool
)

// SetupCooperativeGC disables automatic GC and starts a ticker.
func SetupCooperativeGC() {
	// Prevent automatic GC while we're on playstation stack.
	debug.SetGCPercent(-1)

	// Start a background ticker to signal for a GC.
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		for range ticker.C {
			NeedsGC.Store(true)
		}
	}()
}

// CheckAndRunGC checks if GC is pending and runs it.
func CheckAndRunGC() {
	if NeedsGC.Swap(false) {
		runtime.GC()
	}
}

func GuestEnter() {
}

func GuestLeave() {
}
