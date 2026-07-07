package translation

import (
	"runtime"
	"sync/atomic"
	"time"

	"github.com/LamkasDev/sharkie/cmd/lib_structs"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/structs"
)

var (
	syncReadFaults  atomic.Uint64
	syncWriteFaults atomic.Uint64
	syncDownloads   atomic.Uint64
)

// MemorySyncStats returns cumulative memory-sync worker counters (for diagnostics).
func MemorySyncStats() (readFaults, writeFaults, downloads uint64) {
	return syncReadFaults.Load(), syncWriteFaults.Load(), syncDownloads.Load()
}

func (t *GpuTranslator) memorySyncWorker() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	logger.Printf("Memory Sync Worker started.\n")

	var lastLog time.Time
	var lastRead, lastWrite uint64

	for {
		req := structs.WaitForSyncRequest()
		if req.Addr == 0 {
			continue
		}

		if req.IsWrite {
			syncWriteFaults.Add(1)
			t.InvalidateMemory(req.Addr, lib_structs.SystemPageSize)
		} else {
			syncReadFaults.Add(1)
			before := syncDownloads.Load()
			t.ReadMemory(req.Addr, lib_structs.SystemPageSize)
			if syncDownloads.Load() > before {
				// downloadVkImageToGuest increments via recordDownload
			}
		}

		structs.CompleteSyncRequest()

		now := time.Now()
		if now.Sub(lastLog) < 2*time.Second {
			continue
		}
		read := syncReadFaults.Load()
		write := syncWriteFaults.Load()
		dl := syncDownloads.Load()
		dRead := read - lastRead
		dWrite := write - lastWrite
		if dRead+dWrite > 100 {
			logger.Printf("MemorySync: +%d read faults, +%d write faults, %d downloads total (last 2s)\n",
				dRead, dWrite, dl)
		}
		lastLog = now
		lastRead = read
		lastWrite = write
	}
}

func recordSyncDownload() {
	syncDownloads.Add(1)
}
