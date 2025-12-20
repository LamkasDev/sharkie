package elf

import "encoding/binary"

const (
	PF_X = 1 // Execute
	PF_W = 2 // Write
	PF_R = 4 // Read
)

type ElfLoadSection struct {
	PFlags  uint32
	POffset uint64
	PVaddr  uint64
	PFilesz uint64
	PMemsz  uint64
}

// NewLoadSection loads the PT_LOAD section at offset.
func (e *Elf) NewLoadSection(data []byte, offset uint64) *ElfLoadSection {
	pFlags := binary.LittleEndian.Uint32(data[offset+0x04:])
	pOffset := binary.LittleEndian.Uint64(data[offset+0x08:])
	pVaddr := binary.LittleEndian.Uint64(data[offset+0x10:])
	pFilesz := binary.LittleEndian.Uint64(data[offset+0x20:])
	pMemsz := binary.LittleEndian.Uint64(data[offset+0x28:])

	return &ElfLoadSection{
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

	if s.POffset+s.PFilesz > uint64(len(data)) {
		available := uint64(len(data)) - s.POffset
		copy(e.Memory[s.PVaddr:], data[s.POffset:s.POffset+available])
	} else {
		copy(e.Memory[s.PVaddr:], data[s.POffset:s.POffset+s.PFilesz])
	}
}
