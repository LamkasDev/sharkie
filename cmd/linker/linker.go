package linker

import (
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/asm"
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/mem"
	"github.com/gookit/color"
)

var GlobalLinker = NewLinker()

// Linker keeps track of linking state.
type Linker struct {
	GenerationCounter uintptr
	MaxTlsIndex       uintptr
	StaticTlsSize     uint64
}

// NewLinker creates a new instance of Linker.
func NewLinker() *Linker {
	return &Linker{
		GenerationCounter: 1,
	}
}

// Link loads the ELF file into memory and performs relocations.
func (l *Linker) Link(e *elf.Elf, data []byte) {
	e.BaseAddress = mem.AllocExecututableMemory(uintptr(e.MemSize))
	e.Memory = unsafe.Slice((*byte)(unsafe.Pointer(e.BaseAddress)), e.MemSize)
	fmt.Printf(
		"PT_LOAD data loaded into memory at %s (%s bytes).\n",
		color.Yellow.Sprintf("0x%X", e.BaseAddress),
		color.Gray.Sprintf("%d", len(e.Memory)),
	)

	for _, loadSection := range e.LoadSections {
		ProcessLoadSection(e, loadSection, data)
	}

	if e.TlsSection != nil {
		l.GenerationCounter++
		l.MaxTlsIndex++
		e.TlsSection.ModuleIndex = uint64(l.MaxTlsIndex)
		e.TlsSection.Offset = l.StaticTlsSize
		l.StaticTlsSize += e.TlsSection.ImageSize
	}

	if e.DynamicInfo != nil {
		ProcessRelocations(e)
	} else {
		fmt.Println("Dynamic section size is 0, skipping relocations...")
	}

	for _, symbol := range e.SymbolTable.Symbols {
		if symbol.Address == 0 || symbol.Type != elf.STT_FUNC {
			continue
		}
		stub, ok := asm.Stubs[symbol.HashIndex]
		if ok {
			// MOV trampolineAddr, RAX
			patch := []byte{0x48, 0xB8}
			patch = binary.LittleEndian.AppendUint64(patch, uint64(stub.Address))

			// JMP RAX
			patch = append(patch, 0xFF, 0xE0)

			copy(e.Memory[symbol.Address:], patch)
			fmt.Printf(
				"Patched stubbed symbol %s inside %s at %s.\n",
				color.Blue.Sprintf("%s:%s", symbol.LibraryName, symbol.ReadableName),
				color.Blue.Sprintf(e.Name),
				color.Yellow.Sprintf("0x%X", symbol.Address),
			)
		}
	}

	if e.DynamicInfo.InitFuncOffset != nil {
		funcPtr := uint64(e.BaseAddress) + *e.DynamicInfo.InitFuncOffset
		e.DynamicInfo.InitFunc = &funcPtr
	}

	if e.DynamicInfo.InitArraySize > 0 {
		initArrayStart := e.DynamicInfo.InitArrayOffset
		initArrayData := e.Memory[initArrayStart : initArrayStart+e.DynamicInfo.InitArraySize]
		offset := 0
		for offset < len(initArrayData) {
			funcPtr := binary.LittleEndian.Uint64(initArrayData[offset:])
			e.DynamicInfo.InitArray = append(e.DynamicInfo.InitArray, funcPtr)
			offset += 8
		}
	}

	if e.DynamicInfo.PreInitArraySize > 0 {
		preInitArrayStart := e.DynamicInfo.PreInitArrayOffset
		preInitArrayData := e.Memory[preInitArrayStart : preInitArrayStart+e.DynamicInfo.PreInitArraySize]
		offset := 0
		for offset < len(preInitArrayData) {
			funcPtr := binary.LittleEndian.Uint64(preInitArrayData[offset:])
			e.DynamicInfo.PreInitArray = append(e.DynamicInfo.PreInitArray, funcPtr)
			offset += 8
		}
	}
}
