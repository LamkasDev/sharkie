package pthread

import (
	"unsafe"

	"github.com/gookit/color"
)

const (
	PthreadCondInitializer = 0
)

type PthreadCond struct {
	KernelId uintptr
	Flags    uint32
	_        [20]byte // Bigggg padding!
	Name     string
}

const PthreadCondSize = unsafe.Sizeof(PthreadCond{})

func GetCondNameText(c *PthreadCond, addr uintptr) string {
	return color.Blue.Sprint(c.Name)
}
