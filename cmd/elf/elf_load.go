package elf

import (
	"encoding/binary"

	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// Program header flags indicating segment permissions.
const (
	PF_X = 1 // Execute permission
	PF_W = 2 // Write permission
	PF_R = 4 // Read permission
)

// ElfLoadSection represents a PT_LOAD program header entry in an ELF file.
type ElfLoadSection struct {
	PType   uint32 // Type of segment
	PFlags  uint32 // Segment flags
	POffset uint64 // Offset of the segment in file
	PVaddr  uint64 // Virtual address where the segment should be loaded
	PFilesz uint64 // Size of the segment in file
	PMemsz  uint64 // Size of the segment in memory

	Address    uintptr // Address where the section is actually loaded
	LoadedSize uint64  // Size of the section that is actually loaded
}

// NewLoadSection creates a new ElfLoadSection by parsing the program header entry
// from the ELF data at the given offset.
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

// ProcessLoadSection copies the data specified by the ElfLoadSection from an ELF file
// into allocated memory space for the ELF.
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
