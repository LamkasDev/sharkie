// Package elf handles ELF (Executable Linkable File) parsing.
package elf

import (
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/sys_struct"
	"github.com/gookit/color"
)

// Program header types as defined in ELF specification.
const (
	PT_LOAD           = 1          // Loadable segment
	PT_DYNAMIC        = 2          // Dynamic linking information
	PT_TLS            = 7          // Thread-local storage segment
	PT_SCE_DYNLIBDATA = 0x61000000 // SCE-specific dynamic library data
	PT_SCE_PROCPARAM  = 0x61000001 // SCE-specific process parameters
	PT_GNU_EH_FRAME   = 0x6474e550 // GNU exception handling frame
)

// Exception frame encoding constants.
const (
	EFRAME_PCREL  = 0x10 // PC-relative encoding
	EFRAME_SDATA4 = 0x0B // Signed 4-byte data
)

// Elf represents a parsed ELF (Executable and Linkable Format) file.
type Elf struct {
	ModuleIndex uint64
	Name        string
	Path        string
	Linked      bool

	BaseAddress  uintptr
	EntryAddress uint64
	Memory       []byte

	MemSize                   uint64              // Bytes of memory allocated for the ELF
	DynLibDataOffset          uint64              // Offset to dynamic library data
	LoadSections              []*ElfLoadSection   // List of loadable sections
	ExceptionFrameSection     *ElfLoadSection     // Section containing exception handling frames
	ExceptionFrameDataAddress uintptr             // Address of exception frame data
	ExceptionFrameDataSize    uint64              // Size of exception frame data
	ProcessParamSection       *ElfLoadSection     // Section containing process parameters
	TlsSection                *ElfTlsSection      // Section containing thread-local storage information
	DynamicInfo               *ElfDynamicSection  // Dynamic linking information
	SymbolTable               *ElfSymbolTable     // Symbol table
	RelaRelocationTable       *ElfRelocationTable // Rela relocation table
	PltRelocationTable        *ElfRelocationTable // PLT relocation table

	// Temporary.
	CallerToFunctionName map[uintptr]uint64
}

// NewElf creates a new instance of Elf by parsing the provided file data.
func NewElf(data []byte) *Elf {
	e := &Elf{
		LoadSections:         []*ElfLoadSection{},
		CallerToFunctionName: map[uintptr]uint64{},
	}

	// Check magic of the file.
	if string(data[0:4]) != "\x7fELF" {
		panic("Not a valid ELF file")
	}

	e.EntryAddress = binary.LittleEndian.Uint64(data[0x18:])
	phOff := int(binary.LittleEndian.Uint64(data[0x20:]))
	phEntSize := int(binary.LittleEndian.Uint16(data[0x36:]))
	phNum := int(binary.LittleEndian.Uint16(data[0x38:]))

	// Iterate over sections and process independent ones.
	for i := range phNum {
		offset := phOff + i*phEntSize
		pType := binary.LittleEndian.Uint32(data[offset:])
		switch pType {
		case PT_LOAD:
			loadSection := e.NewLoadSection(data, uint64(offset))
			size := loadSection.PVaddr + loadSection.PMemsz
			if size > e.MemSize {
				e.MemSize = size
			}
			e.LoadSections = append(e.LoadSections, loadSection)
		case PT_SCE_DYNLIBDATA:
			e.DynLibDataOffset = binary.LittleEndian.Uint64(data[offset+0x08:])
		case PT_TLS:
			e.TlsSection = e.NewTlsSection(data, uint64(offset))
		case PT_DYNAMIC:
			_ = 0
		case PT_GNU_EH_FRAME:
			e.ExceptionFrameSection = e.NewLoadSection(data, uint64(offset))
		case PT_SCE_PROCPARAM:
			e.ProcessParamSection = e.NewLoadSection(data, uint64(offset))
		default:
			logger.Print(color.Gray.Sprintf("  Unhandled ELF section type %d.\n", pType))
		}
	}

	// We need to make sure PT_SCE_DYNLIBDATA was loaded first.
	for i := range phNum {
		offset := phOff + i*phEntSize
		pType := binary.LittleEndian.Uint32(data[offset:])
		switch pType {
		case PT_DYNAMIC:
			dynamicOffset := binary.LittleEndian.Uint64(data[offset+0x08:])
			dynamicSize := binary.LittleEndian.Uint64(data[offset+0x20:])
			e.DynamicInfo = e.NewDynamicSection(data, dynamicOffset, dynamicSize)
		}
	}

	// Load relocation tables.
	e.RelaRelocationTable = NewRelocationTable(data, e.DynamicInfo.RelaOffset, e.DynamicInfo.RelaSize, e.DynamicInfo.RelaEnt)
	if e.DynamicInfo.PltRelEnt == 0 {
		e.DynamicInfo.PltRelEnt = e.DynamicInfo.RelaEnt
	}
	e.PltRelocationTable = NewRelocationTable(data, e.DynamicInfo.PltRelOffset, e.DynamicInfo.PltRelSize, e.DynamicInfo.PltRelEnt)
	e.SymbolTable = e.NewSymbolTable(data)

	// Allocate memory and load sections.
	e.BaseAddress, _ = sys_struct.AllocExecutableMemory(uintptr(e.MemSize))
	e.Memory = unsafe.Slice((*byte)(unsafe.Pointer(e.BaseAddress)), e.MemSize)
	logger.Printf(
		"PT_LOAD data loaded into memory at %s (%s bytes).\n",
		color.Yellow.Sprintf("0x%X", e.BaseAddress),
		color.Gray.Sprintf("%d", len(e.Memory)),
	)

	for _, loadSection := range e.LoadSections {
		ProcessLoadSection(e, loadSection, data)
	}
	if e.ProcessParamSection != nil {
		ProcessLoadSection(e, e.ProcessParamSection, data)
	}
	if e.ExceptionFrameSection != nil {
		ProcessExceptionFrameSection(e)
	}

	logger.Printf(
		"Loaded module with %s imports & %s exports.\n",
		color.Yellow.Sprintf("%d", e.DynamicInfo.ImportLibrariesCount),
		color.Yellow.Sprintf("%d", e.DynamicInfo.ExportLibrariesCount),
	)

	return e
}

// GetAlignedSize returns memsz aligned on align boundary.
// It calculates the smallest multiple of 'align' that is greater than or equal to 'memsz'.
func GetAlignedSize(memsz uint64, align uint64) uint64 {
	if align > 0 {
		return (memsz + (align - 1)) & ^(align - 1)
	}
	return memsz
}

// ReadInt32 reads an int32 at address belonging to the module (very silly).
func (e *Elf) ReadInt32(addr uintptr) int32 {
	offset := addr - e.BaseAddress
	return int32(binary.LittleEndian.Uint32(e.Memory[offset : offset+4]))
}

// ReadInt64 reads an int64 at address belonging to the module (very silly).
func (e *Elf) ReadInt64(addr uintptr) int64 {
	offset := addr - e.BaseAddress
	return int64(binary.LittleEndian.Uint64(e.Memory[offset : offset+8]))
}
