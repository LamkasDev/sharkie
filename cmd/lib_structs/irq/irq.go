package irq

import "sync"

type InterruptId uint32

const (
	InterruptIdCompute0RelMem = InterruptId(0x00)
	InterruptIdCompute1RelMem = InterruptId(0x01)
	InterruptIdCompute2RelMem = InterruptId(0x02)
	InterruptIdCompute3RelMem = InterruptId(0x03)
	InterruptIdCompute4RelMem = InterruptId(0x04)
	InterruptIdCompute5RelMem = InterruptId(0x05)
	InterruptIdCompute6RelMem = InterruptId(0x06)

	InterruptIdGraphicsFlip = InterruptId(0x08)
	InterruptIdGpuIdle      = InterruptId(0x09)
	InterruptIdGraphicsEop  = InterruptId(0x40)

	InterruptIdMax = InterruptId(0x40)
)

type InterruptHandler struct {
	listeners     map[InterruptId][]func(InterruptId)
	onceListeners map[InterruptId][]func(InterruptId)
	mu            sync.Mutex
}

func NewInterruptHandler() *InterruptHandler {
	return &InterruptHandler{
		listeners:     make(map[InterruptId][]func(InterruptId)),
		onceListeners: make(map[InterruptId][]func(InterruptId)),
	}
}

func (h *InterruptHandler) Register(typ InterruptId, fn func(InterruptId)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.listeners[typ] = append(h.listeners[typ], fn)
}

func (h *InterruptHandler) RegisterOnce(typ InterruptId, fn func(InterruptId)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onceListeners[typ] = append(h.onceListeners[typ], fn)
}

func (h *InterruptHandler) Signal(typ InterruptId) {
	h.mu.Lock()
	callbacks := append([]func(InterruptId){}, h.listeners[typ]...)
	if once, ok := h.onceListeners[typ]; ok && len(once) > 0 {
		callbacks = append(callbacks, once...)
		h.onceListeners[typ] = nil
	}
	h.mu.Unlock()

	for _, cb := range callbacks {
		cb(typ)
	}
}

var GlobalInterruptHandler = NewInterruptHandler()
