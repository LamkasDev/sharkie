package linker

import (
	"encoding/binary"
	"fmt"

	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/gookit/color"
)

// ProcessRelocations processes all relocation tables.
func ProcessRelocations(e *elf.Elf) {
	ProcessRelocationTable(e, e.RelaRelocationTable, "DT_RELA")
	ProcessRelocationTable(e, e.PltRelocationTable, "DT_JMPREL")
}

// ProcessRelocationTable processes a relocation table.
func ProcessRelocationTable(e *elf.Elf, table *elf.ElfRelocationTable, tableName string) {
	if table == nil {
		return
	}

	fmt.Printf(
		"Processing %s relocation section (%s entries)...\n",
		color.Blue.Sprint(tableName),
		color.Gray.Sprintf("%d", len(table.Relocations)),
	)

	relativeCount := 0
	externalCount := 0
	for _, r := range table.Relocations {
		switch r.Type {
		case elf.R_AMD64_RELATIVE:
			newAddr := uint64(int64(e.BaseAddress) + r.Addend)
			if r.Offset+8 <= uint64(len(e.Memory)) {
				binary.LittleEndian.PutUint64(e.Memory[r.Offset:], newAddr)
				relativeCount++
			}
			break
		case elf.R_AMD64_64:
			if r.Symbol == 0 {
				newAddr := uint64(int64(e.BaseAddress) + r.Addend)
				if r.Offset+8 <= uint64(len(e.Memory)) {
					binary.LittleEndian.PutUint64(e.Memory[r.Offset:], newAddr)
					relativeCount++
				}
				break
			}
			fallthrough
		case elf.R_AMD64_GLOB_DAT, elf.R_AMD64_JUMP_SLOT:
			if int(r.Symbol) >= len(e.SymbolTable.Symbols) {
				break
			}
			symbol := e.SymbolTable.Symbols[r.Symbol]
			if addr, ok := elf.GetSymbolAddress(symbol); ok {
				e.CallerToFunctionName[r.Offset] = symbol.HashIndex
				newAddr := addr + uint64(r.Addend)
				if r.Offset+8 <= uint64(len(e.Memory)) {
					binary.LittleEndian.PutUint64(e.Memory[r.Offset:], newAddr)
					externalCount++
				}
			} else {
				elf.FakeAddressMap[elf.FakeAddress] = fmt.Sprintf("%s:%s", symbol.LibraryName, symbol.ReadableName)
				newAddr := elf.FakeAddress + uint64(r.Addend)
				if r.Addend != 0 {
					color.Grayf("  Unhandled addend %d.\n", r.Addend)
				}
				if r.Offset+8 <= uint64(len(e.Memory)) {
					binary.LittleEndian.PutUint64(e.Memory[r.Offset:], newAddr)
					externalCount++
				}
				color.Grayf("  Added fake address for %s:%s.\n", symbol.LibraryName, symbol.ReadableName)
				elf.FakeAddress += 8
			}
			break
		case elf.R_AMD64_DTPOFF64:
			var symbolValue uint64
			if r.Symbol != 0 && int(r.Symbol) < len(e.SymbolTable.Symbols) {
				symbol := e.SymbolTable.Symbols[r.Symbol]
				symbolValue = symbol.Address
			}

			newAddr := symbolValue + uint64(r.Addend)
			if r.Offset+8 <= uint64(len(e.Memory)) {
				binary.LittleEndian.PutUint64(e.Memory[r.Offset:], newAddr)
			}
			break
		default:
			color.Grayf("  Unhandled relocation type %d.\n", r.Type)
			break
		}
	}
	fmt.Printf(
		"  Applied %s relative & %s external relocations.\n",
		color.Yellow.Sprintf("%d", relativeCount),
		color.Yellow.Sprintf("%d", externalCount),
	)
}
