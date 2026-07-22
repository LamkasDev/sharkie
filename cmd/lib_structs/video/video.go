package video

const (
	VideoOutMaxHandles    = 3
	VideoOutMaxBuffers    = 16
	VideoOutMaxAttributes = 16
)

const (
	VideoOutInternalEventIdFlip   = 0x6
	VideoOutInternalEventIdVblank = 0x7
)

type VideoOutEvent struct {
	EqueueHandle uintptr
	UserData     uintptr
}

type VideoOutHandle struct {
	Id                 uint32
	Buffers            [VideoOutMaxBuffers]VideoOutBuffer
	Attributes         [VideoOutMaxAttributes]VideoOutBufferAttribute
	CurrentFlip        *VideoOutFlip
	StagingFlip        *VideoOutFlip
	NextFlip           chan *VideoOutFlip
	FlipRate           uint32
	LabelBufferAddress uintptr
	FlipEvents         []VideoOutEvent
	VblankEvents       []VideoOutEvent
}

type VideoOutFlip struct {
	BufferIndex uint32
	FlipArg     uint64
	GpuAddress  uintptr
}
