package asm

import (
	"runtime"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/LamkasDev/sharkie/cmd/logger"
)

var (
	// NeedsGC indicates if a GC cycle is required.
	NeedsGC atomic.Bool

	// GCInProgress indicates if a GC cycle is currently running.
	GCInProgress atomic.Bool

	// GCFence acts as a barrier, preventing guest threads from re-entering guest code during GC.
	GCFence atomic.Bool

	// ActiveGuestThreads counts the number of guest threads currently executing guest code.
	ActiveGuestThreads atomic.Int32
)

// SetupCooperativeGC disables automatic GC and starts a ticker.
func SetupCooperativeGC() {
	// Prevent automatic GC while we're on guest stack.
	debug.SetGCPercent(-1)

	// Start a background ticker to signal for a GC.
	/* go func() {
		ticker := time.NewTicker(2 * time.Second)
		for range ticker.C {
			NeedsGC.Store(true)
		}
	}() */
}

// CheckAndRunGC checks if we should GC, waits until all threads are back and sweeps.
func CheckAndRunGC() {
	if !NeedsGC.Load() {
		return
	}
	if !GCInProgress.CompareAndSwap(false, true) {
		return
	}

	// Wait for all threads to return.
	logger.Println("GC waiting for threads...")
	GCFence.Store(true)
	start := time.Now()
	for ActiveGuestThreads.Load() != 0 {
		if time.Since(start) > time.Millisecond*500 {
			logger.Println("GC skipped: threads took too long to park.")
			GCFence.Store(false)
			GCInProgress.Store(false)
			return
		}
		runtime.Gosched()
	}

	// Perform GC, stopping all threads from exiting until done.
	NeedsGC.Store(false)
	logger.Println("GC starting...")
	runtime.GC()
	logger.Println("GC finished.")
	GCFence.Store(false)
	GCInProgress.Store(false)
}

// GuestEnter needs to be called everytime we transition from Go to guest code.
func GuestEnter() {
	for GCFence.Load() {
		runtime.Gosched()
	}
	ActiveGuestThreads.Add(1)
}

// GuestLeave needs to be called everytime we transition from guest to Go code.
func GuestLeave() {
	ActiveGuestThreads.Add(-1)
}
