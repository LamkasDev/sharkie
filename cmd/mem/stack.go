package mem

var StackDefaultSize = 2 * 1024 * 1024 // 2MB
var StackArgumentsSize = uintptr(256)

type Stack struct {
	Address          uintptr
	ArgumentsAddress uintptr
	ArgumentsOffset  uintptr
	Contents         []byte
	CurrentPointer   uintptr
}
