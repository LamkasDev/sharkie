package elf

import "encoding/binary"

// Relocation types for AMD64.
const (
	R_AMD64_64        = 1  // Direct 64-bit relocation
	R_AMD64_GLOB_DAT  = 6  // Global data segment relocation
	R_AMD64_JUMP_SLOT = 7  // PLT relocation
	R_AMD64_RELATIVE  = 8  // Relative relocation
	R_AMD64_DTPMOD64  = 16 // TLS DTPMOD64 relocation
	R_AMD64_DTPOFF64  = 17 // TLS DTPOFF64 relocation
)

// ElfRelocation represents a single relocation entry in an ELF file.
type ElfRelocation struct {
	Offset uintptr // Offset at which the relocation should be applied
	Type   uint32  // Type of relocation
	Symbol uint32  // Index of the symbol to which the relocation refers
	Addend uintptr // Value to be added to the symbol's address
}

// ElfRelocationTable holds a list of ElfRelocation entries.
type ElfRelocationTable struct {
	Relocations []ElfRelocation
}

// NewRelocationTable creates a new ElfRelocationTable by parsing the relocation table from an ELF file.
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
		rOffset := uintptr(binary.LittleEndian.Uint64(data[relOffset:]))
		rInfo := binary.LittleEndian.Uint64(data[relOffset+8:])
		rType := uint32(rInfo & 0xFFFFFFFF) // Lower 32 bits for type
		rSym := uint32(rInfo >> 32)         // Upper 32 bits for symbol index
		rAddend := uintptr(binary.LittleEndian.Uint64(data[relOffset+16:]))

		table.Relocations = append(table.Relocations, ElfRelocation{
			Offset: rOffset,
			Type:   rType,
			Symbol: rSym,
			Addend: rAddend,
		})
	}

	return table
}
