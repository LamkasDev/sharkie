package elf

import (
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/LamkasDev/sharkie/cmd/sys_struct"
	"github.com/gookit/color"
)

// https://github.com/shadps4-emu/shadPS4/blob/9e287564ced1c7d84a5a165ce4ad6ba85d561ee1/src/core/loader/elf.h#L264
const (
	PT_LOAD           = 1
	PT_DYNAMIC        = 2
	PT_TLS            = 7
	PT_SCE_DYNLIBDATA = 0x61000000
	PT_SCE_PROCPARAM  = 0x61000001
	PT_GNU_EH_FRAME   = 0x6474e550
)

const (
	EFRAME_PCREL  = 0x10
	EFRAME_SDATA4 = 0x0B
)

type Elf struct {
	ModuleIndex  uint64
	Name         string
	BaseAddress  uintptr
	EntryAddress uint64
	Memory       []byte

	MemSize                   uint64
	DynLibDataOffset          uint64
	LoadSections              []*ElfLoadSection
	ExceptionFrameSection     *ElfLoadSection
	ExceptionFrameDataAddress uintptr
	ExceptionFrameDataSize    uint64
	ProcessParamSection       *ElfLoadSection
	TlsSection                *ElfTlsSection
	DynamicInfo               *ElfDynamicSection
	SymbolTable               *ElfSymbolTable
	RelaRelocationTable       *ElfRelocationTable
	PltRelocationTable        *ElfRelocationTable

	// Temporary, used for mapping generic stub callers.
	CallerToFunctionName map[uintptr]uint64
	Path                 string
	Linked               bool
}

// NewElf creates a new instance of Elf based on file contents.
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
	for i := 0; i < phNum; i++ {
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
			break
		case PT_SCE_DYNLIBDATA:
			e.DynLibDataOffset = binary.LittleEndian.Uint64(data[offset+0x08:])
			break
		case PT_TLS:
			e.TlsSection = e.NewTlsSection(data, uint64(offset))
			break
		case PT_DYNAMIC:
			break
		case PT_GNU_EH_FRAME:
			e.ExceptionFrameSection = e.NewLoadSection(data, uint64(offset))
			break
		case PT_SCE_PROCPARAM:
			e.ProcessParamSection = e.NewLoadSection(data, uint64(offset))
			break
		default:
			logger.Print(color.Gray.Sprintf("  Unhandled ELF section type %d.\n", pType))
			break
		}
	}

	// We need to make sure PT_SCE_DYNLIBDATA was loaded first.
	for i := 0; i < phNum; i++ {
		offset := phOff + i*phEntSize
		pType := binary.LittleEndian.Uint32(data[offset:])
		switch pType {
		case PT_DYNAMIC:
			dynamicOffset := binary.LittleEndian.Uint64(data[offset+0x08:])
			dynamicSize := binary.LittleEndian.Uint64(data[offset+0x20:])
			e.DynamicInfo = e.NewDynamicSection(data, dynamicOffset, dynamicSize)
			break
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
		// Now we need to actually parse the section and figure out the exception frame address.
		headerAddr := e.BaseAddress + uintptr(e.ExceptionFrameSection.PVaddr)
		memOffset := uintptr(e.ExceptionFrameSection.PVaddr)

		e.ExceptionFrameSection.Address = headerAddr
		e.ExceptionFrameSection.LoadedSize = e.ExceptionFrameSection.PMemsz

		// Ensure we can read the header.
		if uint64(memOffset+8) <= e.MemSize {
			encoding := e.Memory[memOffset+1]
			switch encoding {
			case EFRAME_PCREL | EFRAME_SDATA4:
				relOffset := int32(binary.LittleEndian.Uint32(e.Memory[memOffset+4:]))
				dataAddr := uintptr(int64(headerAddr) + 4 + int64(relOffset))
				e.ExceptionFrameDataAddress = dataAddr

				// Not sure how big it is, really. Let's just let it run until 0.
				if dataAddr >= e.BaseAddress {
					offset := uint64(dataAddr - e.BaseAddress)
					if offset < e.MemSize {
						e.ExceptionFrameDataSize = e.MemSize - offset
					}
				}

				logger.Printf("Resolved %s data via header (headerAddr=%s, dataAddr=%s, size=%s).\n",
					color.Blue.Sprint(".eh_frame"),
					color.Yellow.Sprintf("0x%X", headerAddr),
					color.Yellow.Sprintf("0x%X", dataAddr),
					color.Green.Sprint(e.ExceptionFrameDataSize),
				)
				break
			default:
				logger.Print(color.Gray.Sprintf(
					"Unknown .eh_frame_hdr encoding 0x%X, assuming data follows header.\n",
					encoding,
				))
				e.ExceptionFrameDataAddress = headerAddr + uintptr(e.ExceptionFrameSection.PMemsz)
				break
			}
		}
	}

	logger.Printf(
		"Loaded module with %s imports & %s exports.\n",
		color.Yellow.Sprintf("%d", e.DynamicInfo.ImportLibrariesCount),
		color.Yellow.Sprintf("%d", e.DynamicInfo.ExportLibrariesCount),
	)

	return e
}

// GetAlignedSize returns memsz aligned on align boundary.
// https://github.com/shadps4-emu/shadPS4/blob/9e287564ced1c7d84a5a165ce4ad6ba85d561ee1/src/core/module.cpp#L24
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
