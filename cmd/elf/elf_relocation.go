package elf

import "encoding/binary"

const (
	// https://docs.oracle.com/cd/E19120-01/open.solaris/819-0690/chapter7-2/index.html
	R_AMD64_64        = 1
	R_AMD64_GLOB_DAT  = 6
	R_AMD64_JUMP_SLOT = 7
	R_AMD64_RELATIVE  = 8
	R_AMD64_DTPOFF64  = 16
)

type ElfRelocation struct {
	Offset uint64
	Type   uint32
	Symbol uint32
	Addend int64
}

type ElfRelocationTable struct {
	Relocations []ElfRelocation
}

// NewRelocationTable loads a relocation table at offset.
func NewRelocationTable(data []byte, tableOffset, tableSize, tableEnt uint64) *ElfRelocationTable {
	if tableSize == 0 || tableEnt == 0 {
		return nil
	}

	table := &ElfRelocationTable{
		Relocations: make([]ElfRelocation, 0, tableSize/tableEnt),
	}

	for i := uint64(0); i < tableSize; i += tableEnt {
		relOffset := tableOffset + i
		if relOffset+24 > uint64(len(data)) {
			break
		}
		rOffset := binary.LittleEndian.Uint64(data[relOffset:])
		rInfo := binary.LittleEndian.Uint64(data[relOffset+8:])
		rType := uint32(rInfo & 0xFFFFFFFF)
		rSym := uint32(rInfo >> 32)
		rAddend := int64(binary.LittleEndian.Uint64(data[relOffset+16:]))

		table.Relocations = append(table.Relocations, ElfRelocation{
			Offset: rOffset,
			Type:   rType,
			Symbol: rSym,
			Addend: rAddend,
		})
	}

	return table
}
