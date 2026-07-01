package renderer

import (
	"sync/atomic"

	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/spirv/structs"
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
	OnSubmit  func()
}

func NewFrameSource() *FrameSource {
	return &FrameSource{Channel: make(chan Frame, 2)}
}

func (s *FrameSource) Submit(gpuAddress uintptr, flipArg uint64) {
	if s.IsClosing.Load() {
		return
	}

	select {
	case s.Channel <- Frame{Number: s.Count, GpuAddress: structs.GetPhysicalGpuAddress(gpuAddress), FlipArg: flipArg}:
		if s.OnSubmit != nil {
			s.OnSubmit()
		}
		logger.Printf("[%s] submitted to channel 0x%X.\n",
			color.Blue.Sprintf("Frame %d", s.Count), gpuAddress,
		)
		s.Count++
	default:
	}
}
