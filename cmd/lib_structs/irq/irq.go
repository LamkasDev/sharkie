package irq

import "sync"

type InterruptType int

const (
	InterruptGraphicsFlip InterruptType = iota
)

type InterruptHandler struct {
	listeners map[InterruptType][]func(InterruptType)
	mu        sync.Mutex
}

func NewInterruptHandler() *InterruptHandler {
	return &InterruptHandler{
		listeners: make(map[InterruptType][]func(InterruptType)),
	}
}

func (h *InterruptHandler) Register(typ InterruptType, fn func(InterruptType)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.listeners[typ] = append(h.listeners[typ], fn)
}

func (h *InterruptHandler) Signal(typ InterruptType) {
	h.mu.Lock()
	callbacks := append([]func(InterruptType){}, h.listeners[typ]...)
	h.mu.Unlock()

	for _, cb := range callbacks {
		cb(typ)
	}
}

var GlobalInterruptHandler = NewInterruptHandler()
