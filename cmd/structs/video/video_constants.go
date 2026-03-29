package video

type VideoOutPixelFormat int16

const (
	VideoOutPixelFormat_A8R8G8B8_SRGB    = VideoOutPixelFormat(0x0000)
	VideoOutPixelFormat_A8B8G8R8_SRGB    = VideoOutPixelFormat(0x2200)
	VideoOutPixelFormat_A2R10G10B10_SRGB = VideoOutPixelFormat(0x6000)
)

type VideoOutTilingMode int8

const (
	VideoOutTilingModeTile   = VideoOutTilingMode(0)
	VideoOutTilingModeLinear = VideoOutTilingMode(1)
)

type VideoOutAspectRatio int8

const VideoOutAspectRatio_16_9 = VideoOutAspectRatio(0)

const (
	SCE_VIDEO_OUT_ERROR_INVALID_VALUE    = uintptr(0x80A20002)
	SCE_VIDEO_OUT_ERROR_INVALID_HANDLE   = uintptr(0x80A2000B)
	SCE_VIDEO_OUT_ERROR_INVALID_POINTER  = uintptr(0x80A20011)
	SCE_VIDEO_OUT_ERROR_UNSUPPORTED_MODE = uintptr(0x80A2001B)
	SCE_VIDEO_OUT_ERROR_NOT_IMPLEMENTED  = uintptr(0x80A200FF)
)
