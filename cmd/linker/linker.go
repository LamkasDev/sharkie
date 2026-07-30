package linker

import (
	"encoding/binary"
	"unsafe"

	"github.com/LamkasDev/sharkie/cmd/asm"
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/lib_structs/tcb"
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
		l.StaticTlsSize = (l.StaticTlsSize + tcb.TcbAlignment - 1) &^ (tcb.TcbAlignment - 1)
		l.StaticTlsSize += e.TlsSection.ImageSize
		e.TlsSection.Offset = l.StaticTlsSize
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
	if e.Name == "Minecraft.Client.sprx" {
		e.SymbolTable.RegisterSymbol(&elf.ElfSymbol{
			HashIndex:    elf.GetSymbolHashIndex("Minecraft.Client.sprx", "sub_1D4CF0"),
			LibraryName:  "Minecraft.Client.sprx",
			ReadableName: "sub_1D4CF0",
			Address:      0x00000000001D4CF0,
			Type:         elf.STT_FUNC,
			Binding:      elf.STB_LOCAL,
		})
		e.SymbolTable.RegisterSymbol(&elf.ElfSymbol{
			HashIndex:    elf.GetSymbolHashIndex("Minecraft.Client.sprx", "sub_20280"),
			LibraryName:  "Minecraft.Client.sprx",
			ReadableName: "sub_20280",
			Address:      0x0000000000020280,
			Type:         elf.STT_FUNC,
			Binding:      elf.STB_LOCAL,
		})
		e.SymbolTable.RegisterSymbol(&elf.ElfSymbol{
			HashIndex:    elf.GetSymbolHashIndex("Minecraft.Client.sprx", "sub_7510F0"),
			LibraryName:  "Minecraft.Client.sprx",
			ReadableName: "sub_7510F0",
			Address:      0x00000000007510F0,
			Type:         elf.STT_FUNC,
			Binding:      elf.STB_LOCAL,
		})
		e.SymbolTable.RegisterSymbol(&elf.ElfSymbol{
			HashIndex:    elf.GetSymbolHashIndex("Minecraft.Client.sprx", "sub_133DD0"),
			LibraryName:  "Minecraft.Client.sprx",
			ReadableName: "sub_133DD0",
			Address:      0x0000000000133DD0,
			Type:         elf.STT_FUNC,
			Binding:      elf.STB_LOCAL,
		})
		e.SymbolTable.RegisterSymbol(&elf.ElfSymbol{
			HashIndex:    elf.GetSymbolHashIndex("PS4Player_Il2cpp.sprx", "sub_840240"),
			LibraryName:  "PS4Player_Il2cpp.sprx",
			ReadableName: "sub_840240",
			Address:      0x0000000000840240,
			Type:         elf.STT_FUNC,
			Binding:      elf.STB_LOCAL,
		})

		if logger.GameDebugMode {
			// We would have to put function pointers in this + 0xA0/0xA8.
			// iggyDebugFlag := unsafe.Slice((*uint64)(unsafe.Add(unsafe.Pointer(e.BaseAddress), 0xA214C0)), 1)
			// iggyDebugFlag[0] = 0xFFFFFFFFFFFFFFFF

			ailDebugFlag1 := unsafe.Slice((*uint32)(unsafe.Add(unsafe.Pointer(e.BaseAddress), 0xA214D0)), 1)
			ailDebugFlag1[0] = 0xFFFFFFFF

			ailDebugFlag2 := unsafe.Slice((*uint32)(unsafe.Add(unsafe.Pointer(e.BaseAddress), 0xA214D4)), 1)
			ailDebugFlag2[0] = 0xFFFFFFFF
		}
	}
	if logger.FiosDebugMode && e.Name == "libSceFios2.sprx" {
		debugFlags := unsafe.Slice((*byte)(unsafe.Add(unsafe.Pointer(e.BaseAddress), 0x17C520)), 4)
		// debugFlags := unsafe.Slice((*byte)(unsafe.Add(unsafe.Pointer(e.BaseAddress), 0x178520)), 4)
		debugFlags[0] = 0xFF
		debugFlags[1] = 0xFF
		debugFlags[2] = 0xFF
		debugFlags[3] = 0xFF
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
