// Package dce contains structs to emulate the Display Core Engine (/dev/dce device).
package dce

import (
	. "github.com/LamkasDev/sharkie/cmd/structs/video"
)

var GlobalDisplayCoreEngine *DisplayCoreEngine

// DisplayCoreEngine keeps state of the /dev/dce device.
type DisplayCoreEngine struct {
	Handles                [VideoOutMaxHandles]VideoOutHandle
	AttributeBufferAddress uintptr
	AttributeBufferSize    uintptr
}

func NewDisplayCoreEngine() *DisplayCoreEngine {
	dce := &DisplayCoreEngine{
		AttributeBufferSize: 0x4000,
	}
	for i := range dce.Handles {
		dce.Handles[i].Id = i + 1
	}

	return dce
}

func SetupDisplayCoreEngine() {
	GlobalDisplayCoreEngine = NewDisplayCoreEngine()
}

func (dce *DisplayCoreEngine) HandleById(id int) *VideoOutHandle {
	if id < 1 || id > VideoOutMaxHandles {
		return nil
	}

	return &dce.Handles[id-1]
}
