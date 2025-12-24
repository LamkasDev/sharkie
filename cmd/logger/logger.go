package logger

import (
	"fmt"
	"os"
	"sync"
)

const LogToFile = false

var LogFile *os.File
var LogLock = sync.Mutex{}

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
	if err := LogFile.Close(); err != nil {
		panic(err)
	}
	LogFile = nil
}

func CleanupAndExit() {
	StopProfiling()
	StopLogging()
	os.Exit(0)
}

func Print(a ...any) {
	message := fmt.Sprint(a...)
	fmt.Print(message)
	if LogToFile {
		LogLock.Lock()
		defer LogLock.Unlock()
		if _, err := LogFile.Write([]byte(message)); err != nil {
			panic(err)
		}
	}
}

func Printf(format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	fmt.Print(message)
	if LogToFile {
		LogLock.Lock()
		defer LogLock.Unlock()
		if _, err := LogFile.Write([]byte(message)); err != nil {
			panic(err)
		}
	}
}

func Println(a ...any) {
	message := fmt.Sprintln(a...)
	fmt.Print(message)
	if LogToFile {
		LogLock.Lock()
		defer LogLock.Unlock()
		if _, err := LogFile.Write([]byte(message)); err != nil {
			panic(err)
		}
	}
}
