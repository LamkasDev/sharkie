package elf

import (
	"encoding/binary"

	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

const (
	PF_X = 1 // Execute
	PF_W = 2 // Write
	PF_R = 4 // Read
)

type ElfLoadSection struct {
	PType   uint32
	PFlags  uint32
	POffset uint64
	PVaddr  uint64
	PFilesz uint64
	PMemsz  uint64

	Address    uintptr
	LoadedSize uint64
}

// NewLoadSection loads the PT_LOAD section at offset.
func (e *Elf) NewLoadSection(data []byte, offset uint64) *ElfLoadSection {
	pType := binary.LittleEndian.Uint32(data[offset:])
	pFlags := binary.LittleEndian.Uint32(data[offset+0x04:])
	pOffset := binary.LittleEndian.Uint64(data[offset+0x08:])
	pVaddr := binary.LittleEndian.Uint64(data[offset+0x10:])
	pFilesz := binary.LittleEndian.Uint64(data[offset+0x20:])
	pMemsz := binary.LittleEndian.Uint64(data[offset+0x28:])

	return &ElfLoadSection{
		PType:   pType,
		PFlags:  pFlags,
		POffset: pOffset,
		PVaddr:  pVaddr,
		PFilesz: pFilesz,
		PMemsz:  pMemsz,
	}
}

// ProcessLoadSection copies data the by PT_LOAD section into memory.
func ProcessLoadSection(e *Elf, s *ElfLoadSection, data []byte) {
	if s.PFilesz == 0 {
		return
	}

	s.Address = e.BaseAddress + uintptr(s.PVaddr)
	s.LoadedSize = s.PFilesz
	if s.POffset+s.LoadedSize > uint64(len(data)) {
		logger.Print(color.Red.Sprintf("Loaded only %d bytes of section sized %d.\n",
			uint64(len(data))-s.POffset,
			s.LoadedSize,
		))
		s.LoadedSize = uint64(len(data)) - s.POffset
	}
	copy(e.Memory[s.PVaddr:], data[s.POffset:s.POffset+s.LoadedSize])
}
