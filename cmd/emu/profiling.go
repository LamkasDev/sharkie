package emu

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
	pprof.StopCPUProfile()
	ProfilerFile.Close()
}
