package logger

import (
	"os"
	"runtime/pprof"
)

var ProfilerFile *os.File

func StartProfiling() {
	var err error
	ProfilerFile, err = os.Create("sharkie.prof")
	if err != nil {
		panic(err)
	}
	err = pprof.StartCPUProfile(ProfilerFile)
	if err != nil {
		panic(err)
	}
}

func StopProfiling() {
	if ProfilerFile == nil {
		return
	}
	pprof.StopCPUProfile()
	if err := ProfilerFile.Close(); err != nil {
		panic(err)
	}
	ProfilerFile = nil
}
