package elf

import (
	"encoding/binary"
	"hash/fnv"
	"io"
	"strings"

	"github.com/LamkasDev/sharkie/cmd/symbol"
)

const (
	STT_NOTYPE    = 0
	STT_OBJECT    = 1
	STT_FUNC      = 2
	STT_SECTION   = 3
	STT_FILE      = 4
	STT_TLS       = 6
	STT_GNU_IFUNC = 10
)

const (
	STB_LOCAL  = 0
	STB_GLOBAL = 1
	STB_WEAK   = 2
)

var SymbolHasher = fnv.New64a()

type ElfSymbolTable struct {
	Symbols    []*ElfSymbol
	SymbolsMap map[uint64]*ElfSymbol
}

type ElfSymbol struct {
	HashIndex    uint64
	OriginalName string
	ReadableName string
	Address      uint64
	Type         uint8
	Binding      uint8

	ModuleIndex  uint16
	LibraryIndex uint16
	LibraryName  string
}

// ResolveSymbolInfo resolves the library name, module index, and readable name for a symbol.
func (e *Elf) ResolveSymbolInfo(s *ElfSymbol, stShndx uint16) {
	parts := strings.Split(s.OriginalName, "#")
	if len(parts) >= 3 {
		// For NID imports/exports, indexes are encoded inside characters between #.
		libIDEncoded := parts[1][0]
		modIDEncoded := parts[2][0]
		s.LibraryIndex = symbol.DecodeNidChar(libIDEncoded)
		s.ModuleIndex = symbol.DecodeNidChar(modIDEncoded)
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

func GetSymbolHashIndex(libraryName, symbolName string) uint64 {
	SymbolHasher.Reset()
	io.WriteString(SymbolHasher, libraryName)
	SymbolHasher.Write([]byte{0})
	io.WriteString(SymbolHasher, symbolName)

	return SymbolHasher.Sum64()
}

// NewSymbolTable loads a symbol table based on dynamic section.
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
		stValue := binary.LittleEndian.Uint64(data[symEntryOffset+8:])

		name := e.DynamicInfo.StringTable[uint64(stName)]
		symbol := &ElfSymbol{
			OriginalName: name,
			ReadableName: symbol.MangledToReadable(name),
			Address:      stValue,
			Type:         stInfo & 0xf,
			Binding:      stInfo >> 4,
		}
		e.ResolveSymbolInfo(symbol, stShndx)
		symbolTable.RegisterSymbol(symbol)
	}

	return symbolTable
}

// RegisterSymbol adds a symbol to the symbol table.
func (st *ElfSymbolTable) RegisterSymbol(s *ElfSymbol) {
	st.Symbols = append(st.Symbols, s)
	st.SymbolsMap[s.HashIndex] = s
}
