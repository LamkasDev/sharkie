package linker

import (
	"encoding/binary"

	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/logger"
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

	logger.Printf(
		"Processing %s relocation section (%s entries)...\n",
		color.Blue.Sprint(tableName),
		color.Gray.Sprintf("%d", len(table.Relocations)),
	)

	relativeCount := 0
	externalCount := 0
	for _, r := range table.Relocations {
		switch r.Type {
		case elf.R_AMD64_RELATIVE:
			newAddr := e.BaseAddress + r.Addend
			if r.Offset+8 <= uintptr(len(e.Memory)) {
				binary.LittleEndian.PutUint64(e.Memory[r.Offset:], uint64(newAddr))
				relativeCount++
			}
		case elf.R_AMD64_64:
			if r.Symbol == 0 {
				newAddr := e.BaseAddress + r.Addend
				if r.Offset+8 <= uintptr(len(e.Memory)) {
					binary.LittleEndian.PutUint64(e.Memory[r.Offset:], uint64(newAddr))
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
				newAddr := addr + r.Addend
				if r.Offset+8 <= uintptr(len(e.Memory)) {
					e.CallerToFunctionName[r.Offset] = symbol
					binary.LittleEndian.PutUint64(e.Memory[r.Offset:], uint64(newAddr))
					externalCount++
				}
			} else {
				logger.Print(color.Gray.Sprintf("  Skipped fake address for %s:%s.\n", symbol.LibraryName, symbol.ReadableName))
			}
		case elf.R_AMD64_DTPMOD64:
			// TODO: handle symbols outside of current module (rewrite GetSymbolAddress to FindSymbol or smth).
			if int(r.Symbol) >= len(e.SymbolTable.Symbols) {
				break
			}
			symbol := e.SymbolTable.Symbols[r.Symbol]
			moduleIndex := e.ModuleIndex
			if symbol.Type != elf.STT_SECTION && symbol.OriginalName != "" {
				if module := elf.GetDefiningModule(symbol); module != nil {
					moduleIndex = module.ModuleIndex
				} else {
					logger.Print(color.Gray.Sprintf(
						"  Failed finding defining module for %s:%s.\n",
						symbol.LibraryName,
						symbol.ReadableName,
					))
				}
			}
			if r.Offset+8 <= uintptr(len(e.Memory)) {
				binary.LittleEndian.PutUint64(e.Memory[r.Offset:], moduleIndex)
			}
		case elf.R_AMD64_DTPOFF64:
			// TODO: handle symbols outside of current module (rewrite GetSymbolAddress to FindSymbol or smth).
			if int(r.Symbol) >= len(e.SymbolTable.Symbols) {
				break
			}
			symbol := e.SymbolTable.Symbols[r.Symbol]

			newAddr := symbol.Address + r.Addend
			if r.Offset+8 <= uintptr(len(e.Memory)) {
				e.CallerToFunctionName[r.Offset] = symbol
				binary.LittleEndian.PutUint64(e.Memory[r.Offset:], uint64(newAddr))
			}
		default:
			logger.Print(color.Gray.Sprintf(
				"  Unhandled relocation type %d.\n",
				r.Type,
			))
		}
	}
	logger.Printf(
		"  Applied %s relative & %s external relocations.\n",
		color.Yellow.Sprintf("%d", relativeCount),
		color.Yellow.Sprintf("%d", externalCount),
	)
}
