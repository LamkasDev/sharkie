package elf

import (
	"encoding/binary"
	"hash/fnv"
	"io"

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
	STB_LOCAL  = 0 // Local symbol
	STB_GLOBAL = 1 // Global symbol
	STB_WEAK   = 2 // Weak symbol
)

// SymbolHasher is a FNV-64a hasher used for generating hash indexes for symbols.
var SymbolHasher = fnv.New64a()

// ElfSymbolTable represents the symbol table of an ELF file.
type ElfSymbolTable struct {
	Symbols    []*ElfSymbol          // List of all symbols
	SymbolsMap map[uint64]*ElfSymbol // Map of symbols indexed by their hash
}

// ElfSymbol represents a single symbol entry in the ELF symbol table.
type ElfSymbol struct {
	HashIndex    uint64  // Hash index for the symbol
	OriginalName string  // Original name of the symbol
	ReadableName string  // Human-readable name of the symbol
	NidBase      string  // Clean 11-character NID base
	Address      uintptr // Virtual address of the symbol
	Type         uint8   // Type of the symbol
	Binding      uint8   // Binding of the symbol

	ModuleIndex  uint16 // Index of the module the symbol belongs to
	LibraryIndex uint16 // Index of the library the symbol belongs to
	LibraryName  string // Name of the library the symbol belongs to
}

// ResolveSymbolInfo resolves the library name, module index, and readable name for a given symbol.
// It handles both NID (Name ID) encoded symbols and symbols with direct section indices.
func (e *Elf) ResolveSymbolInfo(s *ElfSymbol, stShndx uint16) {
	if s.Type == STT_SECTION {
		return
	}
	s.NidBase = symbol.ExtractNidBase(s.OriginalName)

	// A NID string with indices is minimum 15 chars: [11 NID] [1 sep] [1 lib] [1 sep] [1 mod].
	if len(s.OriginalName) >= 15 {
		sep1 := s.OriginalName[11]
		sep2 := s.OriginalName[13]
		if sep1 == '#' && sep2 == '#' {
			libIDEncoded := s.OriginalName[12]
			modIDEncoded := s.OriginalName[14]

			s.LibraryIndex = symbol.DecodeNidChar(libIDEncoded)
			s.ModuleIndex = symbol.DecodeNidChar(modIDEncoded)
		} else if stShndx != 0 {
			s.LibraryIndex = stShndx
		}
	} else if stShndx != 0 {
		// For non-NID imports/exports, stShndx is the direct index.
		s.LibraryIndex = stShndx
	}
	if s.Address == 0 {
		if lib := e.DynamicInfo.ImportLibraries[s.LibraryIndex]; lib != nil {
			s.LibraryName = lib.Name
		}
	} else {
		if lib := e.DynamicInfo.ExportLibraries[s.LibraryIndex]; lib != nil {
			s.LibraryName = lib.Name
		}
	}
	s.HashIndex = GetSymbolHashIndex(s.LibraryName, s.ReadableName)
}

// GetSymbolHashIndex generates a unique hash for a symbol based on its library name and symbol name.
func GetSymbolHashIndex(libraryName, symbolName string) uint64 {
	SymbolHasher.Reset()
	io.WriteString(SymbolHasher, libraryName)
	SymbolHasher.Write([]byte{0})
	io.WriteString(SymbolHasher, symbolName)

	return SymbolHasher.Sum64()
}

// NewSymbolTable parses the symbol table from the ELF data based on information in the dynamic section.
// It iterates through symbol entries, extracts their properties and resolves additional information.
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
			ReadableName: symbol.MangledToReadable(name),
			Address:      stValue,
			Type:         stInfo & 0xf, // Lower 4 bits for type
			Binding:      stInfo >> 4,  // Upper 4 bits for binding
		}
		e.ResolveSymbolInfo(symbol, stShndx)
		symbolTable.RegisterSymbol(symbol)
	}

	return symbolTable
}

// RegisterSymbol adds a given ElfSymbol to the symbol table's slice and map.
func (st *ElfSymbolTable) RegisterSymbol(s *ElfSymbol) {
	st.Symbols = append(st.Symbols, s)
	st.SymbolsMap[s.HashIndex] = s
}
