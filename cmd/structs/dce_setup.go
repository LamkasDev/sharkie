package structs

import "unsafe"

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

const (
	SCE_DCE_REFRESH_RATE_UNKNOWN  = 0
	SCE_DCE_REFRESH_RATE_23_98HZ  = 1
	SCE_DCE_REFRESH_RATE_50HZ     = 2
	SCE_DCE_REFRESH_RATE_59_94HZ  = 3
	SCE_DCE_REFRESH_RATE_119_88HZ = 13
	SCE_DCE_REFRESH_RATE_89_91HZ  = 35
	SCE_DCE_REFRESH_RATE_ANY      = 0xFFFFFFFFFFFFFFFF
)

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
