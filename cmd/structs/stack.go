package structs

import (
	"encoding/binary"
	"unsafe"
)

const StackAlignment = 8

const StackDefaultSize = 2 * 1024 * 1024 // 2MB
const StackMinimumSize = 0x4000
const StackArgumentsSize = 256

type Stack struct {
	Address                 uintptr
	Top                     uintptr
	Size                    uint64
	ArgumentsAddress        uintptr
	ArgumentsCurrentPointer uintptr
	CurrentPointer          uintptr
}

// NewStack creates a new stack with the defined size.
func NewStack(stackSize uint64) *Stack {
	stackPtr, err := AllocKernelMemory(0, stackSize, PROT_READ|PROT_WRITE|PROT_EXEC, MAP_ANON|MAP_PRIVATE)
	if stackPtr == 0 {
		panic(err)
	}

	// Clear a 128-byte red zone and align to 16-bytes.
	// https://wiki.osdev.org/System_V_ABI
	stackPtr &^= 15
	stackTop := stackPtr + uintptr(stackSize)

	return &Stack{
		Address:                 stackPtr,
		Top:                     stackTop,
		Size:                    stackSize,
		ArgumentsAddress:        stackTop - StackArgumentsSize,
		ArgumentsCurrentPointer: stackTop - StackArgumentsSize,
		CurrentPointer:          stackTop - StackArgumentsSize - 128,
	}
}

// PushUint32 pushes an uint32 argument onto the stack.
// Next argument will also be aligned on 8-byte boundary as per System V ABI AMD64.
// https://c9x.me/compile/doc/abi.html
func (s *Stack) PushUint32(v uint32) uintptr {
	addr := s.ArgumentsCurrentPointer
	binary.LittleEndian.PutUint32(unsafe.Slice((*byte)(unsafe.Pointer(s.ArgumentsCurrentPointer)), 4), v)
	s.ArgumentsCurrentPointer += 8
	return addr
}

// PushUint64 pushes an uint64 argument onto the stack.
// Next argument will also be aligned on 8-byte boundary as per System V ABI AMD64.
// https://c9x.me/compile/doc/abi.html
func (s *Stack) PushUint64(v uint64) uintptr {
	addr := s.ArgumentsCurrentPointer
	binary.LittleEndian.PutUint64(unsafe.Slice((*byte)(unsafe.Pointer(s.ArgumentsCurrentPointer)), 8), v)
	s.ArgumentsCurrentPointer += 8
	return addr
}

// PushString pushes a string argument onto the stack.
// Next argument will also be aligned on 8-byte boundary as per System V ABI AMD64.
// https://c9x.me/compile/doc/abi.html
func (s *Stack) PushString(v string) uintptr {
	addr := s.ArgumentsCurrentPointer
	vLength := uintptr(len(v))
	copy(unsafe.Slice((*byte)(unsafe.Pointer(s.ArgumentsCurrentPointer)), vLength), v)
	padding := (StackAlignment - (vLength % StackAlignment)) % StackAlignment
	s.ArgumentsCurrentPointer += vLength + padding
	return addr
}
