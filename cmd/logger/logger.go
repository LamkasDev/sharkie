package logger

import (
	"fmt"
	"os"
)

const LogToFile = false

const (
	// Logs mutexes, conditions, semaphores, etc.
	LogSyncing = false

	// Logs syncing timeouts, attempts, etc.
	LogSyncingFail = true

	// Logs when thread goes to sleep.
	LogSleep = false

	// Logs memory allocations.
	LogAlloc = false

	// Logs gnm & video functions.
	LogGraphics = false

	// Logs things that clutter log (value returns, etc).
	LogMisc = false

	// Logs error returns (can't ignore them forever).
	LogErrorRet = false
)

const FiosDebugMode = false

var LogFile *os.File

func StartLogging() {
	if !LogToFile {
		return
	}
	var err error
	LogFile, err = os.Create("sharkie.log")
	if err != nil {
		panic(err)
	}
}

func StopLogging() {
	if LogFile == nil {
		return
	}
	file := LogFile
	LogFile = nil
	if err := file.Close(); err != nil {
		panic(err)
	}
}

func CleanupAndExit() {
	fmt.Println("Exiting...")
	StopProfiling()
	StopLogging()
	os.Exit(0)
}

func Print(a ...any) {
	message := fmt.Sprint(a...)
	fmt.Print(message)
	if LogToFile && LogFile != nil {
		if _, err := LogFile.Write([]byte(message)); err != nil {
			panic(err)
		}
	}
}

func Printf(format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	fmt.Print(message)
	if LogToFile && LogFile != nil {
		if _, err := LogFile.Write([]byte(message)); err != nil {
			panic(err)
		}
	}
}

func Println(a ...any) {
	message := fmt.Sprintln(a...)
	fmt.Print(message)
	if LogToFile && LogFile != nil {
		if _, err := LogFile.Write([]byte(message)); err != nil {
			panic(err)
		}
	}
}
