package renderer

import (
	"sync/atomic"

	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

type Frame struct {
	Number     uint64
	GpuAddress uintptr
	FlipArg    uint64
}

type FrameSource struct {
	Count     uint64
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
	case s.Channel <- Frame{Number: s.Count, GpuAddress: gpuAddress, FlipArg: flipArg}:
		logger.Printf("[%s] submitted to channel.\n",
			color.Blue.Sprintf("Frame %d", s.Count),
		)
		s.Count++
	default:
	}
}
