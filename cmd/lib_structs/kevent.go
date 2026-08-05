package lib_structs

const (
	EV_ADD     = 0x0001
	EV_DELETE  = 0x0002
	EV_ENABLE  = 0x0004
	EV_DISABLE = 0x0008
	EV_ONESHOT = 0x0010
	EV_CLEAR   = 0x0020
	EV_EOF     = 0x8000
	EV_ERROR   = 0x4000
)

const (
	KernelEventFilterNone         = 0
	KernelEventFilterRead         = -1
	KernelEventFilterWrite        = -2
	KernelEventFilterAio          = -3
	KernelEventFilterVnode        = -4
	KernelEventFilterProc         = -5
	KernelEventFilterSignal       = -6
	KernelEventFilterTimer        = -7
	KernelEventFilterFs           = -9
	KernelEventFilterLio          = -10
	KernelEventFilterUser         = -11
	KernelEventFilterPolling      = -12
	KernelEventFilterVideoOut     = -13
	KernelEventFilterGraphicsCore = -14
	KernelEventFilterHrTimer      = -15
)

type KernelEvent struct {
	Id          uint64
	Filter      int16
	Flags       uint16
	FilterFlags uint32
	FilterData  uint64
	UserData    uintptr
}
