package elf

import (
	"encoding/binary"
	"hash/fnv"
	"io"
	"strings"

	"github.com/LamkasDev/sharkie/cmd/elf_symbol"
)

// Symbol types as defined in ELF specification.
const (
	STT_NOTYPE    = 0  // No type specified
	STT_OBJECT    = 1  // Data object
	STT_FUNC      = 2  // Function
	STT_SECTION   = 3  // Section
	STT_FILE      = 4  // File
	STT_TLS       = 6  // Thread-local storage entity
	STT_GNU_IFUNC = 10 // GNU indirect function
)

// Symbol binding types as defined in ELF specification.
const (
	STB_LOCAL  = 0 // Local elf_symbol
	STB_GLOBAL = 1 // Global elf_symbol
	STB_WEAK   = 2 // Weak elf_symbol
)

// SymbolHasher is a FNV-64a hasher used for generating hash indexes for symbols.
var SymbolHasher = fnv.New64a()

// ElfSymbolTable represents the elf_symbol table of an ELF file.
type ElfSymbolTable struct {
	Symbols    []*ElfSymbol          // List of all symbols
	SymbolsMap map[uint64]*ElfSymbol // Map of symbols indexed by their hash
}

// ElfSymbol represents a single elf_symbol entry in the ELF elf_symbol table.
type ElfSymbol struct {
	HashIndex    uint64  // Hash index for the elf_symbol
	OriginalName string  // Original name of the elf_symbol
	ReadableName string  // Human-readable name of the elf_symbol
	Address      uintptr // Virtual address of the elf_symbol
	Type         uint8   // Type of the elf_symbol
	Binding      uint8   // Binding of the elf_symbol

	ModuleIndex  uint16 // Index of the module the elf_symbol belongs to
	LibraryIndex uint16 // Index of the library the elf_symbol belongs to
	LibraryName  string // Name of the library the elf_symbol belongs to
}

// ResolveSymbolInfo resolves the library name, module index, and readable name for a given elf_symbol.
// It handles both NID (Name ID) encoded symbols and symbols with direct section indices.
func (e *Elf) ResolveSymbolInfo(s *ElfSymbol, stShndx uint16) {
	if s.Type == STT_SECTION {
		return
	}

	parts := strings.Split(s.OriginalName, "#")
	if len(parts) >= 3 {
		// For NID imports/exports, indexes are encoded inside characters between #.
		libIDEncoded := parts[1][0]
		modIDEncoded := parts[2][0]
		s.LibraryIndex = elf_symbol.DecodeNidChar(libIDEncoded)
		s.ModuleIndex = elf_symbol.DecodeNidChar(modIDEncoded)
	} else if stShndx != 0 {
		// For non-NID imports/exports, stShndx is the direct index.
		s.LibraryIndex = stShndx
	}
	if s.Address == 0 {
		if int(s.LibraryIndex) < len(e.DynamicInfo.ImportLibraries) {
			s.LibraryName = e.DynamicInfo.ImportLibraries[s.LibraryIndex].Name
		}
	} else {
		if int(s.LibraryIndex) < len(e.DynamicInfo.ExportLibraries) {
			s.LibraryName = e.DynamicInfo.ExportLibraries[s.LibraryIndex].Name
		}
	}
	s.HashIndex = GetSymbolHashIndex(s.LibraryName, s.ReadableName)
}

// GetSymbolHashIndex generates a unique hash for a elf_symbol based on its library name and elf_symbol name.
func GetSymbolHashIndex(libraryName, symbolName string) uint64 {
	SymbolHasher.Reset()
	io.WriteString(SymbolHasher, libraryName)
	SymbolHasher.Write([]byte{0})
	io.WriteString(SymbolHasher, symbolName)

	return SymbolHasher.Sum64()
}

// NewSymbolTable parses the elf_symbol table from the ELF data based on information in the dynamic section.
// It iterates through elf_symbol entries, extracts their properties and resolves additional information.
func (e *Elf) NewSymbolTable(data []byte) *ElfSymbolTable {
	if e.DynamicInfo.SymEnt == 0 {
		return nil
	}

	numSymbols := binary.LittleEndian.Uint32(data[e.DynamicInfo.HashOffset+4:])
	symbolTable := &ElfSymbolTable{
		Symbols:    []*ElfSymbol{},
		SymbolsMap: map[uint64]*ElfSymbol{},
	}
	for i := uint32(0); i < numSymbols; i++ {
		symEntryOffset := e.DynamicInfo.SymTabOffset + (uint64(i) * e.DynamicInfo.SymEnt)
		if symEntryOffset+24 > uint64(len(data)) {
			break
		}

		stName := binary.LittleEndian.Uint32(data[symEntryOffset:])
		stInfo := data[symEntryOffset+4]
		stShndx := binary.LittleEndian.Uint16(data[symEntryOffset+6:])
		stValue := uintptr(binary.LittleEndian.Uint64(data[symEntryOffset+8:]))

		name := e.DynamicInfo.StringTable[uint64(stName)]
		symbol := &ElfSymbol{
			OriginalName: name,
			ReadableName: elf_symbol.MangledToReadable(name),
			Address:      stValue,
			Type:         stInfo & 0xf, // Lower 4 bits for type
			Binding:      stInfo >> 4,  // Upper 4 bits for binding
		}
		e.ResolveSymbolInfo(symbol, stShndx)
		symbolTable.RegisterSymbol(symbol)
	}

	return symbolTable
}

// RegisterSymbol adds a given ElfSymbol to the elf_symbol table's slice and map.
func (st *ElfSymbolTable) RegisterSymbol(s *ElfSymbol) {
	st.Symbols = append(st.Symbols, s)
	st.SymbolsMap[s.HashIndex] = s
}
