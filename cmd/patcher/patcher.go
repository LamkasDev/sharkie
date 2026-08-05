package patcher

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/LamkasDev/sharkie/cmd/config"
	"github.com/LamkasDev/sharkie/cmd/elf"
	"github.com/LamkasDev/sharkie/cmd/logger"
	"github.com/bpfsnoop/gapstone"
	"github.com/gookit/color"
)

var GlobalPatcher = NewPatcher()

// Patcher keeps track of patching state.
type Patcher struct {
	FastDisassembler           gapstone.Engine
	DetailedDisassembler       gapstone.Engine
	NeededTcbAccessTrampolines []gapstone.Instruction

	ForceGenerate bool
}

// NewPatcher creates a new instance of Patcher.
func NewPatcher() *Patcher {
	var err error
	p := &Patcher{}
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
	p.NeededTcbAccessTrampolines = []gapstone.Instruction{}

	var patchDir string
	if strings.HasPrefix(e.Path, config.GetLibDir()) {
		cachePath, _ := config.AppScope.CacheDir()
		patchDir = filepath.Join(cachePath, "patches")
	} else {
		patchDir = filepath.Join(config.GetGameCacheDir(), "patches")
	}
	patchPath := filepath.Join(patchDir, fmt.Sprintf("%s.patch", e.Name))
	failedPatchPath := filepath.Join(patchDir, fmt.Sprintf("%s.failed_patch", e.Name))
	if !p.ForceGenerate {
		if _, err := os.Stat(patchPath); err == nil {
			return p.PatchFast(e, patchPath, failedPatchPath)
		}
	}

	return p.PatchSlow(e, patchPath, failedPatchPath)
}

// PatchFast loads instruction offsets from a file and patches them.
func (p *Patcher) PatchFast(e *elf.Elf, patchPath string, failedPatchPath string) error {
	logger.Printf(
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
			logger.Print(color.Warn.Sprintf("Invalid offset in patch file %s.\n", offsetStr))
			continue
		}

		action, err := p.ProcessSingleInstruction(e, offset)
		if err != nil {
			return err
		}
		if action == TcbAccessDirect || action == TcbAccessTrampoline {
			patchCount++
		}
	}

	// Load failed patches.
	if _, err = os.Stat(failedPatchPath); err == nil {
		logger.Printf(
			"Loading failed patches for %s from %s...\n",
			color.Blue.Sprint(e.Name),
			color.Blue.Sprint(failedPatchPath),
		)
		failedFile, err := os.Open(failedPatchPath)
		if err == nil {
			defer failedFile.Close()
			failedScanner := bufio.NewScanner(failedFile)
			for failedScanner.Scan() {
				offsetStr := failedScanner.Text()
				if offsetStr == "" {
					continue
				}
				offset, err := strconv.ParseUint(offsetStr, 10, 64)
				if err == nil {
					GlobalPatcherRuntime.FailedPatchAddresses[uint64(e.BaseAddress)+offset] = true
				}
			}
		}
	}

	// Process trampoline candidates.
	for _, inst := range p.NeededTcbAccessTrampolines {
		p.CreateTcbAccessTrampoline(e, inst)
	}

	if patchCount == 0 {
		logger.Print(color.Gray.Sprintf("Didn't patch any instructions...\n"))
	} else {
		logger.Printf(
			"Patched %s instructions.\n",
			color.Green.Sprintf("%d", patchCount),
		)
	}

	return nil
}

// PatchSlow scans the entire binary, applies patches and saves the offsets to a file.
func (p *Patcher) PatchSlow(e *elf.Elf, patchPath string, failedPatchPath string) error {
	logger.Printf(
		"Scanning %s for patches...\n",
		color.Blue.Sprint(e.Name),
	)

	var patchOffsets []uint64
	var failedPatchOffsets []uint64
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
				action, err := p.ProcessSingleInstruction(e, uint64(instruction.Address))
				if err != nil {
					return err
				}
				if action == TcbAccessDirect || action == TcbAccessTrampoline {
					patchOffsets = append(patchOffsets, uint64(instruction.Address))
				} else if action == TcbAccessEmulate {
					failedPatchOffsets = append(failedPatchOffsets, uint64(instruction.Address))
				}
			}
		}
	}

	// Process trampoline candidates.
	for _, instruction := range p.NeededTcbAccessTrampolines {
		p.CreateTcbAccessTrampoline(e, instruction)
	}

	if len(patchOffsets) == 0 {
		logger.Print(color.Gray.Sprintf("Didn't patch any instructions...\n"))
	} else {
		logger.Printf(
			"Patched %s instructions.\n",
			color.Green.Sprintf("%d", len(patchOffsets)),
		)
	}

	// Save patches to a file.
	if err := os.MkdirAll(filepath.Dir(patchPath), 0755); err != nil {
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

	logger.Printf(
		"Saved %s patches to %s.\n",
		color.Green.Sprintf("%d", len(patchOffsets)),
		color.Blue.Sprint(patchPath),
	)

	// Save failed patches.
	if len(failedPatchOffsets) > 0 {
		failedFile, err := os.Create(failedPatchPath)
		if err == nil {
			defer failedFile.Close()
			for _, offset := range failedPatchOffsets {
				failedFile.WriteString(fmt.Sprintf("%d\n", offset))
			}
			logger.Printf(
				"Saved %s failed patches to %s.\n",
				color.Green.Sprintf("%d", len(failedPatchOffsets)),
				color.Blue.Sprint(failedPatchPath),
			)
		}
	}

	return nil
}

// ProcessSingleInstruction disassembles and attempts to patch a specific instruction.
func (p *Patcher) ProcessSingleInstruction(e *elf.Elf, offset uint64) (int, error) {
	// Disassemble with details.
	instructionData := e.Memory[offset:]
	detailedInstructions, err := p.DetailedDisassembler.Disasm(instructionData, offset, 1)
	if err != nil || len(detailedInstructions) == 0 {
		return TcbAccessNoPatch, err
	}

	// Try applying patches.
	instruction := detailedInstructions[0]
	switch p.FilterTcbAccess(instruction) {
	case TcbAccessDirect:
		instructionData = e.Memory[int(instruction.Address) : int(instruction.Address)+len(instruction.Bytes)]
		p.PatchTcbAccess(instruction, instructionData)
		return TcbAccessDirect, nil
	case TcbAccessTrampoline:
		p.NeededTcbAccessTrampolines = append(p.NeededTcbAccessTrampolines, instruction)
		return TcbAccessTrampoline, nil
	case TcbAccessEmulate:
		GlobalPatcherRuntime.FailedPatchAddresses[uint64(e.BaseAddress)+offset] = true
		return TcbAccessEmulate, nil
	}

	return TcbAccessNoPatch, nil
}
