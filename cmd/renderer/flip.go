package renderer

import (
	"sync/atomic"

	"github.com/LamkasDev/sharkie/cmd/lib_structs/video"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

type Flip struct {
	Number uint64
	Flip   *video.VideoOutFlip
}

type FlipSource struct {
	Count     uint64
	Channel   chan *Flip
	IsClosing atomic.Bool
}

func NewFlipSource() *FlipSource {
	return &FlipSource{Channel: make(chan *Flip, 2)}
}

func (s *FlipSource) Submit(flip *video.VideoOutFlip) {
	if s.IsClosing.Load() {
		return
	}

	select {
	case s.Channel <- &Flip{Number: s.Count, Flip: flip}:
		logger.Printf("[%s] submitted flip to channel 0x%X.\n",
			color.Blue.Sprintf("Frame %d", s.Count), flip.GpuAddress,
		)
		s.Count++
	default:
	}
}
