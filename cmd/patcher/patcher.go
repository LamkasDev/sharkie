package patcher

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/sys_struct"
	"github.com/bpfsnoop/gapstone"
	"github.com/gookit/color"
)

var GlobalPatcher = NewPatcher()

// Patcher keeps track of patching state.
type Patcher struct {
	FastDisassembler           gapstone.Engine
	DetailedDisassembler       gapstone.Engine
	NeededTcbAccessTrampolines []gapstone.Instruction

	ForceGenerate    bool
	PatchesDirectory string
}

// NewPatcher creates a new instance of Patcher.
func NewPatcher() *Patcher {
	var err error
	p := &Patcher{
		PatchesDirectory: path.Join("data", "patches"),
	}
	p.FastDisassembler, err = gapstone.New(gapstone.CS_ARCH_X86, gapstone.CS_MODE_64)
	if err != nil {
		panic(err)
	}
	p.DetailedDisassembler, err = gapstone.New(gapstone.CS_ARCH_X86, gapstone.CS_MODE_64)
	if err != nil {
		panic(err)
	}
	if err = p.DetailedDisassembler.SetOption(gapstone.CS_OPT_DETAIL, gapstone.CS_OPT_ON); err != nil {
		panic(err)
	}

	return p
}

// Patch patches the ELF file.
func (p *Patcher) Patch(e *elf.Elf) error {
	sys_struct.TlsOnce.Do(sys_struct.AllocTlsSlot)
	if sys_struct.TlsSlot >= 64 {
		return errors.New("tls slot is too high, cannot patch tcb access")
	}

	p.NeededTcbAccessTrampolines = []gapstone.Instruction{}
	patchPath := filepath.Join(p.PatchesDirectory, fmt.Sprintf("%s.patch", e.Name))
	if !p.ForceGenerate {
		if _, err := os.Stat(patchPath); err != nil {
			color.Gray.Printf("Didn't patch any instructions...\n")
			return nil
		}
		return p.PatchFast(e, patchPath)
	}

	return p.PatchSlow(e, patchPath)
}

// PatchFast loads instruction offsets from a file and patches them.
func (p *Patcher) PatchFast(e *elf.Elf, patchPath string) error {
	fmt.Printf(
		"Loading patches for %s from %s...\n",
		color.Blue.Sprint(e.Name),
		color.Blue.Sprint(patchPath),
	)
	file, err := os.Open(patchPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	patchCount := 0
	for scanner.Scan() {
		offsetStr := scanner.Text()
		if offsetStr == "" {
			continue
		}

		offset, err := strconv.ParseUint(offsetStr, 10, 64)
		if err != nil {
			color.Warn.Printf("Invalid offset in patch file %s.\n", offsetStr)
			continue
		}

		patched, err := p.ProcessSingleInstruction(e, offset)
		if err != nil {
			return err
		}
		if patched {
			patchCount++
		}
	}

	// Process trampoline candidates.
	for _, inst := range p.NeededTcbAccessTrampolines {
		p.CreateTcbAccessTrampoline(e, inst)
	}

	fmt.Printf(
		"Patched %s instructions.\n",
		color.Green.Sprintf("%d", patchCount),
	)
	return nil
}

// PatchSlow scans the entire binary, applies patches and saves the offsets to a file.
func (p *Patcher) PatchSlow(e *elf.Elf, patchPath string) error {
	fmt.Printf(
		"Scanning %s for patches...\n",
		color.Blue.Sprint(e.Name),
	)

	patchOffsets := []uint64{}
	for _, s := range e.LoadSections {
		if (s.PFlags & elf.PF_X) == 0 {
			continue
		}

		sectionStart := s.PVaddr
		sectionEnd := s.PVaddr + s.PFilesz
		if sectionEnd > uint64(len(e.Memory)) {
			sectionEnd = uint64(len(e.Memory))
		}
		sectionOffset := uint64(0)
		sectionSize := sectionEnd - sectionStart

		for sectionOffset < sectionSize {
			// We try only 512 at a time, so if we error out we can advance over the bad bytes (probably headers).
			offset := sectionStart + sectionOffset
			instructionData := e.Memory[offset:]
			instructions, err := p.FastDisassembler.Disasm(instructionData, offset, 512)
			if err != nil || len(instructions) == 0 {
				sectionOffset++
				continue
			}
			for _, instruction := range instructions {
				sectionOffset += uint64(len(instruction.Bytes))
				if instruction.Mnemonic != "mov" {
					continue
				}
				patched, err := p.ProcessSingleInstruction(e, uint64(instruction.Address))
				if err != nil {
					return err
				}
				if patched {
					patchOffsets = append(patchOffsets, uint64(instruction.Address))
				}
			}
		}
	}

	// Process trampoline candidates.
	for _, instruction := range p.NeededTcbAccessTrampolines {
		p.CreateTcbAccessTrampoline(e, instruction)
	}

	if len(patchOffsets) == 0 {
		color.Gray.Printf("Didn't patch any instructions...\n")
		return nil
	}
	fmt.Printf(
		"Patched %s instructions.\n",
		color.Green.Sprintf("%d", len(patchOffsets)),
	)

	// Save patches to a file.
	if err := os.MkdirAll(p.PatchesDirectory, 0755); err != nil {
		return err
	}
	file, err := os.Create(patchPath)
	if err != nil {
		return err
	}
	defer file.Close()
	for _, offset := range patchOffsets {
		file.WriteString(fmt.Sprintf("%d\n", offset))
	}

	fmt.Printf(
		"Saved %s patches to %s.\n",
		color.Green.Sprintf("%d", len(patchOffsets)),
		color.Blue.Sprint(patchPath),
	)
	return nil
}

// ProcessSingleInstruction disassembles and attempts to patch a specific instruction.
func (p *Patcher) ProcessSingleInstruction(e *elf.Elf, offset uint64) (bool, error) {
	// Disassemble with details.
	instructionData := e.Memory[offset:]
	detailedInstructions, err := p.DetailedDisassembler.Disasm(instructionData, offset, 1)
	if err != nil || len(detailedInstructions) == 0 {
		return false, err
	}

	// Try applying patches.
	instruction := detailedInstructions[0]
	switch p.FilterTcbAccess(instruction) {
	case TcbAccessDirect:
		instructionData = e.Memory[int(instruction.Address) : int(instruction.Address)+len(instruction.Bytes)]
		return p.PatchTcbAccess(instruction, instructionData), nil
	case TcbAccessTrampoline:
		p.NeededTcbAccessTrampolines = append(p.NeededTcbAccessTrampolines, instruction)
		return true, nil
	}

	return false, nil
}
