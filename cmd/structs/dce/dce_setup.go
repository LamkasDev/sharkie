package dce

import "unsafe"

type DceResolutionStatus struct {
	Width            uint32
	Height           uint32
	PaneWidth        uint32
	PaneHeight       uint32
	RefreshRate      uint64
	ScreenSizeInches float32
	Flags            uint16
	_                [14]byte
}

const DceResolutionStatusSize = unsafe.Sizeof(DceResolutionStatus{})

type DcePortStatusInfo struct {
	Connected uint8
	_         [47]byte
}

const DcePortStatusInfoSize = unsafe.Sizeof(DcePortStatusInfo{})
