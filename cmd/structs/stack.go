package structs

import (
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/sys_struct"
)

var StackDefaultSize = 2 * 1024 * 1024 // 2MB
var StackArgumentsSize = uintptr(256)

type Stack struct {
	Address          uintptr
	ArgumentsAddress uintptr
	ArgumentsOffset  uintptr
	Contents         []byte
	CurrentPointer   uintptr
}

// NewStack creates a new stack with the defined size.
func NewStack(stackSize uintptr) *Stack {
	addr, err := sys_struct.AllocReadWriteMemory(stackSize)
	if err != nil {
		panic(err)
	}

	return &Stack{
		Address:          addr,
		ArgumentsAddress: addr + (stackSize - StackArgumentsSize),
		ArgumentsOffset:  stackSize - StackArgumentsSize,
		Contents:         unsafe.Slice((*byte)(unsafe.Pointer(addr)), stackSize),
		CurrentPointer:   addr + (stackSize - StackArgumentsSize),
	}
}

// PushUint32 pushes an uint32 argument onto the stack.
// Next argument will also be aligned on 8-byte boundary as per System V ABI AMD64.
// https://c9x.me/compile/doc/abi.html
func (s *Stack) PushUint32(v uint32) {
	binary.LittleEndian.PutUint32(s.Contents[s.ArgumentsOffset:], v)
	s.ArgumentsOffset += 8
}

// PushUint64 pushes an uint64 argument onto the stack.
// Next argument will also be aligned on 8-byte boundary as per System V ABI AMD64.
// https://c9x.me/compile/doc/abi.html
func (s *Stack) PushUint64(v uint64) {
	binary.LittleEndian.PutUint64(s.Contents[s.ArgumentsOffset:], v)
	s.ArgumentsOffset += 8
}

// PushString pushes a string argument onto the stack.
// Next argument will also be aligned on 8-byte boundary as per System V ABI AMD64.
// https://c9x.me/compile/doc/abi.html
func (s *Stack) PushString(v string) {
	copy(s.Contents[s.ArgumentsOffset:], v)
	s.ArgumentsOffset += uintptr(len(v)) + uintptr(8-(len(v)%8))
}
