package libc

import (
	"unsafe"

	. "github.com/LamkasDev/sharkie/cmd/lib_structs"
	. "github.com/LamkasDev/sharkie/cmd/lib_structs/fs"
)

type LibcFile struct {
	Mode       LibcFileMode
	Index      uint8
	_          uint8
	Handle     FileDescriptor
	Buffer     uintptr
	BufferEnd  uintptr
	BufferNext uintptr
	ReadEnd    uintptr
	WriteEnd   uintptr
	RBack      uintptr
	WRBack     uintptr
	WBack      [2]uint16
	Unk1       uint16
	_          [2]uint8
	RSave      uintptr
	WREnd      uintptr
	WWEnd      uintptr
	WState     [16]uint8
	TempName   uintptr
	Back       [6]uint8
	CharBuffer uint8
	Unk2       uint8
	Mutex      uintptr
	Reserved   [312]uint8
}

var LibcFileSize = unsafe.Sizeof(LibcFile{})

func NewLibcFile(address uintptr) *LibcFile {
	file := (*LibcFile)(unsafe.Pointer(address))
	cbufPtr := uintptr(unsafe.Pointer(&file.CharBuffer))
	file.Buffer = cbufPtr
	file.BufferEnd = uintptr(unsafe.Pointer(&file.Unk2))
	file.BufferNext = cbufPtr
	file.ReadEnd = cbufPtr
	file.WriteEnd = cbufPtr
	file.RBack = cbufPtr
	file.WRBack = uintptr(unsafe.Pointer(&file.Unk1))
	file.WREnd = cbufPtr
	file.WWEnd = cbufPtr

	return file
}

var LibcFiles []*LibcFile

func InitStandardFiles() {
	if len(LibcFiles) > 0 {
		return
	}
	fileAddress := GlobalGoAllocator.MallocAligned(unsafe.Sizeof(LibcFile{})*3, 8)

	stdinFile := NewLibcFile(fileAddress)
	stdinFile.Mode = M_ACT | M_OPENR
	stdinFile.Index = 0
	stdinFile.Handle = 0

	stdoutFile := NewLibcFile(fileAddress + LibcFileSize)
	stdoutFile.Mode = M_ACT | M_OPENW
	stdoutFile.Index = 1
	stdoutFile.Handle = 1

	stderrFile := NewLibcFile(fileAddress + LibcFileSize*2)
	stderrFile.Mode = M_ACT | M_OPENW
	stderrFile.Index = 2
	stderrFile.Handle = 2

	LibcFiles = append(LibcFiles, stdinFile, stdoutFile, stderrFile)
}
