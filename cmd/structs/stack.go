package structs

import (
	"encoding/binary"
	"unsafe"
)

const StackAlignment = 8

const StackDefaultSize = uintptr(2 * 1024 * 1024) // 2MB
const StackMinimumSize = 0x4000
const StackArgumentsSize = uintptr(256)

type Stack struct {
	Address                 uintptr
	Top                     uintptr
	ArgumentsAddress        uintptr
	ArgumentsCurrentPointer uintptr
	Contents                []byte
	CurrentPointer          uintptr
}

// NewStack creates a new stack with the defined size.
func NewStack(stackSize uintptr) *Stack {
	stackPtr, err := AllocKernelMemory(0, stackSize, PROT_READ|PROT_WRITE|PROT_EXEC, MAP_ANON|MAP_PRIVATE)
	if stackPtr == 0 {
		panic(err)
	}
	stackPtr &^= 15
	stackTop := stackPtr + stackSize

	return &Stack{
		Address:                 stackPtr,
		Top:                     stackTop,
		ArgumentsAddress:        stackTop - StackArgumentsSize,
		ArgumentsCurrentPointer: stackTop - StackArgumentsSize,
		Contents:                unsafe.Slice((*byte)(unsafe.Pointer(stackPtr)), stackSize),
		CurrentPointer:          stackTop - StackArgumentsSize,
	}
}

// PushUint32 pushes an uint32 argument onto the stack.
// Next argument will also be aligned on 8-byte boundary as per System V ABI AMD64.
// https://c9x.me/compile/doc/abi.html
func (s *Stack) PushUint32(v uint32) uintptr {
	addr := s.ArgumentsCurrentPointer
	binary.LittleEndian.PutUint32(s.Contents[s.ArgumentsCurrentPointer:], v)
	s.ArgumentsCurrentPointer += 8
	return addr
}

// PushUint64 pushes an uint64 argument onto the stack.
// Next argument will also be aligned on 8-byte boundary as per System V ABI AMD64.
// https://c9x.me/compile/doc/abi.html
func (s *Stack) PushUint64(v uint64) uintptr {
	addr := s.ArgumentsCurrentPointer
	binary.LittleEndian.PutUint64(s.Contents[s.ArgumentsCurrentPointer:], v)
	s.ArgumentsCurrentPointer += 8
	return addr
}

// PushString pushes a string argument onto the stack.
// Next argument will also be aligned on 8-byte boundary as per System V ABI AMD64.
// https://c9x.me/compile/doc/abi.html
func (s *Stack) PushString(v string) uintptr {
	addr := s.ArgumentsCurrentPointer
	copy(s.Contents[s.ArgumentsCurrentPointer:], v)
	vLength := uintptr(len(v))
	padding := (StackAlignment - (vLength % StackAlignment)) % StackAlignment
	s.ArgumentsCurrentPointer += vLength + padding
	return addr
}
