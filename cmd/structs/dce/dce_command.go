package dce

import "unsafe"

const (
	SCE_DCE_IOCTL_CMD              = 0xC0308203
	SCE_DCE_IOCTL_REGISTER_BUFFERS = 0xC0308207
)

const (
	SCE_DCE_IOCTL_CMD_SET_IMAGE_ADDRESS        = 1
	SCE_DCE_IOCTL_CMD_GET_FINISHED_EOP_HANDLE  = 2
	SCE_DCE_IOCTL_CMD_SUB_6DF0                 = 4
	SCE_DCE_IOCTL_CMD_GET_CONNECTION_STATUS    = 5
	SCE_DCE_IOCTL_CMD_GET_RESOLUTION_SUPPORT   = 6
	SCE_DCE_IOCTL_CMD_SET_WINDOW_MODE_MARGINS  = 7
	SCE_DCE_IOCTL_CMD_SUB_3D30_2               = 8
	SCE_DCE_IOCTL_CMD_GET_ATTR_BUFFER_SIZE     = 9
	SCE_DCE_IOCTL_CMD_GET_FLIP_STATUS          = 10
	SCE_DCE_IOCTL_CMD_SUB_1860                 = 11
	SCE_DCE_IOCTL_CMD_SUB_3D30                 = 12
	SCE_DCE_IOCTL_CMD_SET_OUTPUT_CSC           = 17
	SCE_DCE_IOCTL_CMD_GET_VBLANK_STATUS        = 18
	SCE_DCE_IOCTL_CMD_GET_RESOLUTION_STATUS    = 19
	SCE_DCE_IOCTL_CMD_SET_DIPLAY_PARAMETERS    = 20
	SCE_DCE_IOCTL_CMD_SUBMIT_SUB_WINDOW_LAYOUT = 21
	SCE_DCE_IOCTL_CMD_CURSOR_OPERATION         = 24
	SCE_DCE_IOCTL_CMD_GET_PORT_STATUS_INFO     = 25
	SCE_DCE_IOCTL_CMD_SET_INVERTED_COLORS      = 29
	SCE_DCE_IOCTL_CMD_SET_ATTR_BUFFER_ADDRESS  = 31 // idk about this
	SCE_DCE_IOCTL_CMD_RESET_ZOOM_BUFFER        = 32
	SCE_DCE_IOCTL_CMD_SUB_41A0                 = 33
	SCE_DCE_IOCTL_CMD_SET_TONE_MAP             = 34
	SCE_DCE_IOCTL_CMD_GET_STATUS_FOR_WEBCORE   = 36
)

type DceCommand struct {
	CommandId uint32
	_         [4]byte
	Handle    uintptr
	Param1    uintptr
	Param2    uintptr
	Param3    uintptr
	_         [8]byte
}

const DceCommandSize = unsafe.Sizeof(DceCommand{})

type DceRegisterBuffers struct {
	CommandId uint32
	_         [4]byte
	Handle    uint32
	Index     uint32
	Address   uint64
	Size      uint64
	Flags     uint64
	_         [8]byte
}

const DceRegisterBuffersSize = unsafe.Sizeof(DceRegisterBuffers{})
