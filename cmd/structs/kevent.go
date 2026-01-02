package structs

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
	EVFILT_VBLANK = -13
	EVFILT_USER   = 0
)

type Kevent struct {
	Id          uint64
	Filter      int16
	Flags       uint16
	FilterFlags uint32
	FilterData  int64
	UserData    uintptr
}
