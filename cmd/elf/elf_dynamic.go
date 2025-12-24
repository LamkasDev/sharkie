package elf

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

const (
	// Regular ELF dynamic tags
	// https://github.com/shadps4-emu/shadPS4/blob/9e287564ced1c7d84a5a165ce4ad6ba85d561ee1/src/core/loader/elf.h#L347
	DT_NEEDED          = 1
	DT_PLTRELSZ        = 2
	DT_STRTAB          = 5
	DT_SYMTAB          = 6
	DT_RELA            = 7
	DT_RELASZ          = 8
	DT_RELAENT         = 9
	DT_STRSZ           = 10
	DT_SYMENT          = 11
	DT_INIT            = 0x0000000c
	DT_DEBUG           = 21
	DT_TEXTREL         = 22
	DT_PLTREL          = 20
	DT_JMPREL          = 23
	DT_INIT_ARRAY      = 0x00000019
	DT_INIT_ARRAYSZ    = 0x0000001b
	DT_FLAGS           = 30
	DT_PREINIT_ARRAY   = 0x00000020
	DT_PREINIT_ARRAYSZ = 0x00000021

	// Playstation specific dynamic tags
	// https://github.com/OpenOrbis/OpenOrbis-PS4-Toolchain/wiki/PS4-ELF-Specification---Dynlib-Data
	// https://github.com/shadps4-emu/shadPS4/blob/9e287564ced1c7d84a5a165ce4ad6ba85d561ee1/src/core/loader/elf.h#L347
	DT_SCE_FINGERPRINT   = 0x61000007
	DT_SCE_FILENAME      = 0x61000009
	DT_SCE_MODULE_ATTR   = 0x6100000B
	DT_SCE_MODULE_INFO   = 0x6100000D
	DT_SCE_NEEDED_MODULE = 0x6100000F
	DT_SCE_EXPORT_LIB    = 0x61000013
	DT_SCE_IMPORT_LIB    = 0x61000015
	DT_SCE_HASH          = 0x61000025
	DT_SCE_PLTGOT        = 0x61000027
	DT_SCE_JMPREL        = 0x61000029
	DT_SCE_PLTREL        = 0x6100002B
	DT_SCE_PLTRELSZ      = 0x6100002D
	DT_SCE_RELA          = 0x6100002F
	DT_SCE_RELASZ        = 0x61000031
	DT_SCE_RELAENT       = 0x61000033
	DT_SCE_STRTAB        = 0x61000035
	DT_SCE_STRSZ         = 0x61000037
	DT_SCE_SYMTAB        = 0x61000039
	DT_SCE_SYMENT        = 0x6100003B
	DT_SCE_HASHSZ        = 0x6100003D
)

type ElfLibrary struct {
	Name         string
	LibraryIndex uint16
}

type ElfModule struct {
	Name        string
	ModuleIndex uint16
}

type ElfDynamicSection struct {
	RelaOffset, RelaSize, RelaEnt       uint64
	PltRelOffset, PltRelSize, PltRelEnt uint64
	SymTabOffset, SymEnt                uint64
	HashOffset, HashSize                uint64

	Needed               []string
	ImportModules        []ElfModule
	ImportModulesCount   uint16
	ImportLibraries      []ElfLibrary
	ImportLibrariesCount uint16
	ExportLibraries      []ElfLibrary
	ExportLibrariesCount uint16
	StringTable          map[uint64]string

	InitFuncOffset                       *uint64
	InitArrayOffset, InitArraySize       uint64
	PreInitArrayOffset, PreInitArraySize uint64
	InitFunc                             *uint64
	InitArray                            []uint64
	PreInitArray                         []uint64
}

// NewDynamicSection loads the ELF dynamic section of dynSize starting at dynOffset.
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
			break
		case DT_STRSZ, DT_SCE_STRSZ:
			strTabSize = value
			break
		case DT_INIT:
			section.InitFuncOffset = &value
			break
		case DT_INIT_ARRAY:
			section.InitArrayOffset = value
			break
		case DT_INIT_ARRAYSZ:
			section.InitArraySize = value
			break
		case DT_PREINIT_ARRAY:
			section.PreInitArrayOffset = value
			break
		case DT_PREINIT_ARRAYSZ:
			section.PreInitArraySize = value
			break
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
			break
		case DT_SCE_NEEDED_MODULE:
			moduleIndex := value >> 48
			nameOffset := value & 0xFFF
			moduleName := section.StringTable[nameOffset]

			section.ImportModules[moduleIndex] = ElfModule{
				Name:        moduleName,
				ModuleIndex: uint16(moduleIndex),
			}
			section.ImportModulesCount++
			break
		case DT_SCE_IMPORT_LIB:
			libraryIndex := value >> 48
			nameOffset := value & 0xFFF
			libraryName := section.StringTable[nameOffset]

			section.ImportLibraries[libraryIndex] = ElfLibrary{
				Name:         libraryName,
				LibraryIndex: uint16(libraryIndex),
			}
			section.ImportLibrariesCount++
			break
		case DT_SCE_EXPORT_LIB:
			libraryIndex := value >> 48
			nameOffset := value & 0xFFF
			libraryName := section.StringTable[nameOffset]

			section.ExportLibraries[libraryIndex] = ElfLibrary{
				Name:         libraryName,
				LibraryIndex: uint16(libraryIndex),
			}
			section.ExportLibrariesCount++
			break
		case DT_SCE_MODULE_INFO:
			nameOffset := value & 0xFFF
			moduleName := section.StringTable[nameOffset]

			section.ImportModules[0] = ElfModule{
				Name:        moduleName,
				ModuleIndex: 0,
			}
			e.Name = fmt.Sprintf("%s.sprx", moduleName)
			break
		case DT_RELA, DT_SCE_RELA:
			section.RelaOffset = value
			break
		case DT_RELASZ, DT_SCE_RELASZ:
			section.RelaSize = value
			break
		case DT_RELAENT, DT_SCE_RELAENT:
			section.RelaEnt = value
			break
		case DT_JMPREL, DT_SCE_JMPREL:
			section.PltRelOffset = value
			break
		case DT_PLTRELSZ, DT_SCE_PLTRELSZ:
			section.PltRelSize = value
			break
		case DT_PLTREL, DT_SCE_PLTREL:
			if value != DT_RELA {
				logger.Print(color.Gray.Sprintf("  DT_SCE_PLTREL doesn't match DT_RELA!\n"))
			}
			break
		case DT_SYMTAB, DT_SCE_SYMTAB:
			section.SymTabOffset = value
			break
		case DT_SYMENT, DT_SCE_SYMENT:
			section.SymEnt = value
			break
		case DT_SCE_HASH:
			section.HashOffset = value
			break
		case DT_SCE_HASHSZ:
			section.HashSize = value
			break
		}
	}

	section.RelaOffset += e.DynLibDataOffset
	section.PltRelOffset += e.DynLibDataOffset
	section.SymTabOffset += e.DynLibDataOffset
	section.HashOffset += e.DynLibDataOffset

	return section
}
