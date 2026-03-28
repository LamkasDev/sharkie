package linker

import (
	"encoding/binary"

	"github.com/LamkasDev/sharkie/cmd/asm"
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/gookit/color"
)

var GlobalLinker = NewLinker()

// Linker keeps track of linking state.
type Linker struct {
	GenerationCounter uintptr
	StaticTlsSize     uint64
}

// NewLinker creates a new instance of Linker.
func NewLinker() *Linker {
	return &Linker{
		GenerationCounter: 1,
	}
}

// Link performs relocations and some patches.
func (l *Linker) Link(e *elf.Elf) error {
	if e.TlsSection != nil {
		l.GenerationCounter++
		e.TlsSection.Offset = l.StaticTlsSize
		l.StaticTlsSize += e.TlsSection.ImageSize
	}

	if e.DynamicInfo != nil {
		ProcessRelocations(e)
	} else {
		logger.Print(color.Gray.Sprintf("Dynamic section size is 0, skipping relocations..."))
	}

	// HACK: we need to stub these symbol, but they're private.
	if e.Name == "libkernel.sprx" {
		e.SymbolTable.RegisterSymbol(&elf.ElfSymbol{
			HashIndex:    elf.GetSymbolHashIndex("libkernel", "sub_1590"),
			LibraryName:  "libkernel",
			ReadableName: "sub_1590",
			Address:      0x0000000000001590,
			Type:         elf.STT_FUNC,
			Binding:      elf.STB_LOCAL,
		})
		e.SymbolTable.RegisterSymbol(&elf.ElfSymbol{
			HashIndex:    elf.GetSymbolHashIndex("libkernel", "sub_1D90"),
			LibraryName:  "libkernel",
			ReadableName: "sub_1D90",
			Address:      0x0000000000001D90,
			Type:         elf.STT_FUNC,
			Binding:      elf.STB_LOCAL,
		})
		e.SymbolTable.RegisterSymbol(&elf.ElfSymbol{
			HashIndex:    elf.GetSymbolHashIndex("libkernel", "sub_2BA0"),
			LibraryName:  "libkernel",
			ReadableName: "sub_2BA0",
			Address:      0x0000000000002BA0,
			Type:         elf.STT_FUNC,
			Binding:      elf.STB_LOCAL,
		})
	}

	// Patch a module's own symbols to redirect to stubs.
	for _, symbol := range e.SymbolTable.Symbols {
		if symbol.Address == 0 || symbol.Type != elf.STT_FUNC {
			continue
		}
		stub, ok := asm.Stubs[symbol.HashIndex]
		if ok && stub.SymbolName != "sceFiosInitialize" {
			// MOV trampolineAddr, RAX
			patch := []byte{0x48, 0xB8}
			patch = binary.LittleEndian.AppendUint64(patch, uint64(stub.Address))

			// JMP RAX
			patch = append(patch, 0xFF, 0xE0)

			copy(e.Memory[symbol.Address:], patch)
			/* logger.Printf(
				"Replaced stubbed symbol %s inside %s at %s.\n",
				color.Blue.Sprintf("%s:%s", symbol.LibraryName, symbol.ReadableName),
				color.Blue.Sprintf(e.Name),
				color.Yellow.Sprintf("0x%X", symbol.Address),
			) */
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

	return nil
}
