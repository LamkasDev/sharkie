//go:build windows

package structs

func SetupMemoryManagerSignalHandler() {
}

func cTrackPage(addr uintptr, protState int) {
	_ = addr
	_ = protState
}

func cUntrackPage(addr uintptr) {
	_ = addr
}

// SyncRequest is delivered by the SIGSEGV handler to the memory sync worker.
type SyncRequest struct {
	Addr    uintptr
	IsWrite bool
}

func WaitForSyncRequest() SyncRequest {
	select {}
}

func CompleteSyncRequest() {
}
