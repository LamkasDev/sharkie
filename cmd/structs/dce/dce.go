// Package dce contains structs to emulate the Display Core Engine (/dev/dce device).
package dce

import (
	. "github.com/LamkasDev/sharkie/cmd/structs/video"
)

var GlobalDisplayCoreEngine *DisplayCoreEngine

// DisplayCoreEngine keeps state of the /dev/dce device.
type DisplayCoreEngine struct {
	Handles    map[uint32]*VideoOutHandle
	NextHandle uint32
}

func NewDisplayCoreEngine() *DisplayCoreEngine {
	return &DisplayCoreEngine{
		Handles:    map[uint32]*VideoOutHandle{},
		NextHandle: 0x40000001,
	}
}

func SetupDisplayCoreEngine() {
	GlobalDisplayCoreEngine = NewDisplayCoreEngine()
}
