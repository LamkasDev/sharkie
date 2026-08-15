package ime

import "sync"

type ImeEventId uint32

const (
	ImeEventIdOpen                = ImeEventId(0)
	ImeEventIdUpdateText          = ImeEventId(1)
	ImeEventIdPressClose          = ImeEventId(4)
	ImeEventIdPressEnter          = ImeEventId(5)
	ImeEventIdKeyboardOpen        = ImeEventId(256)
	ImeEventIdKeyboardKeycodeDown = ImeEventId(257)
	ImeEventIdKeyboardKeycodeUp   = ImeEventId(258)
)

type ImeTextAreaMode uint32

const (
	ImeTextAreaModeEdit = ImeTextAreaMode(1)
)

type ImeTextAreaProperty struct {
	Mode   ImeTextAreaMode
	Index  uint32
	Length int32
}

type ImeEditText struct {
	Str        uintptr // pointer to char16_t*
	CaretIndex uint32
	AreaNum    uint32
	TextArea   [4]ImeTextAreaProperty
}

type ImeRect struct {
	X, Y, Width, Height float32
}

type ImeEventParam struct {
	Data [64]byte
}

type ImeEvent struct {
	Id    ImeEventId
	_     [4]byte
	Param ImeEventParam
}

type ImeDevice struct {
	Param  ImeParam
	IsOpen bool

	KeyboardParam  ImeKeyboardParam
	IsKeyboardOpen bool
	KeyboardUserId int32

	DialogParam  ImeDialogParam
	IsDialogOpen bool
	DialogStatus ImeDialogStatus
	DialogResult ImeDialogResult

	InputText string
	Events    []ImeEvent
	Mutex     sync.Mutex
}

var GlobalImeDevice = &ImeDevice{}

func (d *ImeDevice) SendEvent(event ImeEvent) {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	d.Events = append(d.Events, event)
}
