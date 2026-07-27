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

// ProcessRelocationTable applies all relocations in a given section table to the module.
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
		resolved, isRelative := processRelocation(e, &r, true)
		if !resolved {
			e.UnresolvedRelocations = append(e.UnresolvedRelocations, new(r))
			continue
		}
		if isRelative {
			relativeCount++
		} else {
			externalCount++
		}
	}
	logger.Printf(
		"  Applied %s relative & %s external relocations.\n",
		color.Yellow.Sprintf("%d", relativeCount),
		color.Yellow.Sprintf("%d", externalCount),
	)
}

// ResolvePendingRelocations retries any relocations that were deferred during load.
func ResolvePendingRelocations(modules []*elf.Elf) {
	for _, e := range modules {
		if e == nil {
			continue
		}
		var stillUnresolved []*elf.ElfRelocation
		resolvedCount := 0
		for _, r := range e.UnresolvedRelocations {
			resolved, _ := processRelocation(e, r, false)
			if !resolved {
				stillUnresolved = append(stillUnresolved, r)
				continue
			}
			resolvedCount++
		}
		e.UnresolvedRelocations = stillUnresolved
		if resolvedCount > 0 {
			logger.Printf("Resolved %d previously pending relocations for %s.\n", resolvedCount, color.Blue.Sprint(e.Name))
		}
	}
}

// processRelocation attempts to process a single relocation.
// Returns (resolved, isRelative) where resolved is true if successfully handled.
func processRelocation(e *elf.Elf, r *elf.ElfRelocation, firstPass bool) (bool, bool) {
	switch r.Type {
	case elf.R_AMD64_RELATIVE:
		newAddr := e.BaseAddress + r.Addend
		if r.Offset+8 <= uintptr(len(e.Memory)) {
			binary.LittleEndian.PutUint64(e.Memory[r.Offset:], uint64(newAddr))
		}
		return true, true
	case elf.R_AMD64_64:
		if r.Symbol == 0 {
			newAddr := e.BaseAddress + r.Addend
			if r.Offset+8 <= uintptr(len(e.Memory)) {
				binary.LittleEndian.PutUint64(e.Memory[r.Offset:], uint64(newAddr))
			}
			return true, true
		}
		fallthrough
	case elf.R_AMD64_GLOB_DAT, elf.R_AMD64_JUMP_SLOT:
		if int(r.Symbol) >= len(e.SymbolTable.Symbols) {
			return true, false
		}
		symbol := e.SymbolTable.Symbols[r.Symbol]
		if addr, ok := elf.GetSymbolAddress(symbol); ok {
			newAddr := addr + r.Addend
			if r.Offset+8 <= uintptr(len(e.Memory)) {
				e.CallerToFunctionName[r.Offset] = symbol
				binary.LittleEndian.PutUint64(e.Memory[r.Offset:], uint64(newAddr))
			}
			return true, false
		}
		if firstPass {
			logger.Print(color.Gray.Sprintf("  Skipped fake address for %s:%s.\n", symbol.LibraryName, symbol.ReadableName))
		}
		return false, false
	case elf.R_AMD64_DTPMOD64:
		if int(r.Symbol) >= len(e.SymbolTable.Symbols) {
			return true, false
		}
		symbol := e.SymbolTable.Symbols[r.Symbol]
		moduleIndex := e.ModuleIndex
		if symbol.Type != elf.STT_SECTION && symbol.OriginalName != "" {
			module := elf.GetDefiningModule(symbol)
			if module == nil {
				if firstPass {
					logger.Print(color.Gray.Sprintf(
						"  Failed finding defining module for %s:%s.\n",
						symbol.LibraryName,
						symbol.ReadableName,
					))
				}
				return false, false
			}
			moduleIndex = module.ModuleIndex
		}
		if r.Offset+8 <= uintptr(len(e.Memory)) {
			binary.LittleEndian.PutUint64(e.Memory[r.Offset:], moduleIndex)
		}
		return true, false
	case elf.R_AMD64_DTPOFF64:
		if int(r.Symbol) >= len(e.SymbolTable.Symbols) {
			return true, false
		}
		symbol := e.SymbolTable.Symbols[r.Symbol]
		if symbol.Type != elf.STT_SECTION && symbol.OriginalName != "" {
			module := elf.GetDefiningModule(symbol)
			if module == nil {
				if firstPass {
					logger.Print(color.Gray.Sprintf(
						"  Failed finding defining module for %s:%s.\n",
						symbol.LibraryName,
						symbol.ReadableName,
					))
				}
				return false, false
			}
			if definedSymbol := module.SymbolTable.SymbolsMap[symbol.HashIndex]; definedSymbol != nil {
				symbol = definedSymbol
			}
		}

		newAddr := symbol.Address + r.Addend
		if r.Offset+8 <= uintptr(len(e.Memory)) {
			e.CallerToFunctionName[r.Offset] = symbol
			binary.LittleEndian.PutUint64(e.Memory[r.Offset:], uint64(newAddr))
		}
		return true, false
	default:
		if firstPass {
			logger.Print(color.Gray.Sprintf(
				"  Unhandled relocation type %d.\n",
				r.Type,
			))
		}
		return true, false
	}
}
