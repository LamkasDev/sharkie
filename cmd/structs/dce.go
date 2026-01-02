package structs

var GlobalDisplayCoreEngine *DisplayCoreEngine

const (
	SCE_DCE_IOCTL_CMD              = 0xC0308203
	SCE_DCE_IOCTL_REGISTER_BUFFERS = 0xC0308207
)

type DisplayCoreEngine struct {
	AttributeBufferSize    uintptr
	AttributeBufferAddress uintptr
}

func NewDisplayCoreEngine() *DisplayCoreEngine {
	return &DisplayCoreEngine{
		AttributeBufferSize: 0x4000,
	}
}

func SetupDisplayCoreEngine() {
	GlobalDisplayCoreEngine = NewDisplayCoreEngine()
}
