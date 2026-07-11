package video

const (
	VideoOutMaxHandles    = 3
	VideoOutMaxBuffers    = 16
	VideoOutMaxAttributes = 16
)

type VideoOutHandle struct {
	Id                 uint32
	Buffers            [VideoOutMaxBuffers]VideoOutBuffer
	Attributes         [VideoOutMaxAttributes]VideoOutBufferAttribute
	CurrentFlip        *VideoOutFlip
	StagingFlip        *VideoOutFlip
	NextFlip           chan *VideoOutFlip
	FlipRate           uint32
	LabelBufferAddress uintptr
}

type VideoOutFlip struct {
	BufferIndex uint32
	FlipArg     uint64
	GpuAddress  uintptr
}
