package posix

import (
	"unsafe"

	"github.com/gookit/color"
)

const (
	ThrRwlockInitializer = 0
	ThrRwlockDestroyed   = 2
)

type PthreadRwlock struct {
	Name string
}

const PthreadRwlockSize = unsafe.Sizeof(PthreadRwlock{})

type PthreadRwlockAttr struct {
	Shared int32
	Type   int32
}

const PthreadRwlockAttrSize = unsafe.Sizeof(PthreadRwlockAttr{})

func GetRwlockNameText(m *PthreadRwlock, addr uintptr) string {
	return color.Blue.Sprint(m.Name)
}
