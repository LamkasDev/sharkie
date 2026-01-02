package structs

var GlobalGraphicsController *GraphicsController

const (
	SCE_GC_IOCTL_GET_SUBMIT_DONE_ADDRESS = 0xC008811B
	SCE_GC_IOCTL_GET_VM_ID               = 0xC004811F
	SCE_GC_IOCTL_SET_RING_SIZES          = 0xC00C8110
	SCE_GC_IOCTL_SET_MIP_STATS           = 0xC0848119
	SCE_GC_IOCTL_GET_CU_MASK             = 0xC010810B
	SCE_GC_IOCTL_SUBMIT_COMMAND_BUFFERS  = 0xC0108102
)

type GraphicsController struct {
	SubmitDoneAddress uintptr
}

func NewGraphicsController() *GraphicsController {
	return &GraphicsController{
		SubmitDoneAddress: GlobalGoAllocator.Malloc(8),
	}
}

func SetupGraphicsController() {
	GlobalGraphicsController = NewGraphicsController()
}
