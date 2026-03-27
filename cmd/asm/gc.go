package asm

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"sync/atomic"
	"time"
)

var (
	NeedsGC            atomic.Bool
	GCFence            atomic.Bool
	GCInProgress       atomic.Bool
	ActiveGuestThreads atomic.Int32
)

// SetupCooperativeGC disables automatic GC and starts a ticker.
func SetupCooperativeGC() {
	// Prevent automatic GC while we're on playstation stack.
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
	fmt.Println("GC waiting for threads")
	GCFence.Store(true)
	start := time.Now()
	for ActiveGuestThreads.Load() != 0 {
		if time.Since(start) > time.Second {
			panic("GC deadlock: guest thread did not park")
		}
		runtime.Gosched()
	}

	// Perform GC, stopping all threads from exiting until done.
	NeedsGC.Store(false)
	fmt.Println("GC starting")
	runtime.GC()
	fmt.Println("GC finished")
	GCFence.Store(false)
	GCInProgress.Store(false)
}

func GuestEnter() {
	for GCFence.Load() {
		runtime.Gosched()
	}
	ActiveGuestThreads.Add(1)
}

func GuestLeave() {
	ActiveGuestThreads.Add(-1)
}
