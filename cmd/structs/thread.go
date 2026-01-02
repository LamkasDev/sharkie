package structs

import "unsafe"

type ThreadSignalMask struct {
	Low  uint64
	High uint64
}

const ThreadSignalMaskSize = unsafe.Sizeof(ThreadSignalMask{})

type ThreadAffinityMask uint64

type ThreadCpuSet struct {
	Low  uint64
	High uint64
}

const ThreadCpuSetSize = unsafe.Sizeof(ThreadCpuSet{})
