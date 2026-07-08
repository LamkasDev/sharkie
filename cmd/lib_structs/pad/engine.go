// Package pad contains structs to emulate controller devices.
package pad

var GlobalPadEngine *PadEngine

// PadEngine keeps state of controller devices.
type PadEngine struct {
	Handles    map[uint32]*PadHandle
	NextHandle uint32
}

func NewPadEngine() *PadEngine {
	return &PadEngine{
		Handles:    map[uint32]*PadHandle{},
		NextHandle: 0x20000001,
	}
}

func SetupPadEngine() {
	GlobalPadEngine = NewPadEngine()
}
