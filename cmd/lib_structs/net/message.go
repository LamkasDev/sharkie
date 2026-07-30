package net

type Iovec struct {
	Base uintptr
	Len  uint64
}

type Msghdr struct {
	Name       uintptr
	NameLen    uint32
	_          [4]byte
	Iov        uintptr
	IovLen     uint32
	_          [4]byte
	Control    uintptr
	ControlLen uint32
	Flags      uint32
}
