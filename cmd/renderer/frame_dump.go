package renderer

import (
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/LamkasDev/sharkie/cmd/logger"
)

var (
	// DumpRequested is set when the user triggers F9 to dump the next frame.
	DumpRequested atomic.Bool

	// DumpPending indicates that command recording was active for the current frame,
	// and ConsumeFlips should capture the screenshot and save the report.
	DumpPending atomic.Bool
)

// RequestFrameDump sets the flag to capture the next frame upon F9 press.
func RequestFrameDump() {
	DumpRequested.Store(true)
	logger.Println("[F9] Screenshot and debug dump requested for next frame!")
}

// IsDumpRequested checks if a dump was requested by the user.
func IsDumpRequested() bool {
	return DumpRequested.Load()
}

// SaveDebugText writes the accumulated text report to dump/frame_%04d.txt.
func SaveDebugText(report string, frameNumber uint64) (string, error) {
	dumpDir := "/home/cute-foxgirls/.cache/sharkie/dump"
	if err := os.MkdirAll(dumpDir, 0755); err != nil {
		return "", err
	}
	filePath := filepath.Join(dumpDir, fmt.Sprintf("frame_%04d.txt", frameNumber))
	if err := os.WriteFile(filePath, []byte(report), 0644); err != nil {
		return "", err
	}
	return filePath, nil
}
