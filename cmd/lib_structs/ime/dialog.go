package ime

import . "github.com/LamkasDev/sharkie/cmd/lib_structs/user"

type ImeDialogStatus uint32

const (
	ImeDialogStatusNone     = ImeDialogStatus(0)
	ImeDialogStatusRunning  = ImeDialogStatus(1)
	ImeDialogStatusFinished = ImeDialogStatus(2)
)

type ImeDialogEndStatus uint32

const (
	ImeDialogEndStatusOk           = ImeDialogEndStatus(0)
	ImeDialogEndStatusUserCanceled = ImeDialogEndStatus(1)
	ImeDialogEndStatusAborted      = ImeDialogEndStatus(2)
)

type ImeDialogResult struct {
	EndStatus ImeDialogEndStatus
	Reserved  [12]byte
}

type ImeDialogParam struct {
	UserId              UserId
	Type                ImeType
	SupportedLanguages  ImeLanguage
	EnterLabel          ImeEnterLabel
	InputMethod         ImeInputMethod
	TextFilterAddress   uintptr
	Options             ImeOption
	MaxTextLength       uint32
	InputTextBuffer     *uint16
	PosX                float32
	PosY                float32
	HorizontalAlignment ImeHorizontalAlignment
	VerticalAlignment   ImeVerticalAlignment
	Placeholder         *uint16
	Title               *uint16
	Reserved            [16]byte
}
