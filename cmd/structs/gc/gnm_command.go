package gc

import "unsafe"

const (
	SCE_GC_IOCTL_WAIT_FLIP_DONE         = 0xC0048116
	SCE_GC_IOCTL_SWITCH_BUFFER          = 0xC0088101
	SCE_GC_IOCTL_SUBMIT_COMMAND_BUFFERS = 0xC0108102
	SCE_GC_IOCTL_SUBMIT_AND_FLIP        = 0xC020810C
)

type GnmIoctlWaitFlipDone struct {
	Reserved uint32
}

const GnmIoctlWaitFlipDoneSize = unsafe.Sizeof(GnmIoctlWaitFlipDone{})

type GnmIoctlSwitchBuffer struct {
	RingIndex uint32
	_         uint32
}

const GnmIoctlSwitchBufferSize = unsafe.Sizeof(GnmIoctlSwitchBuffer{})

type GnmSubmitCommandBuffers struct {
	ContextId          uint32
	Count              uint32
	IndirectBuffersPtr uintptr
}

const GnmSubmitCommandBuffersSize = unsafe.Sizeof(GnmSubmitCommandBuffers{})

type GnmIoctlSubmitAndFlip struct {
	ContextId          uint32
	Count              uint32
	IndirectBuffersPtr uintptr
	EopAddress         uintptr
	EopValue           uint32
	_                  [4]byte
}

const GnmIoctlSubmitAndFlipSize = unsafe.Sizeof(GnmIoctlSubmitAndFlip{})
