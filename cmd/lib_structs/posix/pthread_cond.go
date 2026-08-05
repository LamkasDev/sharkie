package posix

import (
	"unsafe"

	. "github.com/LamkasDev/sharkie/cmd/lib_structs/time"
	"github.com/gookit/color"
)

const (
	PthreadCondInitializer = 0
)

type PthreadCond struct {
	ClockId ClockId
	Flags   uint32
	Name    string
}

const PthreadCondSize = unsafe.Sizeof(PthreadCond{})

type PthreadCondAttr struct {
	Shared  int32
	ClockId ClockId
}

const PthreadCondAttrSize = unsafe.Sizeof(PthreadCondAttr{})

func GetCondNameText(c *PthreadCond, addr uintptr) string {
	return color.Blue.Sprint(c.Name)
}
