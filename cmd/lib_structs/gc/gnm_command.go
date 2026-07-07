package gc

import "unsafe"

const (
	SCE_GC_IOCTL_SUBMIT_DONE            = 0xC0048116
	SCE_GC_IOCTL_DRAIN_RING             = 0xC0048117
	SCE_GC_IOCTL_SET_WORK_MODE          = 0xC004811D
	SCE_GC_IOCTL_SWITCH_BUFFER          = 0xC0088101
	SCE_GC_IOCTL_SUBMIT_COMMAND_BUFFERS = 0xC0108102
	SCE_GC_IOCTL_DINGDONG               = 0xC010811C
	SCE_GC_IOCTL_SUBMIT_AND_FLIP        = 0xC020810C
)

type GnmSubmitDone struct {
	Reserved uint32
}

const GnmSubmitDoneSize = unsafe.Sizeof(GnmSubmitDone{})

type GnmDrainRing struct {
	Reserved uint32
}

const GnmDrainRingSize = unsafe.Sizeof(GnmDrainRing{})

type GnmSetWorkMode struct {
	Mode uint32
}

const GnmSetWorkModeSize = unsafe.Sizeof(GnmSetWorkMode{})

type GnmSwitchBuffer struct {
	RingSlot uint32
	_        uint32
}

const GnmSwitchBufferSize = unsafe.Sizeof(GnmSwitchBuffer{})

type GnmSubmitCommandBuffers struct {
	ContextId          uint32
	Count              uint32
	IndirectBuffersPtr uintptr
}

const GnmSubmitCommandBuffersSize = unsafe.Sizeof(GnmSubmitCommandBuffers{})

type GnmDingDong struct {
	PipeIndex    uint32
	QueueIndex   uint32
	SlotIndex    uint32
	WritePointer uint32
}

const GnmDingDongSize = unsafe.Sizeof(GnmDingDong{})

type GnmSubmitAndFlip struct {
	ContextId          uint32
	Count              uint32
	IndirectBuffersPtr uintptr
	EopAddress         uintptr
	EopValue           uint32
	_                  [4]byte
}

const GnmSubmitAndFlipSize = unsafe.Sizeof(GnmSubmitAndFlip{})
