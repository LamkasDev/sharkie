package renderer

type Frame struct {
	GpuAddress uintptr
	FlipArg    uint64
}

type FrameSource struct {
	ch chan Frame
}

func NewFrameSource() *FrameSource {
	return &FrameSource{ch: make(chan Frame, 2)}
}

func (s *FrameSource) Submit(gpuAddress uintptr, flipArg uint64) {
	select {
	case s.ch <- Frame{GpuAddress: gpuAddress, FlipArg: flipArg}:
	default:
	}
}
