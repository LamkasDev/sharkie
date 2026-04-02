package renderer

import "sync/atomic"

type Frame struct {
	GpuAddress uintptr
	FlipArg    uint64
}

type FrameSource struct {
	Channel   chan Frame
	IsClosing atomic.Bool
}

func NewFrameSource() *FrameSource {
	return &FrameSource{Channel: make(chan Frame, 2)}
}

func (s *FrameSource) Submit(gpuAddress uintptr, flipArg uint64) {
	if s.IsClosing.Load() {
		return
	}

	select {
	case s.Channel <- Frame{GpuAddress: gpuAddress, FlipArg: flipArg}:
	default:
	}
}
