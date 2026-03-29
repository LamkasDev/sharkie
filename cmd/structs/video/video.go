package video

const (
	VideoOutMaxHandles    = 3
	VideoOutMaxBuffers    = 16
	VideoOutMaxAttributes = 16
)

type VideoOutHandle struct {
	Id            int
	Buffers       [VideoOutMaxBuffers]VideoOutBuffer
	Attributes    [VideoOutMaxAttributes]VideoOutBufferAttribute
	CurrentBuffer uint32
}
