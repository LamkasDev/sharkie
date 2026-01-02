package structs

import "unsafe"

const ModuleInfoHandleOffset = uintptr(0x200)

type ModuleInfo struct {
	Size          uintptr
	Name          [256]byte
	Segments      [4]SegmentInfo
	SegmentsCount uint32
	Fingerprint   [20]byte
}

const ModuleInfoSize = unsafe.Sizeof(ModuleInfo{})

type SegmentInfo struct {
	Address    uintptr
	Size       uint32
	Protection uint32
}

const SegmentInfoSize = unsafe.Sizeof(SegmentInfo{})
