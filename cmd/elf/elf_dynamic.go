package elf

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

// Dynamic array tags as defined in ELF specification and Playstation extensions.
const (
	// Regular ELF dynamic tags.
	DT_NEEDED   = 1 // Name of a needed library
	DT_PLTRELSZ = 2 // Size of PLT relocation entries
	DT_STRTAB   = 5 // Address of the string table
	DT_SYMTAB   = 6 // Address of the symbol table

	DT_RELA    = 7 // Address of the Rela relocation table
	DT_RELASZ  = 8 // Size of the Rela relocation table
	DT_RELAENT = 9 // Size of a single Rela relocation entry

	DT_STRSZ  = 10 // Size of the string table
	DT_SYMENT = 11 // Size of a single symbol table entry
	DT_INIT   = 12 // Address of the initialization function
	DT_DEBUG  = 21 // Address of debugging information

	DT_TEXTREL = 22
	DT_PLTREL  = 20
	DT_JMPREL  = 23

	DT_FLAGS = 30 // ELF dynamic flags

	DT_INIT_ARRAY      = 0x00000019 // Address of the DT_INIT array
	DT_INIT_ARRAYSZ    = 0x0000001b // Size of the DT_INIT array
	DT_PREINIT_ARRAY   = 0x00000020 // Address of the DT_PREINIT array
	DT_PREINIT_ARRAYSZ = 0x00000021 // Size of the DT_PREINIT array

	// Playstation specific dynamic tags.
	DT_SCE_FINGERPRINT   = 0x61000007 // SCE-specific fingerprint
	DT_SCE_FILENAME      = 0x61000009 // SCE-specific filename
	DT_SCE_MODULE_ATTR   = 0x6100000B // SCE-specific module attributes
	DT_SCE_MODULE_INFO   = 0x6100000D // SCE-specific module information
	DT_SCE_NEEDED_MODULE = 0x6100000F // SCE-specific needed module
	DT_SCE_EXPORT_LIB    = 0x61000013 // SCE-specific export library
	DT_SCE_IMPORT_LIB    = 0x61000015 // SCE-specific import library
	DT_SCE_HASH          = 0x61000025 // SCE-specific hash table
	DT_SCE_PLTGOT        = 0x61000027 // SCE-specific PLT/GOT
	DT_SCE_JMPREL        = 0x61000029 // SCE-specific PLT relocation table
	DT_SCE_PLTREL        = 0x6100002B // SCE-specific PLT relocation type
	DT_SCE_PLTRELSZ      = 0x6100002D // SCE-specific PLT relocation table size
	DT_SCE_RELA          = 0x6100002F // SCE-specific Rela relocation table
	DT_SCE_RELASZ        = 0x61000031 // SCE-specific Rela relocation table size
	DT_SCE_RELAENT       = 0x61000033 // SCE-specific Rela relocation entry size
	DT_SCE_STRTAB        = 0x61000035 // SCE-specific string table
	DT_SCE_STRSZ         = 0x61000037 // SCE-specific string table size
	DT_SCE_SYMTAB        = 0x61000039 // SCE-specific symbol table
	DT_SCE_SYMENT        = 0x6100003B // SCE-specific symbol entry size
	DT_SCE_HASHSZ        = 0x6100003D // SCE-specific hash table size
)

// ElfLibrary represents an imported or exported library in the ELF's dynamic section.
type ElfLibrary struct {
	Name         string
	LibraryIndex uint16
}

// ElfModule represents an imported module in the ELF's dynamic section.
type ElfModule struct {
	Name        string
	ModuleIndex uint16
}

// ElfDynamicSection contains parsed information from the .dynamic section of an ELF file.
type ElfDynamicSection struct {
	RelaOffset, RelaSize, RelaEnt       uint64 // Rela relocation table information
	PltRelOffset, PltRelSize, PltRelEnt uint64 // PLT relocation table information
	SymTabOffset, SymEnt                uint64 // Symbol table information
	HashOffset, HashSize                uint64 // Hash table information

	Needed               []string          // List of needed shared libraries
	ImportModules        []ElfModule       // List of imported modules
	ImportModulesCount   uint16            // Number of imported modules
	ImportLibraries      []ElfLibrary      // List of imported libraries
	ImportLibrariesCount uint16            // Number of imported libraries
	ExportLibraries      []ElfLibrary      // List of exported libraries
	ExportLibrariesCount uint16            // Number of exported libraries
	StringTable          map[uint64]string // String table for dynamic section strings

	InitFuncOffset                       *uint64  // Offset to the initialization function
	InitArrayOffset, InitArraySize       uint64   // Initialization array information
	PreInitArrayOffset, PreInitArraySize uint64   // Pre-initialization array information
	InitFunc                             *uint64  // Address of the initialization function
	InitArray                            []uint64 // Array of constructor functions
	PreInitArray                         []uint64 // Array of pre-constructor functions
}

// NewDynamicSection parses the dynamic section of an ELF file (starting at dynOffset for dynSize bytes).
func (e *Elf) NewDynamicSection(data []byte, dynOffset, dynSize uint64) *ElfDynamicSection {
	if dynSize == 0 {
		return nil
	}

	section := &ElfDynamicSection{
		Needed:          []string{},
		ImportModules:   make([]ElfModule, 64),
		ImportLibraries: make([]ElfLibrary, 64),
		ExportLibraries: make([]ElfLibrary, 64),
		StringTable:     map[uint64]string{},
		InitArray:       []uint64{},
		PreInitArray:    []uint64{},
	}

	// First pass to get string table offset and size.
	var strTabOffset, strTabSize uint64
	for i := uint64(0); i < dynSize; i += 16 {
		tag := binary.LittleEndian.Uint64(data[dynOffset+i:])
		value := binary.LittleEndian.Uint64(data[dynOffset+i+8:])
		switch tag {
		case DT_STRTAB, DT_SCE_STRTAB:
			strTabOffset = value
		case DT_STRSZ, DT_SCE_STRSZ:
			strTabSize = value
		case DT_INIT:
			section.InitFuncOffset = &value
		case DT_INIT_ARRAY:
			section.InitArrayOffset = value
		case DT_INIT_ARRAYSZ:
			section.InitArraySize = value
		case DT_PREINIT_ARRAY:
			section.PreInitArrayOffset = value
		case DT_PREINIT_ARRAYSZ:
			section.PreInitArraySize = value
		}
	}

	// Construct the string table.
	stringTableStart := e.DynLibDataOffset + strTabOffset
	stringTableData := data[stringTableStart : stringTableStart+strTabSize]
	offset := 0
	for offset < len(stringTableData) {
		end := bytes.IndexByte(stringTableData[offset:], 0)
		if end == -1 {
			break
		}
		section.StringTable[uint64(offset)] = string(stringTableData[offset : offset+end])
		offset += end + 1
	}

	// Second pass to process all other tags.
	for i := uint64(0); i < dynSize; i += 16 {
		tag := binary.LittleEndian.Uint64(data[dynOffset+i:])
		value := binary.LittleEndian.Uint64(data[dynOffset+i+8:])
		switch tag {
		case DT_NEEDED:
			name := section.StringTable[value]
			section.Needed = append(section.Needed, name)
		case DT_SCE_NEEDED_MODULE:
			moduleIndex := value >> 48
			nameOffset := value & 0xFFF
			moduleName := section.StringTable[nameOffset]

			section.ImportModules[moduleIndex] = ElfModule{
				Name:        moduleName,
				ModuleIndex: uint16(moduleIndex),
			}
			section.ImportModulesCount++
		case DT_SCE_IMPORT_LIB:
			libraryIndex := value >> 48
			nameOffset := value & 0xFFF
			libraryName := section.StringTable[nameOffset]

			section.ImportLibraries[libraryIndex] = ElfLibrary{
				Name:         libraryName,
				LibraryIndex: uint16(libraryIndex),
			}
			section.ImportLibrariesCount++
		case DT_SCE_EXPORT_LIB:
			libraryIndex := value >> 48
			nameOffset := value & 0xFFF
			libraryName := section.StringTable[nameOffset]

			section.ExportLibraries[libraryIndex] = ElfLibrary{
				Name:         libraryName,
				LibraryIndex: uint16(libraryIndex),
			}
			section.ExportLibrariesCount++
		case DT_SCE_MODULE_INFO:
			nameOffset := value & 0xFFF
			moduleName := section.StringTable[nameOffset]

			section.ImportModules[0] = ElfModule{
				Name:        moduleName,
				ModuleIndex: 0,
			}
			e.Name = fmt.Sprintf("%s.sprx", moduleName)
		case DT_RELA, DT_SCE_RELA:
			section.RelaOffset = value
		case DT_RELASZ, DT_SCE_RELASZ:
			section.RelaSize = value
		case DT_RELAENT, DT_SCE_RELAENT:
			section.RelaEnt = value
		case DT_JMPREL, DT_SCE_JMPREL:
			section.PltRelOffset = value
		case DT_PLTRELSZ, DT_SCE_PLTRELSZ:
			section.PltRelSize = value
		case DT_PLTREL, DT_SCE_PLTREL:
			if value != DT_RELA {
				logger.Print(color.Gray.Sprintf("  DT_SCE_PLTREL doesn't match DT_RELA!\n"))
			}
		case DT_SYMTAB, DT_SCE_SYMTAB:
			section.SymTabOffset = value
		case DT_SYMENT, DT_SCE_SYMENT:
			section.SymEnt = value
		case DT_SCE_HASH:
			section.HashOffset = value
		case DT_SCE_HASHSZ:
			section.HashSize = value
		}
	}

	// Adjust offsets by DynLibDataOffset to get absolute file offsets.
	section.RelaOffset += e.DynLibDataOffset
	section.PltRelOffset += e.DynLibDataOffset
	section.SymTabOffset += e.DynLibDataOffset
	section.HashOffset += e.DynLibDataOffset

	return section
}
