package structs

import "unsafe"

const ModuleInfoHandleOffset = uintptr(0x200)

type ModuleInfo struct {
	Size          uint64
	Name          [256]byte
	Segments      [4]SegmentInfo
	SegmentsCount uint32
	Fingerprint   [20]byte
}

const ModuleInfoSize = unsafe.Sizeof(ModuleInfo{})

type ModuleInfoForUnwind struct {
	Size                        uint64
	Name                        [256]byte
	ExceptionFrameHeaderAddress uintptr
	ExceptionFrameAddress       uintptr
	ExceptionFrameSize          uint64
	TextSectionAddress          uintptr
	TextSectionSize             uint64
}

const ModuleInfoForUnwindSize = unsafe.Sizeof(ModuleInfoForUnwind{})

type SegmentInfo struct {
	Address    uintptr
	Size       uint32
	Protection uint32
}

const SegmentInfoSize = unsafe.Sizeof(SegmentInfo{})
