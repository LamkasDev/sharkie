package video

import (
	"github.com/LamkasDev/sharkie/cmd/lib/kernel"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
)

const (
	VideoOutMaxHandles    = 3
	VideoOutMaxBuffers    = 16
	VideoOutMaxAttributes = 16
)

const (
	VideoOutInternalEventIdFlip   = 0x6
	VideoOutInternalEventIdVblank = 0x7
)

type VideoOutHandle struct {
	Id                 uint32
	Buffers            [VideoOutMaxBuffers]VideoOutBuffer
	Attributes         [VideoOutMaxAttributes]VideoOutBufferAttribute
	LabelBufferAddress uintptr
	Resolution         VideoOutResolutionStatus

	CurrentFlip *VideoOutFlip
	NextFlip    chan *VideoOutFlip
	FlipStatus  VideoOutFlipStatus
	FlipEvents  []VideoOutEvent
	FlipRate    uint32

	VblankStatus VideoOutVblankStatus
	VblankEvents []VideoOutEvent
}

func NewVideoOutHandle(id uint32) *VideoOutHandle {
	return &VideoOutHandle{
		Id:                 id,
		LabelBufferAddress: GlobalGoAllocator.Malloc(uintptr(VideoOutMaxBuffers) * 8),
		Resolution: VideoOutResolutionStatus{
			FullWidth:          1920,
			FullHeight:         1080,
			PaneWidth:          1920,
			PaneHeight:         1080,
			RefreshRate:        3,
			ScreenSizeInInches: 50,
		},
		NextFlip: make(chan *VideoOutFlip, VideoOutMaxBuffers),
	}
}

func (vh *VideoOutHandle) SubmitFlip(flip *VideoOutFlip) {
	vh.FlipStatus.GcQueueNumber++
	vh.FlipStatus.FlipPendingNumber++
	vh.FlipStatus.SubmitTsc = uint64(kernel.SceKernelReadTsc())
	vh.NextFlip <- flip
}

type VideoOutEvent struct {
	EqueueHandle uintptr
	UserData     uintptr
}

type VideoOutFlip struct {
	BufferIndex int32
	FlipArg     int64
	GpuAddress  uintptr
}

type VideoOutFlipStatus struct {
	Count             uint64
	ProcessTime       uint64
	Tsc               uint64
	FlipArg           int64
	SubmitTsc         uint64
	Reserved0         uint64
	GcQueueNumber     int32
	FlipPendingNumber int32
	CurrentBuffer     int32
	Reserved1         uint32
}

// TODO: finish this.
type VideoOutVblankStatus struct {
}

type VideoOutResolutionStatus struct {
	FullWidth          int32
	FullHeight         int32
	PaneWidth          int32
	PaneHeight         int32
	RefreshRate        uint64
	ScreenSizeInInches float32
	Flags              uint16
	Reserved0          uint16
	Reserved1          [3]uint32
}
