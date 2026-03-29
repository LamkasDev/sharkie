// Package gc contains structs to emulate the Graphics Controller (/dev/gc device).
package gc

import "github.com/LamkasDev/sharkie/cmd/structs"

var GlobalGraphicsController *GraphicsController

const (
	SCE_GC_IOCTL_GET_VM_ID               = 0xC004811F
	SCE_GC_IOCTL_GET_SUBMIT_DONE_ADDRESS = 0xC008811B
	SCE_GC_IOCTL_SET_MIP_STATS           = 0xC0848119
)

// GraphicsController keeps state of the /dev/gc device.
type GraphicsController struct {
	SubmitDoneAddress uintptr
	ActiveRingSlot    uint32
	PendingSubmits    uint32
}

func NewGraphicsController() *GraphicsController {
	return &GraphicsController{
		SubmitDoneAddress: structs.GlobalGoAllocator.Malloc(8),
	}
}

func SetupGraphicsController() {
	GlobalGraphicsController = NewGraphicsController()
}
