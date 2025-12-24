package structs

import "unsafe"

const (
	PthreadCondInitializer = 0
)

type PthreadCond struct {
	KernelId uintptr
	Flags    uint32
	_        [20]byte // Bigggg padding!
}

const PthreadCondSize = unsafe.Sizeof(PthreadCond{})
