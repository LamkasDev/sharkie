package structs

import "unsafe"

type DceResolutionInfo struct {
	Width       uint32
	Height      uint32
	CropWidth   uint32
	CropHeight  uint32
	RefreshRate uint32
	Interlaced  uint32
	Type        uint32
}

const DceResolutionInfoSize = unsafe.Sizeof(DceResolutionInfo{})

type DcePortStatus struct {
	Connected uint8
	_         [47]byte
}

const DcePortStatusSize = unsafe.Sizeof(DcePortStatus{})
