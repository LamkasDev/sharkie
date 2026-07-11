package renderer

import (
	"sync/atomic"

	"github.com/LamkasDev/sharkie/cmd/lib_structs/gpu"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

type RingWork struct {
	Number   uint64
	RingWork *gpu.RingWork
}

type RingWorkSource struct {
	Count     uint64
	Channel   chan *RingWork
	IsClosing atomic.Bool
	OnSubmit  func()
}

func NewRingWorkSource() *RingWorkSource {
	return &RingWorkSource{Channel: make(chan *RingWork, 2)}
}

func (s *RingWorkSource) Submit(ringWork *gpu.RingWork) {
	if s.IsClosing.Load() {
		return
	}

	select {
	case s.Channel <- &RingWork{Number: s.Count, RingWork: ringWork}:
		if s.OnSubmit != nil {
			s.OnSubmit()
		}
		logger.Printf("[%s] submitted ring work to channel.\n",
			color.Blue.Sprintf("Frame %d", s.Count),
		)
		s.Count++
	default:
	}
}
