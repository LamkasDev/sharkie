//go:build windows

package structs

func SetupMemoryManagerSignalHandler() {
}

func cTrackPage(addr uintptr) {
}

func cUntrackPage(addr uintptr) {
}

func WaitForSyncRequest() uintptr {
	// Block forever on Windows since it's not implemented yet
	select {}
}

func CompleteSyncRequest() {
}
