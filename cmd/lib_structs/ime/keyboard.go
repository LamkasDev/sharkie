package ime

import . "github.com/LamkasDev/sharkie/cmd/lib_structs/user"

type ImeKeyboardOption uint32

const (
	ImeKeyboardOptionDefault                     = ImeKeyboardOption(0)
	ImeKeyboardOptionRepeat                      = ImeKeyboardOption(1 << 0)
	ImeKeyboardOptionRepeatEachKey               = ImeKeyboardOption(1 << 1)
	ImeKeyboardOptionAddOsk                      = ImeKeyboardOption(1 << 2)
	ImeKeyboardOptionEffectiveWithIme            = ImeKeyboardOption(1 << 3)
	ImeKeyboardOptionDisableResume               = ImeKeyboardOption(1 << 4)
	ImeKeyboardOptionDisableCapslockWithoutShift = ImeKeyboardOption(1 << 5)
)

type ImeKeyboardParam struct {
	Option              ImeKeyboardOption
	Reserved1           [4]byte
	Arg                 uintptr
	EventHandlerAddress uintptr
	Reserved2           [8]byte
}

type ImeKeyboardResourceIdArray struct {
	UserId      UserId
	ResourceIds [5]uint32
}
